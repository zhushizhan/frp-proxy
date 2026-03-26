package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/policy/security"
	sermodel "github.com/fatedier/frp/server/http/model"
)

func writeServerConfigFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "frps.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestFileConfigManagerGetSettingsIncludesAdvancedFields(t *testing.T) {
	path := writeServerConfigFile(t, `
bindAddr = "0.0.0.0"
bindPort = 7001
proxyBindAddr = "127.0.0.1"
kcpBindPort = 7002
quicBindPort = 7003
vhostHTTPPort = 80
vhostHTTPSPort = 443
vhostHTTPTimeout = 75
tcpmuxHTTPConnectPort = 1337
tcpmuxPassthrough = true
subDomainHost = "frps.example.com"
enablePrometheus = true
detailedErrorsToClient = false
maxPortsPerClient = 3
userConnTimeout = 15
udpPacketSize = 1400
natholeAnalysisDataReserveHours = 48
allowPorts = [{ start = 2000, end = 3000 }]
custom404Page = "./404.html"

auth.method = "oidc"
auth.additionalScopes = ["HeartBeats", "NewWorkConns"]
auth.oidc.issuer = "https://issuer.example.com"
auth.oidc.audience = "frps"
auth.oidc.skipExpiryCheck = true
auth.oidc.skipIssuerCheck = true

transport.tcpMux = false
transport.tcpMuxKeepaliveInterval = 45
transport.maxPoolCount = 8
transport.heartbeatTimeout = 95
transport.tcpKeepalive = 3600
transport.quic.keepalivePeriod = 12
transport.quic.maxIdleTimeout = 40
transport.quic.maxIncomingStreams = 64
transport.tls.force = true
transport.tls.certFile = "./transport.crt"
transport.tls.keyFile = "./transport.key"
transport.tls.trustedCaFile = "./transport-ca.crt"

webServer.addr = "127.0.0.1"
webServer.port = 7500
webServer.user = "admin"
webServer.password = "admin"
webServer.assetsDir = "./static"
webServer.pprofEnable = true
webServer.tls.certFile = "./dashboard.crt"
webServer.tls.keyFile = "./dashboard.key"
webServer.tls.trustedCaFile = "./dashboard-ca.crt"

log.to = "./frps.log"
log.level = "debug"
log.maxDays = 9
log.disablePrintColor = true

sshTunnelGateway.bindPort = 2200
sshTunnelGateway.privateKeyFile = "./id_rsa"
sshTunnelGateway.autoGenPrivateKeyPath = "./autogen"
sshTunnelGateway.authorizedKeysFile = "./authorized_keys"

[[httpPlugins]]
name = "user-manager"
addr = "127.0.0.1:9000"
path = "/handler"
ops = ["Login", "Ping"]
tlsVerify = true
`)

	mgr := NewFileConfigManager(FileConfigManagerOptions{
		ConfigFilePath: path,
		UnsafeFeatures: security.NewUnsafeFeatures(nil),
	})

	settings, err := mgr.GetSettings()
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}

	if settings.AuthMethod != "oidc" {
		t.Fatalf("unexpected auth method: %s", settings.AuthMethod)
	}
	if settings.TransportTLSCertFile != "./transport.crt" {
		t.Fatalf("unexpected transport tls cert: %s", settings.TransportTLSCertFile)
	}
	if settings.TCPMux {
		t.Fatal("expected tcpMux to be false")
	}
	if settings.QUICKeepalivePeriod != 12 {
		t.Fatalf("unexpected quic keepalive: %d", settings.QUICKeepalivePeriod)
	}
	if settings.DashboardAssetsDir != "./static" {
		t.Fatalf("unexpected dashboard assets dir: %s", settings.DashboardAssetsDir)
	}
	if settings.LogLevel != "debug" {
		t.Fatalf("unexpected log level: %s", settings.LogLevel)
	}
	if settings.SSHTunnelGatewayPort != 2200 {
		t.Fatalf("unexpected ssh gateway port: %d", settings.SSHTunnelGatewayPort)
	}
	if settings.NatHoleRetentionHours != 48 {
		t.Fatalf("unexpected nat hole retention: %d", settings.NatHoleRetentionHours)
	}
	if len(settings.HTTPPlugins) != 1 {
		t.Fatalf("unexpected http plugin count: %d", len(settings.HTTPPlugins))
	}
	if settings.HTTPPlugins[0].Name != "user-manager" {
		t.Fatalf("unexpected http plugin name: %s", settings.HTTPPlugins[0].Name)
	}
	if len(settings.HTTPPlugins[0].Ops) != 2 {
		t.Fatalf("unexpected http plugin ops: %#v", settings.HTTPPlugins[0].Ops)
	}
}

func TestFileConfigManagerUpdateSettingsWritesAdvancedFields(t *testing.T) {
	path := writeServerConfigFile(t, `
bindAddr = "0.0.0.0"
bindPort = 7001
webServer.addr = "127.0.0.1"
webServer.port = 7500
webServer.user = "admin"
webServer.password = "admin"
auth.method = "token"
auth.token = "old-token"
`)

	mgr := NewFileConfigManager(FileConfigManagerOptions{
		ConfigFilePath: path,
		UnsafeFeatures: security.NewUnsafeFeatures(nil),
	})

	err := mgr.UpdateSettings(sermodel.ServerSettings{
		BindAddr:               "127.0.0.1",
		BindPort:               7100,
		ProxyBindAddr:          "127.0.0.1",
		KCPBindPort:            7101,
		QUICBindPort:           7102,
		VhostHTTPPort:          8080,
		VhostHTTPSPort:         8443,
		VhostHTTPTimeout:       88,
		TCPMuxHTTPConnectPort:  1337,
		TCPMuxPassthrough:      true,
		SubdomainHost:          "frps.local",
		AuthMethod:             "token",
		AuthAdditionalScopes:   []string{"HeartBeats"},
		AuthTokenSourceType:    "file",
		AuthTokenSourceFile:    "./token.txt",
		TLSForce:               true,
		TransportTLSCertFile:   "./transport.crt",
		TransportTLSKeyFile:    "./transport.key",
		TransportTLSTrustedCA:  "./transport-ca.crt",
		TCPMux:                 false,
		TCPMuxKeepalive:        33,
		MaxPoolCount:           7,
		HeartbeatTimeout:       91,
		TCPKeepAlive:           120,
		QUICKeepalivePeriod:    14,
		QUICMaxIdleTimeout:     31,
		QUICMaxIncomingStreams: 80,
		MaxPortsPerClient:      6,
		UserConnTimeout:        19,
		UDPPacketSize:          1300,
		NatHoleRetentionHours:  72,
		DetailedErrorsToClient: false,
		AllowPorts:             "2000-3000,4000",
		EnablePrometheus:       true,
		DashboardAddr:          "0.0.0.0",
		DashboardPort:          7600,
		DashboardUser:          "new-admin",
		DashboardPassword:      "new-password",
		DashboardAssetsDir:     "./static",
		DashboardPprofEnable:   true,
		DashboardTLSCertFile:   "./dashboard.crt",
		DashboardTLSKeyFile:    "./dashboard.key",
		DashboardTLSTrustedCA:  "./dashboard-ca.crt",
		LogTo:                  "./frps.log",
		LogLevel:               "warn",
		LogMaxDays:             10,
		LogDisablePrintColor:   true,
		SSHTunnelGatewayPort:   2200,
		SSHPrivateKeyFile:      "./id_rsa",
		SSHAutoGenKeyPath:      "./autogen",
		SSHAuthorizedKeysFile:  "./authorized_keys",
		HTTPPlugins: []sermodel.HTTPPluginSettings{
			{
				Name:      "user-manager",
				Addr:      "127.0.0.1:9000",
				Path:      "/handler",
				Ops:       []string{"Login", "NewProxy"},
				TLSVerify: true,
			},
		},
		Custom404Page: "./404.html",
	})
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}

	cfg, _, err := config.LoadServerConfig(path, true)
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}

	if cfg.BindPort != 7100 || cfg.VhostHTTPSPort != 8443 {
		t.Fatalf("unexpected ports after save: bind=%d https=%d", cfg.BindPort, cfg.VhostHTTPSPort)
	}
	if cfg.Auth.TokenSource == nil || cfg.Auth.TokenSource.Type != "file" {
		t.Fatal("expected file-based token source")
	}
	if cfg.Auth.TokenSource.File == nil || cfg.Auth.TokenSource.File.Path != "./token.txt" {
		t.Fatalf("unexpected token source file: %#v", cfg.Auth.TokenSource.File)
	}
	if cfg.Transport.TLS.CertFile != "./transport.crt" {
		t.Fatalf("unexpected transport cert file: %s", cfg.Transport.TLS.CertFile)
	}
	if cfg.Transport.TCPMux == nil || *cfg.Transport.TCPMux {
		t.Fatal("expected tcpMux to be false")
	}
	if cfg.WebServer.TLS == nil || cfg.WebServer.TLS.CertFile != "./dashboard.crt" {
		t.Fatalf("unexpected dashboard tls: %#v", cfg.WebServer.TLS)
	}
	if cfg.Log.Level != "warn" || cfg.Log.To != "./frps.log" {
		t.Fatalf("unexpected log config: %+v", cfg.Log)
	}
	if cfg.SSHTunnelGateway.BindPort != 2200 {
		t.Fatalf("unexpected ssh gateway port: %d", cfg.SSHTunnelGateway.BindPort)
	}
	if cfg.UDPPacketSize != 1300 || cfg.NatHoleAnalysisDataReserveHours != 72 {
		t.Fatalf("unexpected runtime tuning values: udp=%d nathole=%d", cfg.UDPPacketSize, cfg.NatHoleAnalysisDataReserveHours)
	}
	if len(cfg.HTTPPlugins) != 1 {
		t.Fatalf("unexpected http plugin count: %d", len(cfg.HTTPPlugins))
	}
	if cfg.HTTPPlugins[0].Name != "user-manager" || cfg.HTTPPlugins[0].Addr != "127.0.0.1:9000" {
		t.Fatalf("unexpected http plugin: %#v", cfg.HTTPPlugins[0])
	}
	if len(cfg.HTTPPlugins[0].Ops) != 2 || cfg.HTTPPlugins[0].Ops[1] != "NewProxy" {
		t.Fatalf("unexpected http plugin ops: %#v", cfg.HTTPPlugins[0].Ops)
	}
}

func TestFileConfigManagerUpdateSettingsPreservesExecTokenSource(t *testing.T) {
	path := writeServerConfigFile(t, `
bindAddr = "0.0.0.0"
bindPort = 7001
webServer.addr = "127.0.0.1"
webServer.port = 7500
webServer.user = "admin"
webServer.password = "admin"
auth.method = "token"
auth.tokenSource.type = "exec"
auth.tokenSource.exec.command = "cmd"
auth.tokenSource.exec.args = ["/c", "echo token"]
`)

	mgr := NewFileConfigManager(FileConfigManagerOptions{
		ConfigFilePath: path,
		UnsafeFeatures: security.NewUnsafeFeatures([]string{security.TokenSourceExec}),
	})

	settings, err := mgr.GetSettings()
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}

	settings.BindPort = 7200
	if err := mgr.UpdateSettings(settings); err != nil {
		t.Fatalf("update settings: %v", err)
	}

	cfg, _, err := config.LoadServerConfig(path, true)
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}

	if cfg.BindPort != 7200 {
		t.Fatalf("unexpected bind port: %d", cfg.BindPort)
	}
	if cfg.Auth.TokenSource == nil || cfg.Auth.TokenSource.Type != "exec" {
		t.Fatal("expected exec token source to be preserved")
	}
	if cfg.Auth.TokenSource.Exec == nil || cfg.Auth.TokenSource.Exec.Command != "cmd" {
		t.Fatalf("unexpected exec token source: %#v", cfg.Auth.TokenSource.Exec)
	}
}
