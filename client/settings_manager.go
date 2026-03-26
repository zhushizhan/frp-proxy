package client

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	toml "github.com/pelletier/go-toml/v2"
	yaml "sigs.k8s.io/yaml"

	"github.com/fatedier/frp/client/configmgmt"
	clientmodel "github.com/fatedier/frp/client/http/model"
	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/config/source"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/config/v1/validation"
	"github.com/fatedier/frp/pkg/policy/security"
	"github.com/fatedier/frp/pkg/util/fileutil"
	"github.com/fatedier/frp/pkg/util/log"
)

func (m *serviceConfigManager) GetSettings() (clientmodel.ClientSettings, error) {
	cfg, isLegacy, err := loadClientConfigFileForSettings(m.svr.configFilePath)
	if err != nil {
		return clientmodel.ClientSettings{}, err
	}
	if isLegacy {
		return clientmodel.ClientSettings{}, fmt.Errorf("legacy ini format is not supported for UI settings")
	}
	return clientSettingsFromConfig(m.svr.configFilePath, cfg), nil
}

func (m *serviceConfigManager) UpdateSettings(in clientmodel.ClientSettings) error {
	if err := m.updateSettings(in, false); err != nil {
		return err
	}
	return m.ReloadFromFile(false)
}

func (m *serviceConfigManager) UploadFile(targetPath string, filename string, content []byte) (string, error) {
	if m.svr.configFilePath == "" {
		return "", fmt.Errorf("%w: frpc has no config file path", configmgmt.ErrInvalidArgument)
	}
	return fileutil.WriteUploadTarget(m.svr.configFilePath, targetPath, filename, content)
}

func (m *serviceConfigManager) updateSettings(in clientmodel.ClientSettings, shouldRestart bool) error {
	if m.svr.configFilePath == "" {
		return fmt.Errorf("%w: frpc has no config file path", configmgmt.ErrInvalidArgument)
	}

	cfg, isLegacy, err := loadClientConfigFileForSettings(m.svr.configFilePath)
	if err != nil {
		return fmt.Errorf("%w: %v", configmgmt.ErrInvalidArgument, err)
	}
	if isLegacy {
		return fmt.Errorf("%w: legacy ini format is not supported for UI settings", configmgmt.ErrInvalidArgument)
	}

	doc, format, err := loadClientConfigDocument(m.svr.configFilePath)
	if err != nil {
		return fmt.Errorf("%w: %v", configmgmt.ErrInvalidArgument, err)
	}

	applyClientSettingsToConfig(cfg, in)
	applyClientSettingsToDocument(doc, in)

	content, err := marshalClientConfigDocument(format, doc)
	if err != nil {
		return fmt.Errorf("%w: %v", configmgmt.ErrInvalidArgument, err)
	}

	if err := validateClientConfigContent(m.svr.configFilePath, content, m.svr.unsafeFeatures); err != nil {
		return fmt.Errorf("%w: %v", configmgmt.ErrInvalidArgument, err)
	}

	if err := os.WriteFile(m.svr.configFilePath, content, 0o600); err != nil {
		return err
	}

	if shouldRestart {
		m.scheduleRestart()
	}
	return nil
}

func (m *serviceConfigManager) scheduleRestart() {
	if !m.restartScheduled.CompareAndSwap(false, true) {
		return
	}

	go func() {
		exePath, _ := os.Executable()
		workDir, _ := os.Getwd()

		cmd := exec.Command(exePath, os.Args[1:]...)
		cmd.Dir = workDir
		cmd.Env = append(os.Environ(), "FRP_STARTUP_DELAY_MS=1000")
		if err := cmd.Start(); err != nil {
			log.Errorf("failed to schedule frpc restart: %v", err)
			m.restartScheduled.Store(false)
			return
		}

		time.Sleep(200 * time.Millisecond)
		m.svr.GracefulClose(200 * time.Millisecond)
	}()
}

func loadClientConfigFileForSettings(path string) (*v1.ClientConfig, bool, error) {
	if path == "" {
		return nil, false, fmt.Errorf("frpc has no config file path")
	}
	if config.DetectLegacyINIFormatFromFile(path) {
		return nil, true, nil
	}

	cfg := &v1.ClientConfig{}
	if err := config.LoadConfigureFromFile(path, cfg, false); err != nil {
		return nil, false, err
	}
	if err := cfg.ClientCommonConfig.Complete(); err != nil {
		return nil, false, err
	}
	return cfg, false, nil
}

func loadClientConfigDocument(path string) (map[string]any, string, error) {
	if path == "" {
		return nil, "", fmt.Errorf("frpc has no config file path")
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

func marshalClientConfigDocument(format string, doc map[string]any) ([]byte, error) {
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

func validateClientConfigContent(path string, content []byte, unsafeFeatures *security.UnsafeFeatures) error {
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	tmpFile, err := os.CreateTemp(dir, "frpc-settings-*"+ext)
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

	result, err := config.LoadClientConfigResult(tmpPath, false)
	if err != nil {
		return err
	}

	proxyCfgs, visitorCfgs := config.FilterClientConfigurers(result.Common, result.Proxies, result.Visitors)
	proxyCfgs = config.CompleteProxyConfigurers(proxyCfgs)
	visitorCfgs = config.CompleteVisitorConfigurers(visitorCfgs)

	_, err = validation.ValidateAllClientConfig(result.Common, proxyCfgs, visitorCfgs, unsafeFeatures)
	return err
}

func clientSettingsFromConfig(path string, cfg *v1.ClientConfig) clientmodel.ClientSettings {
	settings := clientmodel.ClientSettings{
		ConfigPath:                            path,
		User:                                  cfg.User,
		ClientID:                              cfg.ClientID,
		ServerAddr:                            cfg.ServerAddr,
		ServerPort:                            cfg.ServerPort,
		NatHoleSTUNServer:                     cfg.NatHoleSTUNServer,
		DNSServer:                             cfg.DNSServer,
		LoginFailExit:                         cfg.LoginFailExit != nil && *cfg.LoginFailExit,
		AuthMethod:                            string(cfg.Auth.Method),
		AuthAdditionalScopes:                  authScopesToStringSlice(cfg.Auth.AdditionalScopes),
		AuthToken:                             cfg.Auth.Token,
		AuthTokenSourceType:                   authTokenSourceType(cfg.Auth.TokenSource),
		OIDCClientID:                          cfg.Auth.OIDC.ClientID,
		OIDCClientSecret:                      cfg.Auth.OIDC.ClientSecret,
		OIDCAudience:                          cfg.Auth.OIDC.Audience,
		OIDCScope:                             cfg.Auth.OIDC.Scope,
		OIDCTokenEndpointURL:                  cfg.Auth.OIDC.TokenEndpointURL,
		OIDCTrustedCaFile:                     cfg.Auth.OIDC.TrustedCaFile,
		OIDCInsecureSkipVerify:                cfg.Auth.OIDC.InsecureSkipVerify,
		OIDCProxyURL:                          cfg.Auth.OIDC.ProxyURL,
		WebServerAddr:                         cfg.WebServer.Addr,
		WebServerPort:                         cfg.WebServer.Port,
		WebServerUser:                         cfg.WebServer.User,
		WebServerPassword:                     cfg.WebServer.Password,
		WebServerPprofEnable:                  cfg.WebServer.PprofEnable,
		StorePath:                             cfg.Store.Path,
		LogTo:                                 cfg.Log.To,
		LogLevel:                              cfg.Log.Level,
		LogMaxDays:                            cfg.Log.MaxDays,
		LogDisablePrintColor:                  cfg.Log.DisablePrintColor,
		TransportProtocol:                     cfg.Transport.Protocol,
		TransportPoolCount:                    cfg.Transport.PoolCount,
		TransportTCPMux:                       boolPtrValue(cfg.Transport.TCPMux),
		TransportTCPMuxKeepaliveInterval:      cfg.Transport.TCPMuxKeepaliveInterval,
		TransportDialServerTimeout:            cfg.Transport.DialServerTimeout,
		TransportDialServerKeepalive:          cfg.Transport.DialServerKeepAlive,
		TransportConnectServerLocalIP:         cfg.Transport.ConnectServerLocalIP,
		TransportProxyURL:                     cfg.Transport.ProxyURL,
		TransportHeartbeatInterval:            cfg.Transport.HeartbeatInterval,
		TransportHeartbeatTimeout:             cfg.Transport.HeartbeatTimeout,
		TransportTLSEnable:                    boolPtrValue(cfg.Transport.TLS.Enable),
		TransportTLSDisableCustomTLSFirstByte: boolPtrValue(cfg.Transport.TLS.DisableCustomTLSFirstByte),
		TransportTLSCertFile:                  cfg.Transport.TLS.CertFile,
		TransportTLSKeyFile:                   cfg.Transport.TLS.KeyFile,
		TransportTLSTrustedCaFile:             cfg.Transport.TLS.TrustedCaFile,
		TransportTLSServerName:                cfg.Transport.TLS.ServerName,
		TransportQUICKeepalivePeriod:          cfg.Transport.QUIC.KeepalivePeriod,
		TransportQUICMaxIdleTimeout:           cfg.Transport.QUIC.MaxIdleTimeout,
		TransportQUICMaxIncomingStreams:       cfg.Transport.QUIC.MaxIncomingStreams,
		UDPPacketSize:                         cfg.UDPPacketSize,
	}
	if cfg.Auth.TokenSource != nil && cfg.Auth.TokenSource.Type == "file" && cfg.Auth.TokenSource.File != nil {
		settings.AuthTokenSourceFile = cfg.Auth.TokenSource.File.Path
	}
	return settings
}

func applyClientSettingsToConfig(cfg *v1.ClientConfig, in clientmodel.ClientSettings) {
	cfg.User = strings.TrimSpace(in.User)
	cfg.ClientID = strings.TrimSpace(in.ClientID)
	cfg.ServerAddr = strings.TrimSpace(in.ServerAddr)
	cfg.ServerPort = in.ServerPort
	cfg.NatHoleSTUNServer = strings.TrimSpace(in.NatHoleSTUNServer)
	cfg.DNSServer = strings.TrimSpace(in.DNSServer)
	cfg.LoginFailExit = boolPtr(in.LoginFailExit)
	cfg.Auth.Method = v1.AuthMethod(strings.TrimSpace(in.AuthMethod))
	if cfg.Auth.Method == "" {
		cfg.Auth.Method = v1.AuthMethodToken
	}
	cfg.Auth.AdditionalScopes = stringSliceToAuthScopes(in.AuthAdditionalScopes)
	switch cfg.Auth.Method {
	case v1.AuthMethodToken:
		switch strings.TrimSpace(in.AuthTokenSourceType) {
		case "", "inline":
			cfg.Auth.Token = strings.TrimSpace(in.AuthToken)
			cfg.Auth.TokenSource = nil
		case "file":
			cfg.Auth.Token = ""
			cfg.Auth.TokenSource = &v1.ValueSource{
				Type: "file",
				File: &v1.FileSource{Path: strings.TrimSpace(in.AuthTokenSourceFile)},
			}
		case "exec":
			// Preserve existing exec source if configured.
			if cfg.Auth.TokenSource == nil || cfg.Auth.TokenSource.Type != "exec" {
				cfg.Auth.Token = strings.TrimSpace(in.AuthToken)
				cfg.Auth.TokenSource = nil
			}
		}
	case v1.AuthMethodOIDC:
		cfg.Auth.Token = ""
		cfg.Auth.TokenSource = nil
		cfg.Auth.OIDC.ClientID = strings.TrimSpace(in.OIDCClientID)
		cfg.Auth.OIDC.ClientSecret = in.OIDCClientSecret
		cfg.Auth.OIDC.Audience = strings.TrimSpace(in.OIDCAudience)
		cfg.Auth.OIDC.Scope = strings.TrimSpace(in.OIDCScope)
		cfg.Auth.OIDC.TokenEndpointURL = strings.TrimSpace(in.OIDCTokenEndpointURL)
		cfg.Auth.OIDC.TrustedCaFile = strings.TrimSpace(in.OIDCTrustedCaFile)
		cfg.Auth.OIDC.InsecureSkipVerify = in.OIDCInsecureSkipVerify
		cfg.Auth.OIDC.ProxyURL = strings.TrimSpace(in.OIDCProxyURL)
	}
	cfg.WebServer.Addr = strings.TrimSpace(in.WebServerAddr)
	cfg.WebServer.Port = in.WebServerPort
	cfg.WebServer.User = strings.TrimSpace(in.WebServerUser)
	cfg.WebServer.Password = in.WebServerPassword
	cfg.WebServer.PprofEnable = in.WebServerPprofEnable
	cfg.Store.Path = strings.TrimSpace(in.StorePath)
	cfg.Log.To = strings.TrimSpace(in.LogTo)
	cfg.Log.Level = strings.TrimSpace(in.LogLevel)
	cfg.Log.MaxDays = in.LogMaxDays
	cfg.Log.DisablePrintColor = in.LogDisablePrintColor
	cfg.Transport.Protocol = strings.TrimSpace(in.TransportProtocol)
	cfg.Transport.PoolCount = in.TransportPoolCount
	cfg.Transport.TCPMux = boolPtr(in.TransportTCPMux)
	cfg.Transport.TCPMuxKeepaliveInterval = in.TransportTCPMuxKeepaliveInterval
	cfg.Transport.DialServerTimeout = in.TransportDialServerTimeout
	cfg.Transport.DialServerKeepAlive = in.TransportDialServerKeepalive
	cfg.Transport.ConnectServerLocalIP = strings.TrimSpace(in.TransportConnectServerLocalIP)
	cfg.Transport.ProxyURL = strings.TrimSpace(in.TransportProxyURL)
	cfg.Transport.HeartbeatInterval = in.TransportHeartbeatInterval
	cfg.Transport.HeartbeatTimeout = in.TransportHeartbeatTimeout
	cfg.Transport.TLS.Enable = boolPtr(in.TransportTLSEnable)
	cfg.Transport.TLS.DisableCustomTLSFirstByte = boolPtr(in.TransportTLSDisableCustomTLSFirstByte)
	cfg.Transport.TLS.CertFile = strings.TrimSpace(in.TransportTLSCertFile)
	cfg.Transport.TLS.KeyFile = strings.TrimSpace(in.TransportTLSKeyFile)
	cfg.Transport.TLS.TrustedCaFile = strings.TrimSpace(in.TransportTLSTrustedCaFile)
	cfg.Transport.TLS.ServerName = strings.TrimSpace(in.TransportTLSServerName)
	cfg.Transport.QUIC.KeepalivePeriod = in.TransportQUICKeepalivePeriod
	cfg.Transport.QUIC.MaxIdleTimeout = in.TransportQUICMaxIdleTimeout
	cfg.Transport.QUIC.MaxIncomingStreams = in.TransportQUICMaxIncomingStreams
	cfg.UDPPacketSize = in.UDPPacketSize
}

func applyClientSettingsToDocument(doc map[string]any, in clientmodel.ClientSettings) {
	setStringOrDelete(doc, strings.TrimSpace(in.User), "user")
	setStringOrDelete(doc, strings.TrimSpace(in.ClientID), "clientID")
	setStringOrDeleteIfDefault(doc, strings.TrimSpace(in.ServerAddr), "0.0.0.0", "serverAddr")
	setIntOrDeleteIfDefault(doc, in.ServerPort, 7000, "serverPort")
	setStringOrDeleteIfDefault(doc, strings.TrimSpace(in.NatHoleSTUNServer), "stun.easyvoip.com:3478", "natHoleStunServer")
	setStringOrDelete(doc, strings.TrimSpace(in.DNSServer), "dnsServer")
	setBoolOrDeleteIfDefault(doc, in.LoginFailExit, true, "loginFailExit")
	setInt64OrDeleteIfDefault(doc, in.UDPPacketSize, 1500, "udpPacketSize")

	setStringOrDelete(doc, strings.TrimSpace(in.StorePath), "store", "path")

	setStringOrDeleteIfDefault(doc, strings.TrimSpace(in.AuthMethod), "token", "auth", "method")
	setStringSliceOrDelete(doc, in.AuthAdditionalScopes, "auth", "additionalScopes")
	switch strings.TrimSpace(in.AuthMethod) {
	case "", "token":
		setTokenSettings(doc, in)
	case "oidc":
		deletePath(doc, "auth", "token")
		deletePath(doc, "auth", "tokenSource")
		setOIDCSettings(doc, in)
	}

	setStringOrDeleteIfDefault(doc, strings.TrimSpace(in.WebServerAddr), "127.0.0.1", "webServer", "addr")
	setIntOrDeleteIfDefault(doc, in.WebServerPort, 0, "webServer", "port")
	setStringOrDelete(doc, strings.TrimSpace(in.WebServerUser), "webServer", "user")
	setStringOrDelete(doc, strings.TrimSpace(in.WebServerPassword), "webServer", "password")
	setBoolOrDeleteIfDefault(doc, in.WebServerPprofEnable, false, "webServer", "pprofEnable")

	setStringOrDeleteIfDefault(doc, strings.TrimSpace(in.LogTo), "console", "log", "to")
	setStringOrDeleteIfDefault(doc, strings.TrimSpace(in.LogLevel), "info", "log", "level")
	setInt64OrDeleteIfDefault(doc, in.LogMaxDays, 3, "log", "maxDays")
	setBoolOrDeleteIfDefault(doc, in.LogDisablePrintColor, false, "log", "disablePrintColor")

	setStringOrDeleteIfDefault(doc, strings.TrimSpace(in.TransportProtocol), "tcp", "transport", "protocol")
	setIntOrDeleteIfDefault(doc, in.TransportPoolCount, 1, "transport", "poolCount")
	setBoolOrDeleteIfDefault(doc, in.TransportTCPMux, true, "transport", "tcpMux")
	setInt64OrDeleteIfDefault(doc, in.TransportTCPMuxKeepaliveInterval, 30, "transport", "tcpMuxKeepaliveInterval")
	setInt64OrDeleteIfDefault(doc, in.TransportDialServerTimeout, 10, "transport", "dialServerTimeout")
	setInt64OrDeleteIfDefault(doc, in.TransportDialServerKeepalive, 7200, "transport", "dialServerKeepalive")
	setStringOrDelete(doc, strings.TrimSpace(in.TransportConnectServerLocalIP), "transport", "connectServerLocalIP")
	setStringOrDelete(doc, strings.TrimSpace(in.TransportProxyURL), "transport", "proxyURL")
	setInt64OrDeleteIfDefault(doc, in.TransportHeartbeatInterval, defaultClientHeartbeatInterval(in.TransportTCPMux), "transport", "heartbeatInterval")
	setInt64OrDeleteIfDefault(doc, in.TransportHeartbeatTimeout, defaultClientHeartbeatTimeout(in.TransportTCPMux), "transport", "heartbeatTimeout")
	setBoolOrDeleteIfDefault(doc, in.TransportTLSEnable, true, "transport", "tls", "enable")
	setBoolOrDeleteIfDefault(doc, in.TransportTLSDisableCustomTLSFirstByte, true, "transport", "tls", "disableCustomTLSFirstByte")
	setStringOrDelete(doc, strings.TrimSpace(in.TransportTLSCertFile), "transport", "tls", "certFile")
	setStringOrDelete(doc, strings.TrimSpace(in.TransportTLSKeyFile), "transport", "tls", "keyFile")
	setStringOrDelete(doc, strings.TrimSpace(in.TransportTLSTrustedCaFile), "transport", "tls", "trustedCaFile")
	setStringOrDelete(doc, strings.TrimSpace(in.TransportTLSServerName), "transport", "tls", "serverName")
	setIntOrDeleteIfDefault(doc, in.TransportQUICKeepalivePeriod, 10, "transport", "quic", "keepalivePeriod")
	setIntOrDeleteIfDefault(doc, in.TransportQUICMaxIdleTimeout, 30, "transport", "quic", "maxIdleTimeout")
	setIntOrDeleteIfDefault(doc, in.TransportQUICMaxIncomingStreams, 100000, "transport", "quic", "maxIncomingStreams")
}

func setTokenSettings(doc map[string]any, in clientmodel.ClientSettings) {
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

func setOIDCSettings(doc map[string]any, in clientmodel.ClientSettings) {
	deletePath(doc, "auth", "token")
	deletePath(doc, "auth", "tokenSource")
	setStringOrDelete(doc, strings.TrimSpace(in.OIDCClientID), "auth", "oidc", "clientID")
	setStringOrDelete(doc, strings.TrimSpace(in.OIDCClientSecret), "auth", "oidc", "clientSecret")
	setStringOrDelete(doc, strings.TrimSpace(in.OIDCAudience), "auth", "oidc", "audience")
	setStringOrDelete(doc, strings.TrimSpace(in.OIDCScope), "auth", "oidc", "scope")
	setStringOrDelete(doc, strings.TrimSpace(in.OIDCTokenEndpointURL), "auth", "oidc", "tokenEndpointURL")
	setStringOrDelete(doc, strings.TrimSpace(in.OIDCTrustedCaFile), "auth", "oidc", "trustedCaFile")
	setBoolOrDeleteIfDefault(doc, in.OIDCInsecureSkipVerify, false, "auth", "oidc", "insecureSkipVerify")
	setStringOrDelete(doc, strings.TrimSpace(in.OIDCProxyURL), "auth", "oidc", "proxyURL")
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

func buildStoreSource(common *v1.ClientCommonConfig, cfgFilePath string) (*source.StoreSource, error) {
	if common == nil || !common.Store.IsEnabled() {
		return nil, nil
	}

	storePath := common.Store.Path
	if storePath != "" && cfgFilePath != "" && !filepath.IsAbs(storePath) {
		storePath = filepath.Join(filepath.Dir(cfgFilePath), storePath)
	}

	return source.NewStoreSource(source.StoreSourceConfig{Path: storePath})
}

func BuildStoreSourceForClientSettings(common *v1.ClientCommonConfig, cfgFilePath string) (*source.StoreSource, error) {
	return buildStoreSource(common, cfgFilePath)
}

func defaultClientHeartbeatInterval(tcpMux bool) int64 {
	if tcpMux {
		return -1
	}
	return 30
}

func defaultClientHeartbeatTimeout(tcpMux bool) int64 {
	if tcpMux {
		return -1
	}
	return 90
}
