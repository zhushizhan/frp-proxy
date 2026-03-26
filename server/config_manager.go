package server

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	toml "github.com/pelletier/go-toml/v2"
	yaml "sigs.k8s.io/yaml"

	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/config/types"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/config/v1/validation"
	"github.com/fatedier/frp/pkg/policy/security"
	"github.com/fatedier/frp/pkg/util/log"
	sermodel "github.com/fatedier/frp/server/http/model"
)

type FileConfigManagerOptions struct {
	Service        *Service
	ConfigFilePath string
	ExecutablePath string
	Args           []string
	WorkDir        string
	UnsafeFeatures *security.UnsafeFeatures
}

type FileConfigManager struct {
	service        *Service
	configFilePath string
	executablePath string
	args           []string
	workDir        string
	unsafeFeatures *security.UnsafeFeatures

	restartScheduled atomic.Bool
}

func NewFileConfigManager(opts FileConfigManagerOptions) *FileConfigManager {
	return &FileConfigManager{
		service:        opts.Service,
		configFilePath: opts.ConfigFilePath,
		executablePath: opts.ExecutablePath,
		args:           opts.Args,
		workDir:        opts.WorkDir,
		unsafeFeatures: opts.UnsafeFeatures,
	}
}

func (m *FileConfigManager) GetSettings() (sermodel.ServerSettings, error) {
	cfg, isLegacy, err := m.loadConfig()
	if err != nil {
		return sermodel.ServerSettings{}, err
	}
	if isLegacy {
		return sermodel.ServerSettings{}, fmt.Errorf("legacy ini format is not supported for UI settings")
	}
	return m.toSettings(cfg), nil
}

func (m *FileConfigManager) UpdateSettings(in sermodel.ServerSettings) error {
	cfg, isLegacy, err := m.loadConfig()
	if err != nil {
		return err
	}
	if isLegacy {
		return fmt.Errorf("legacy ini format is not supported for UI settings")
	}

	allowPorts := []types.PortsRange{}
	if strings.TrimSpace(in.AllowPorts) != "" {
		allowPorts, err = types.NewPortsRangeSliceFromString(in.AllowPorts)
		if err != nil {
			return fmt.Errorf("invalid allowPorts: %w", err)
		}
	}

	cfg.BindAddr = strings.TrimSpace(in.BindAddr)
	cfg.BindPort = in.BindPort
	cfg.ProxyBindAddr = strings.TrimSpace(in.ProxyBindAddr)
	cfg.KCPBindPort = in.KCPBindPort
	cfg.QUICBindPort = in.QUICBindPort
	cfg.VhostHTTPPort = in.VhostHTTPPort
	cfg.VhostHTTPSPort = in.VhostHTTPSPort
	cfg.VhostHTTPTimeout = in.VhostHTTPTimeout
	cfg.TCPMuxHTTPConnectPort = in.TCPMuxHTTPConnectPort
	cfg.TCPMuxPassthrough = in.TCPMuxPassthrough
	cfg.SubDomainHost = strings.TrimSpace(in.SubdomainHost)
	cfg.Auth.Method = v1.AuthMethod(strings.TrimSpace(in.AuthMethod))
	if cfg.Auth.Method == "" {
		cfg.Auth.Method = v1.AuthMethodToken
	}
	cfg.Auth.AdditionalScopes = stringSliceToAuthScopes(in.AuthAdditionalScopes)
	switch cfg.Auth.Method {
	case v1.AuthMethodToken:
		cfg.Auth.OIDC = v1.AuthOIDCServerConfig{}
		switch strings.TrimSpace(in.AuthTokenSourceType) {
		case "", "inline":
			cfg.Auth.Token = strings.TrimSpace(in.AuthToken)
			cfg.Auth.TokenSource = nil
		case "file":
			path := strings.TrimSpace(in.AuthTokenSourceFile)
			if path == "" {
				cfg.Auth.Token = strings.TrimSpace(in.AuthToken)
				cfg.Auth.TokenSource = nil
			} else {
				cfg.Auth.Token = ""
				cfg.Auth.TokenSource = &v1.ValueSource{
					Type: "file",
					File: &v1.FileSource{
						Path: path,
					},
				}
			}
		case "exec":
			if cfg.Auth.TokenSource == nil || cfg.Auth.TokenSource.Type != "exec" {
				return fmt.Errorf("exec auth token source is not editable in UI; please use the text config")
			}
			cfg.Auth.Token = ""
		default:
			return fmt.Errorf("unsupported authTokenSourceType: %s", in.AuthTokenSourceType)
		}
	case v1.AuthMethodOIDC:
		cfg.Auth.Token = ""
		cfg.Auth.TokenSource = nil
		cfg.Auth.OIDC = v1.AuthOIDCServerConfig{
			Issuer:          strings.TrimSpace(in.OIDCIssuer),
			Audience:        strings.TrimSpace(in.OIDCAudience),
			SkipExpiryCheck: in.OIDCSkipExpiryCheck,
			SkipIssuerCheck: in.OIDCSkipIssuerCheck,
		}
	default:
		return fmt.Errorf("unsupported auth method: %s", cfg.Auth.Method)
	}
	cfg.Transport.TLS.Force = in.TLSForce
	cfg.Transport.TLS.CertFile = strings.TrimSpace(in.TransportTLSCertFile)
	cfg.Transport.TLS.KeyFile = strings.TrimSpace(in.TransportTLSKeyFile)
	cfg.Transport.TLS.TrustedCaFile = strings.TrimSpace(in.TransportTLSTrustedCA)
	cfg.Transport.TCPMux = boolPtr(in.TCPMux)
	cfg.Transport.TCPMuxKeepaliveInterval = in.TCPMuxKeepalive
	cfg.Transport.MaxPoolCount = in.MaxPoolCount
	cfg.Transport.HeartbeatTimeout = in.HeartbeatTimeout
	cfg.Transport.TCPKeepAlive = in.TCPKeepAlive
	cfg.Transport.QUIC.KeepalivePeriod = in.QUICKeepalivePeriod
	cfg.Transport.QUIC.MaxIdleTimeout = in.QUICMaxIdleTimeout
	cfg.Transport.QUIC.MaxIncomingStreams = in.QUICMaxIncomingStreams
	cfg.MaxPortsPerClient = in.MaxPortsPerClient
	cfg.UserConnTimeout = in.UserConnTimeout
	cfg.UDPPacketSize = in.UDPPacketSize
	cfg.NatHoleAnalysisDataReserveHours = in.NatHoleRetentionHours
	cfg.DetailedErrorsToClient = &in.DetailedErrorsToClient
	cfg.AllowPorts = allowPorts
	cfg.EnablePrometheus = in.EnablePrometheus
	cfg.WebServer.Addr = strings.TrimSpace(in.DashboardAddr)
	cfg.WebServer.Port = in.DashboardPort
	cfg.WebServer.User = strings.TrimSpace(in.DashboardUser)
	cfg.WebServer.Password = in.DashboardPassword
	cfg.WebServer.AssetsDir = strings.TrimSpace(in.DashboardAssetsDir)
	cfg.WebServer.PprofEnable = in.DashboardPprofEnable
	applyDashboardTLSSettings(&cfg.WebServer, in)
	cfg.Log.To = strings.TrimSpace(in.LogTo)
	cfg.Log.Level = strings.TrimSpace(in.LogLevel)
	cfg.Log.MaxDays = in.LogMaxDays
	cfg.Log.DisablePrintColor = in.LogDisablePrintColor
	cfg.SSHTunnelGateway.BindPort = in.SSHTunnelGatewayPort
	cfg.SSHTunnelGateway.PrivateKeyFile = strings.TrimSpace(in.SSHPrivateKeyFile)
	cfg.SSHTunnelGateway.AutoGenPrivateKeyPath = strings.TrimSpace(in.SSHAutoGenKeyPath)
	cfg.SSHTunnelGateway.AuthorizedKeysFile = strings.TrimSpace(in.SSHAuthorizedKeysFile)
	cfg.HTTPPlugins = toHTTPPluginOptions(in.HTTPPlugins)
	cfg.Custom404Page = strings.TrimSpace(in.Custom404Page)

	if err := cfg.Complete(); err != nil {
		return err
	}

	validator := validation.NewConfigValidator(m.unsafeFeatures)
	if _, err := validator.ValidateServerConfig(cfg); err != nil {
		return err
	}

	if err := m.writeConfig(cfg); err != nil {
		return err
	}

	m.scheduleRestart()
	return nil
}

func (m *FileConfigManager) toSettings(cfg *v1.ServerConfig) sermodel.ServerSettings {
	settings := sermodel.ServerSettings{
		ConfigPath:             m.configFilePath,
		AutoRestart:            true,
		BindAddr:               cfg.BindAddr,
		BindPort:               cfg.BindPort,
		ProxyBindAddr:          cfg.ProxyBindAddr,
		KCPBindPort:            cfg.KCPBindPort,
		QUICBindPort:           cfg.QUICBindPort,
		VhostHTTPPort:          cfg.VhostHTTPPort,
		VhostHTTPSPort:         cfg.VhostHTTPSPort,
		VhostHTTPTimeout:       cfg.VhostHTTPTimeout,
		TCPMuxHTTPConnectPort:  cfg.TCPMuxHTTPConnectPort,
		TCPMuxPassthrough:      cfg.TCPMuxPassthrough,
		SubdomainHost:          cfg.SubDomainHost,
		AuthMethod:             string(cfg.Auth.Method),
		AuthAdditionalScopes:   authScopesToStringSlice(cfg.Auth.AdditionalScopes),
		AuthToken:              cfg.Auth.Token,
		AuthTokenSourceType:    authTokenSourceType(cfg.Auth.TokenSource),
		OIDCIssuer:             cfg.Auth.OIDC.Issuer,
		OIDCAudience:           cfg.Auth.OIDC.Audience,
		OIDCSkipExpiryCheck:    cfg.Auth.OIDC.SkipExpiryCheck,
		OIDCSkipIssuerCheck:    cfg.Auth.OIDC.SkipIssuerCheck,
		TLSForce:               cfg.Transport.TLS.Force,
		TransportTLSCertFile:   cfg.Transport.TLS.CertFile,
		TransportTLSKeyFile:    cfg.Transport.TLS.KeyFile,
		TransportTLSTrustedCA:  cfg.Transport.TLS.TrustedCaFile,
		TCPMux:                 boolPtrValue(cfg.Transport.TCPMux),
		TCPMuxKeepalive:        cfg.Transport.TCPMuxKeepaliveInterval,
		MaxPoolCount:           cfg.Transport.MaxPoolCount,
		HeartbeatTimeout:       cfg.Transport.HeartbeatTimeout,
		TCPKeepAlive:           cfg.Transport.TCPKeepAlive,
		QUICKeepalivePeriod:    cfg.Transport.QUIC.KeepalivePeriod,
		QUICMaxIdleTimeout:     cfg.Transport.QUIC.MaxIdleTimeout,
		QUICMaxIncomingStreams: cfg.Transport.QUIC.MaxIncomingStreams,
		MaxPortsPerClient:      cfg.MaxPortsPerClient,
		UserConnTimeout:        cfg.UserConnTimeout,
		UDPPacketSize:          cfg.UDPPacketSize,
		NatHoleRetentionHours:  cfg.NatHoleAnalysisDataReserveHours,
		DetailedErrorsToClient: cfg.DetailedErrorsToClient != nil && *cfg.DetailedErrorsToClient,
		AllowPorts:             types.PortsRangeSlice(cfg.AllowPorts).String(),
		EnablePrometheus:       cfg.EnablePrometheus,
		DashboardAddr:          cfg.WebServer.Addr,
		DashboardPort:          cfg.WebServer.Port,
		DashboardUser:          cfg.WebServer.User,
		DashboardPassword:      cfg.WebServer.Password,
		DashboardAssetsDir:     cfg.WebServer.AssetsDir,
		DashboardPprofEnable:   cfg.WebServer.PprofEnable,
		LogTo:                  cfg.Log.To,
		LogLevel:               cfg.Log.Level,
		LogMaxDays:             cfg.Log.MaxDays,
		LogDisablePrintColor:   cfg.Log.DisablePrintColor,
		SSHTunnelGatewayPort:   cfg.SSHTunnelGateway.BindPort,
		SSHPrivateKeyFile:      cfg.SSHTunnelGateway.PrivateKeyFile,
		SSHAutoGenKeyPath:      cfg.SSHTunnelGateway.AutoGenPrivateKeyPath,
		SSHAuthorizedKeysFile:  cfg.SSHTunnelGateway.AuthorizedKeysFile,
		HTTPPlugins:            fromHTTPPluginOptions(cfg.HTTPPlugins),
		Custom404Page:          cfg.Custom404Page,
	}
	if cfg.Auth.TokenSource != nil && cfg.Auth.TokenSource.Type == "file" && cfg.Auth.TokenSource.File != nil {
		settings.AuthTokenSourceFile = cfg.Auth.TokenSource.File.Path
	}
	if cfg.WebServer.TLS != nil {
		settings.DashboardTLSCertFile = cfg.WebServer.TLS.CertFile
		settings.DashboardTLSKeyFile = cfg.WebServer.TLS.KeyFile
		settings.DashboardTLSTrustedCA = cfg.WebServer.TLS.TrustedCaFile
	}
	return settings
}

func (m *FileConfigManager) loadConfig() (*v1.ServerConfig, bool, error) {
	if m.configFilePath == "" {
		return nil, false, fmt.Errorf("frps is not running with a config file")
	}
	return config.LoadServerConfig(m.configFilePath, true)
}

func (m *FileConfigManager) writeConfig(cfg *v1.ServerConfig) error {
	if m.configFilePath == "" {
		return fmt.Errorf("frps is not running with a config file")
	}

	var (
		data []byte
		err  error
	)

	switch strings.ToLower(filepath.Ext(m.configFilePath)) {
	case ".toml":
		data, err = toml.Marshal(cfg)
	case ".yaml", ".yml":
		data, err = yaml.Marshal(cfg)
	case ".json":
		data, err = json.MarshalIndent(cfg, "", "  ")
	default:
		return fmt.Errorf("unsupported config format for UI settings")
	}
	if err != nil {
		return err
	}

	return os.WriteFile(m.configFilePath, data, 0o600)
}

func (m *FileConfigManager) scheduleRestart() {
	if m.service == nil || m.executablePath == "" {
		return
	}
	if !m.restartScheduled.CompareAndSwap(false, true) {
		return
	}

	go func() {
		cmd := exec.Command(m.executablePath, m.args...)
		cmd.Dir = m.workDir
		cmd.Env = append(os.Environ(), "FRP_STARTUP_DELAY_MS=1000")
		if err := cmd.Start(); err != nil {
			log.Errorf("failed to schedule frps restart: %v", err)
			m.restartScheduled.Store(false)
			return
		}

		time.Sleep(200 * time.Millisecond)
		_ = m.service.Close()
	}()
}

func applyDashboardTLSSettings(webServer *v1.WebServerConfig, in sermodel.ServerSettings) {
	certFile := strings.TrimSpace(in.DashboardTLSCertFile)
	keyFile := strings.TrimSpace(in.DashboardTLSKeyFile)
	trustedCAFile := strings.TrimSpace(in.DashboardTLSTrustedCA)
	if certFile == "" && keyFile == "" && trustedCAFile == "" {
		webServer.TLS = nil
		return
	}
	if webServer.TLS == nil {
		webServer.TLS = &v1.TLSConfig{}
	}
	webServer.TLS.CertFile = certFile
	webServer.TLS.KeyFile = keyFile
	webServer.TLS.TrustedCaFile = trustedCAFile
}

func authScopesToStringSlice(scopes []v1.AuthScope) []string {
	if len(scopes) == 0 {
		return nil
	}
	out := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		out = append(out, string(scope))
	}
	return out
}

func stringSliceToAuthScopes(scopes []string) []v1.AuthScope {
	if len(scopes) == 0 {
		return nil
	}
	out := make([]v1.AuthScope, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		out = append(out, v1.AuthScope(scope))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func authTokenSourceType(source *v1.ValueSource) string {
	if source == nil {
		return ""
	}
	return source.Type
}

func boolPtr(v bool) *bool {
	return &v
}

func boolPtrValue(v *bool) bool {
	return v != nil && *v
}

func toHTTPPluginOptions(in []sermodel.HTTPPluginSettings) []v1.HTTPPluginOptions {
	if len(in) == 0 {
		return nil
	}
	out := make([]v1.HTTPPluginOptions, 0, len(in))
	for _, plugin := range in {
		out = append(out, v1.HTTPPluginOptions{
			Name:      strings.TrimSpace(plugin.Name),
			Addr:      strings.TrimSpace(plugin.Addr),
			Path:      strings.TrimSpace(plugin.Path),
			Ops:       trimStringSlice(plugin.Ops),
			TLSVerify: plugin.TLSVerify,
		})
	}
	return out
}

func fromHTTPPluginOptions(in []v1.HTTPPluginOptions) []sermodel.HTTPPluginSettings {
	if len(in) == 0 {
		return nil
	}
	out := make([]sermodel.HTTPPluginSettings, 0, len(in))
	for _, plugin := range in {
		out = append(out, sermodel.HTTPPluginSettings{
			Name:      plugin.Name,
			Addr:      plugin.Addr,
			Path:      plugin.Path,
			Ops:       append([]string(nil), plugin.Ops...),
			TLSVerify: plugin.TLSVerify,
		})
	}
	return out
}

func trimStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
