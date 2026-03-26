package client

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/config/source"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/policy/security"
)

func TestServiceConfigManagerGetAndUpdateSettingsPreservesInlineConfigs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "frpc.toml")
	if err := os.WriteFile(path, []byte(`
serverAddr = "127.0.0.1"
serverPort = 7000
webServer.addr = "127.0.0.1"
webServer.port = 7400
webServer.user = "admin"
webServer.password = "admin"
auth.method = "token"
auth.token = "old-token"

[[proxies]]
name = "ssh"
type = "tcp"
localIP = "127.0.0.1"
localPort = 22
remotePort = 6000
`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	mgr := &serviceConfigManager{
		svr: &Service{
			configFilePath: path,
			unsafeFeatures: security.NewUnsafeFeatures(nil),
		},
	}

	settings, err := mgr.GetSettings()
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}

	settings.ServerAddr = "frps.example.com"
	settings.ServerPort = 7001
	settings.StorePath = "./frpc_store.json"
	settings.WebServerPort = 7401
	settings.AuthToken = "new-token"

	if err := mgr.updateSettings(settings, false); err != nil {
		t.Fatalf("update settings: %v", err)
	}

	result, err := config.LoadClientConfigResult(path, false)
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}

	if result.Common.ServerAddr != "frps.example.com" || result.Common.ServerPort != 7001 {
		t.Fatalf("unexpected server endpoint: %s:%d", result.Common.ServerAddr, result.Common.ServerPort)
	}
	if result.Common.Store.Path != "./frpc_store.json" {
		t.Fatalf("unexpected store path: %s", result.Common.Store.Path)
	}
	if result.Common.WebServer.Port != 7401 {
		t.Fatalf("unexpected web server port: %d", result.Common.WebServer.Port)
	}
	if result.Common.Auth.Token != "new-token" {
		t.Fatalf("unexpected auth token: %s", result.Common.Auth.Token)
	}
	if len(result.Proxies) != 1 || result.Proxies[0].GetBaseConfig().Name != "ssh" {
		t.Fatalf("inline proxies were not preserved: %+v", result.Proxies)
	}
}

func TestBuildStoreSourceForClientSettingsResolvesRelativePath(t *testing.T) {
	cfgFile := filepath.Join(t.TempDir(), "frpc.toml")
	common := &v1.ClientCommonConfig{
		Store: v1.StoreConfig{Path: "./store.json"},
	}

	storeSource, err := BuildStoreSourceForClientSettings(common, cfgFile)
	if err != nil {
		t.Fatalf("build store source: %v", err)
	}
	if storeSource == nil {
		t.Fatal("expected store source")
	}
}

func TestServiceConfigManagerUpdateSettingsOmitsDefaultValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "frpc.toml")
	if err := os.WriteFile(path, []byte(`
auth.token = "token123"

[[proxies]]
name = "ssh"
type = "tcp"
localIP = "127.0.0.1"
localPort = 22
remotePort = 6000
`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	mgr := &serviceConfigManager{
		svr: &Service{
			configFilePath: path,
			unsafeFeatures: security.NewUnsafeFeatures(nil),
		},
	}

	settings, err := mgr.GetSettings()
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}

	settings.StorePath = "./frpc_store.json"

	if err := mgr.updateSettings(settings, false); err != nil {
		t.Fatalf("update settings: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	text := string(content)

	if !strings.Contains(text, "path = './frpc_store.json'") {
		t.Fatalf("expected store path to be written, got:\n%s", text)
	}

	unexpected := []string{
		"serverAddr",
		"serverPort",
		"natHoleStunServer",
		"loginFailExit",
		"udpPacketSize",
		"log.level",
		"log.maxDays",
		"transport.protocol",
		"transport.poolCount",
		"transport.tls.enable",
		"transport.quic.keepalivePeriod",
		"password = ''",
	}
	for _, item := range unexpected {
		if strings.Contains(text, item) {
			t.Fatalf("did not expect %q in saved config:\n%s", item, text)
		}
	}
}

func TestServiceConfigManagerUpdateSettingsReloadsRuntimeSources(t *testing.T) {
	path := filepath.Join(t.TempDir(), "frpc.toml")
	if err := os.WriteFile(path, []byte(`
auth.token = "token123"

[[proxies]]
name = "ssh"
type = "tcp"
localIP = "127.0.0.1"
localPort = 22
remotePort = 6000
`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	configSource := source.NewConfigSource()
	aggregator := source.NewAggregator(configSource)
	svr := &Service{
		configFilePath: path,
		unsafeFeatures: security.NewUnsafeFeatures(nil),
		configSource:   configSource,
		aggregator:     aggregator,
		reloadCommon:   &v1.ClientCommonConfig{},
	}
	mgr := &serviceConfigManager{svr: svr}

	settings, err := mgr.GetSettings()
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}

	settings.StorePath = "./frpc_store.json"

	if err := mgr.UpdateSettings(settings); err != nil {
		t.Fatalf("update settings: %v", err)
	}

	if svr.storeSource == nil {
		t.Fatal("expected store source to be reloaded")
	}
	if aggregator.StoreSource() == nil {
		t.Fatal("expected aggregator store source to be reloaded")
	}
}
