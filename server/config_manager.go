package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	toml "github.com/pelletier/go-toml/v2"
	yaml "sigs.k8s.io/yaml"

	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/config/types"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/config/v1/validation"
	"github.com/fatedier/frp/pkg/policy/security"
	"github.com/fatedier/frp/pkg/util/fileutil"
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
	if m.configFilePath == "" {
		return fmt.Errorf("frps is not running with a config file")
	}

	cfg, isLegacy, err := m.loadConfig()
	if err != nil {
		return err
	}
	if isLegacy {
		return fmt.Errorf("legacy ini format is not supported for UI settings")
	}
	doc, format, err := loadServerConfigDocument(m.configFilePath)
	if err != nil {
		return err
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

	applyServerSettingsToDocument(doc, in, allowPorts)

	data, err := marshalServerConfigDocument(format, doc)
	if err != nil {
		return err
	}
	if err := validateServerConfigContent(m.configFilePath, data, m.unsafeFeatures); err != nil {
		return err
	}
	if err := os.WriteFile(m.configFilePath, data, 0o600); err != nil {
		return err
	}

	m.scheduleRestart()
	return nil
}

func (m *FileConfigManager) UploadFile(targetPath string, filename string, content []byte) (string, error) {
	if m.configFilePath == "" {
		return "", fmt.Errorf("frps is not running with a config file")
	}
	return fileutil.WriteUploadTarget(m.configFilePath, targetPath, filename, content)
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

func loadServerConfigDocument(path string) (map[string]any, string, error) {
	if path == "" {
		return nil, "", fmt.Errorf("frps is not running with a config file")
	}
	if config.DetectLegacyINIFormatFromFile(path) {
		return nil, "", fmt.Errorf("legacy ini format is not supported for UI settings")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}

	format := strings.ToLower(filepath.Ext(path))
	doc := map[string]any{}
	switch format {
	case ".toml":
		if err := toml.Unmarshal(content, &doc); err != nil {
			return nil, format, err
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(content, &doc); err != nil {
			return nil, format, err
		}
	case ".json":
		if err := json.Unmarshal(content, &doc); err != nil {
			return nil, format, err
		}
	default:
		return nil, format, fmt.Errorf("unsupported config format for UI settings")
	}
	return doc, format, nil
}

func marshalServerConfigDocument(format string, doc map[string]any) ([]byte, error) {
	switch format {
	case ".toml":
		return toml.Marshal(doc)
	case ".yaml", ".yml":
		return yaml.Marshal(doc)
	case ".json":
		return json.MarshalIndent(doc, "", "  ")
	default:
		return nil, fmt.Errorf("unsupported config format for UI settings")
	}
}

func validateServerConfigContent(path string, content []byte, unsafeFeatures *security.UnsafeFeatures) error {
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	tmpFile, err := os.CreateTemp(dir, "frps-settings-*"+ext)
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmpFile.Write(content); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	cfg, isLegacy, err := config.LoadServerConfig(tmpPath, true)
	if err != nil {
		return err
	}
	if isLegacy {
		return fmt.Errorf("legacy ini format is not supported for UI settings")
	}

	validator := validation.NewConfigValidator(unsafeFeatures)
	_, err = validator.ValidateServerConfig(cfg)
	return err
}

func (m *FileConfigManager) scheduleRestart() {
	if m.executablePath == "" || m.configFilePath == "" {
		return
	}
	if !m.restartScheduled.CompareAndSwap(false, true) {
		return
	}

	go func() {
		scriptPath, err := ensureFRPSServiceScript(m.executablePath)
		if err != nil {
			log.Errorf("failed to prepare frps service script: %v", err)
			m.restartScheduled.Store(false)
			return
		}

		cmd, err := buildFRPSServiceCommand(scriptPath, os.Getpid(), m.executablePath, m.configFilePath, m.workDir)
		if err != nil {
			log.Errorf("failed to build frps service restart command: %v", err)
			m.restartScheduled.Store(false)
			return
		}
		if err := cmd.Start(); err != nil {
			log.Errorf("failed to schedule frps restart: %v", err)
			m.restartScheduled.Store(false)
			return
		}
	}()
}

func applyServerSettingsToDocument(doc map[string]any, in sermodel.ServerSettings, allowPorts []types.PortsRange) {
	bindAddr := strings.TrimSpace(in.BindAddr)
	if bindAddr == "" {
		bindAddr = "0.0.0.0"
	}

	setStringOrDeleteIfDefault(doc, bindAddr, "0.0.0.0", "bindAddr")
	setIntOrDeleteIfDefault(doc, in.BindPort, 7000, "bindPort")
	setStringOrDeleteIfDefault(doc, strings.TrimSpace(in.ProxyBindAddr), bindAddr, "proxyBindAddr")
	setIntOrDeleteIfDefault(doc, in.KCPBindPort, 0, "kcpBindPort")
	setIntOrDeleteIfDefault(doc, in.QUICBindPort, 0, "quicBindPort")
	setIntOrDeleteIfDefault(doc, in.VhostHTTPPort, 0, "vhostHTTPPort")
	setIntOrDeleteIfDefault(doc, in.VhostHTTPSPort, 0, "vhostHTTPSPort")
	setInt64OrDeleteIfDefault(doc, in.VhostHTTPTimeout, 60, "vhostHTTPTimeout")
	setIntOrDeleteIfDefault(doc, in.TCPMuxHTTPConnectPort, 0, "tcpmuxHTTPConnectPort")
	setBoolOrDeleteIfDefault(doc, in.TCPMuxPassthrough, false, "tcpmuxPassthrough")
	setStringOrDelete(doc, strings.TrimSpace(in.SubdomainHost), "subDomainHost")
	setStringOrDelete(doc, strings.TrimSpace(in.Custom404Page), "custom404Page")

	setStringOrDeleteIfDefault(doc, strings.TrimSpace(in.AuthMethod), "token", "auth", "method")
	setStringSliceOrDelete(doc, in.AuthAdditionalScopes, "auth", "additionalScopes")
	switch strings.TrimSpace(in.AuthMethod) {
	case "", "token":
		applyServerTokenSettings(doc, in)
	case "oidc":
		deletePath(doc, "auth", "token")
		deletePath(doc, "auth", "tokenSource")
		applyServerOIDCSettings(doc, in)
	}

	if shouldPersistServerTLSForce(in) {
		setValue(doc, true, "transport", "tls", "force")
	} else {
		deletePath(doc, "transport", "tls", "force")
	}
	setStringOrDelete(doc, strings.TrimSpace(in.TransportTLSCertFile), "transport", "tls", "certFile")
	setStringOrDelete(doc, strings.TrimSpace(in.TransportTLSKeyFile), "transport", "tls", "keyFile")
	setStringOrDelete(doc, strings.TrimSpace(in.TransportTLSTrustedCA), "transport", "tls", "trustedCaFile")
	setBoolOrDeleteIfDefault(doc, in.TCPMux, true, "transport", "tcpMux")
	setInt64OrDeleteIfDefault(doc, in.TCPMuxKeepalive, 30, "transport", "tcpMuxKeepaliveInterval")
	setInt64OrDeleteIfDefault(doc, in.MaxPoolCount, 5, "transport", "maxPoolCount")
	setInt64OrDeleteIfDefault(doc, in.HeartbeatTimeout, defaultServerHeartbeatTimeout(in.TCPMux), "transport", "heartbeatTimeout")
	setInt64OrDeleteIfDefault(doc, in.TCPKeepAlive, 7200, "transport", "tcpKeepalive")
	setIntOrDeleteIfDefault(doc, in.QUICKeepalivePeriod, 10, "transport", "quic", "keepalivePeriod")
	setIntOrDeleteIfDefault(doc, in.QUICMaxIdleTimeout, 30, "transport", "quic", "maxIdleTimeout")
	setIntOrDeleteIfDefault(doc, in.QUICMaxIncomingStreams, 100000, "transport", "quic", "maxIncomingStreams")

	setInt64OrDeleteIfDefault(doc, in.MaxPortsPerClient, 0, "maxPortsPerClient")
	setInt64OrDeleteIfDefault(doc, in.UserConnTimeout, 10, "userConnTimeout")
	setInt64OrDeleteIfDefault(doc, in.UDPPacketSize, 1500, "udpPacketSize")
	setInt64OrDeleteIfDefault(doc, in.NatHoleRetentionHours, 7*24, "natholeAnalysisDataReserveHours")
	setBoolOrDeleteIfDefault(doc, in.DetailedErrorsToClient, true, "detailedErrorsToClient")
	setAllowPortsOrDelete(doc, allowPorts, "allowPorts")
	setBoolOrDeleteIfDefault(doc, in.EnablePrometheus, false, "enablePrometheus")

	setStringOrDeleteIfDefault(doc, strings.TrimSpace(in.DashboardAddr), "127.0.0.1", "webServer", "addr")
	setIntOrDeleteIfDefault(doc, in.DashboardPort, 0, "webServer", "port")
	setStringOrDelete(doc, strings.TrimSpace(in.DashboardUser), "webServer", "user")
	setStringOrDelete(doc, strings.TrimSpace(in.DashboardPassword), "webServer", "password")
	setStringOrDelete(doc, strings.TrimSpace(in.DashboardAssetsDir), "webServer", "assetsDir")
	setBoolOrDeleteIfDefault(doc, in.DashboardPprofEnable, false, "webServer", "pprofEnable")
	setDashboardTLSDocument(doc, in)

	setStringOrDeleteIfDefault(doc, strings.TrimSpace(in.LogTo), "console", "log", "to")
	setStringOrDeleteIfDefault(doc, strings.TrimSpace(in.LogLevel), "info", "log", "level")
	setInt64OrDeleteIfDefault(doc, in.LogMaxDays, 3, "log", "maxDays")
	setBoolOrDeleteIfDefault(doc, in.LogDisablePrintColor, false, "log", "disablePrintColor")

	setIntOrDeleteIfDefault(doc, in.SSHTunnelGatewayPort, 0, "sshTunnelGateway", "bindPort")
	setStringOrDelete(doc, strings.TrimSpace(in.SSHPrivateKeyFile), "sshTunnelGateway", "privateKeyFile")
	setStringOrDeleteIfDefault(doc, strings.TrimSpace(in.SSHAutoGenKeyPath), "./.autogen_ssh_key", "sshTunnelGateway", "autoGenPrivateKeyPath")
	setStringOrDelete(doc, strings.TrimSpace(in.SSHAuthorizedKeysFile), "sshTunnelGateway", "authorizedKeysFile")

	setHTTPPluginsOrDelete(doc, in.HTTPPlugins, "httpPlugins")
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

func applyServerTokenSettings(doc map[string]any, in sermodel.ServerSettings) {
	deletePath(doc, "auth", "oidc")
	switch strings.TrimSpace(in.AuthTokenSourceType) {
	case "file":
		deletePath(doc, "auth", "token")
		deletePath(doc, "auth", "tokenSource", "exec")
		setValue(doc, "file", "auth", "tokenSource", "type")
		setStringOrDelete(doc, strings.TrimSpace(in.AuthTokenSourceFile), "auth", "tokenSource", "file", "path")
	case "exec":
		deletePath(doc, "auth", "token")
		deletePath(doc, "auth", "tokenSource", "file")
	default:
		deletePath(doc, "auth", "tokenSource")
		setStringOrDelete(doc, strings.TrimSpace(in.AuthToken), "auth", "token")
	}
}

func applyServerOIDCSettings(doc map[string]any, in sermodel.ServerSettings) {
	setStringOrDelete(doc, strings.TrimSpace(in.OIDCIssuer), "auth", "oidc", "issuer")
	setStringOrDelete(doc, strings.TrimSpace(in.OIDCAudience), "auth", "oidc", "audience")
	setBoolOrDeleteIfDefault(doc, in.OIDCSkipExpiryCheck, false, "auth", "oidc", "skipExpiryCheck")
	setBoolOrDeleteIfDefault(doc, in.OIDCSkipIssuerCheck, false, "auth", "oidc", "skipIssuerCheck")
}

func shouldPersistServerTLSForce(in sermodel.ServerSettings) bool {
	if !in.TLSForce {
		return false
	}
	return strings.TrimSpace(in.TransportTLSTrustedCA) == ""
}

func setDashboardTLSDocument(doc map[string]any, in sermodel.ServerSettings) {
	certFile := strings.TrimSpace(in.DashboardTLSCertFile)
	keyFile := strings.TrimSpace(in.DashboardTLSKeyFile)
	trustedCAFile := strings.TrimSpace(in.DashboardTLSTrustedCA)
	if certFile == "" && keyFile == "" && trustedCAFile == "" {
		deletePath(doc, "webServer", "tls")
		return
	}
	setStringOrDelete(doc, certFile, "webServer", "tls", "certFile")
	setStringOrDelete(doc, keyFile, "webServer", "tls", "keyFile")
	setStringOrDelete(doc, trustedCAFile, "webServer", "tls", "trustedCaFile")
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

func setAllowPortsOrDelete(doc map[string]any, ranges []types.PortsRange, path ...string) {
	if len(ranges) == 0 {
		deletePath(doc, path...)
		return
	}

	values := make([]map[string]any, 0, len(ranges))
	for _, portRange := range ranges {
		item := map[string]any{}
		if portRange.Single > 0 {
			item["single"] = portRange.Single
		} else {
			item["start"] = portRange.Start
			item["end"] = portRange.End
		}
		values = append(values, item)
	}
	setValue(doc, values, path...)
}

func setHTTPPluginsOrDelete(doc map[string]any, plugins []sermodel.HTTPPluginSettings, path ...string) {
	if len(plugins) == 0 {
		deletePath(doc, path...)
		return
	}

	values := make([]map[string]any, 0, len(plugins))
	for _, plugin := range plugins {
		item := map[string]any{
			"name": strings.TrimSpace(plugin.Name),
			"addr": strings.TrimSpace(plugin.Addr),
			"path": strings.TrimSpace(plugin.Path),
		}
		if ops := trimStringSlice(plugin.Ops); len(ops) > 0 {
			item["ops"] = ops
		}
		if plugin.TLSVerify {
			item["tlsVerify"] = true
		}
		values = append(values, item)
	}
	setValue(doc, values, path...)
}

func setValue(doc map[string]any, value any, path ...string) {
	if len(path) == 0 {
		return
	}
	current := doc
	for _, key := range path[:len(path)-1] {
		next, ok := current[key].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[key] = next
		}
		current = next
	}
	current[path[len(path)-1]] = value
}

func setStringOrDelete(doc map[string]any, value string, path ...string) {
	if value == "" {
		deletePath(doc, path...)
		return
	}
	setValue(doc, value, path...)
}

func setStringSliceOrDelete(doc map[string]any, values []string, path ...string) {
	if len(values) == 0 {
		deletePath(doc, path...)
		return
	}
	setValue(doc, values, path...)
}

func setStringOrDeleteIfDefault(doc map[string]any, value string, defaultValue string, path ...string) {
	if value == "" || value == defaultValue {
		deletePath(doc, path...)
		return
	}
	setValue(doc, value, path...)
}

func setIntOrDeleteIfDefault(doc map[string]any, value int, defaultValue int, path ...string) {
	if value == defaultValue {
		deletePath(doc, path...)
		return
	}
	setValue(doc, value, path...)
}

func setInt64OrDeleteIfDefault(doc map[string]any, value int64, defaultValue int64, path ...string) {
	if value == defaultValue {
		deletePath(doc, path...)
		return
	}
	setValue(doc, value, path...)
}

func setBoolOrDeleteIfDefault(doc map[string]any, value bool, defaultValue bool, path ...string) {
	if value == defaultValue {
		deletePath(doc, path...)
		return
	}
	setValue(doc, value, path...)
}

func deletePath(doc map[string]any, path ...string) {
	if len(path) == 0 {
		return
	}
	deletePathRecursive(doc, path, 0)
}

func deletePathRecursive(node map[string]any, path []string, index int) bool {
	key := path[index]
	if index == len(path)-1 {
		delete(node, key)
		return len(node) == 0
	}
	child, ok := node[key].(map[string]any)
	if !ok {
		return false
	}
	if deletePathRecursive(child, path, index+1) {
		delete(node, key)
	}
	return len(node) == 0
}

func defaultServerHeartbeatTimeout(tcpMux bool) int64 {
	if tcpMux {
		return -1
	}
	return 90
}
