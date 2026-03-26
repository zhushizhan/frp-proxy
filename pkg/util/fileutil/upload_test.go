package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveUploadTargetWithDefaultRelativePath(t *testing.T) {
	cfgFile := filepath.Join(t.TempDir(), "frps.toml")
	savedPath, absolutePath, err := ResolveUploadTarget(cfgFile, "", "server.crt")
	if err != nil {
		t.Fatalf("resolve upload target: %v", err)
	}
	if savedPath != "certs/server.crt" {
		t.Fatalf("unexpected saved path: %s", savedPath)
	}
	expected := filepath.Join(filepath.Dir(cfgFile), "certs", "server.crt")
	if absolutePath != expected {
		t.Fatalf("unexpected absolute path: got %s want %s", absolutePath, expected)
	}
}

func TestWriteUploadTargetCreatesParentDirectories(t *testing.T) {
	cfgFile := filepath.Join(t.TempDir(), "frpc.toml")
	if err := os.WriteFile(cfgFile, []byte("serverAddr = \"127.0.0.1\"\n"), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	savedPath, err := WriteUploadTarget(cfgFile, "./certs/client.crt", "client.crt", []byte("cert-data"))
	if err != nil {
		t.Fatalf("write upload target: %v", err)
	}
	if savedPath != "./certs/client.crt" {
		t.Fatalf("unexpected saved path: %s", savedPath)
	}

	absolutePath := filepath.Join(filepath.Dir(cfgFile), "certs", "client.crt")
	content, err := os.ReadFile(absolutePath)
	if err != nil {
		t.Fatalf("read uploaded file: %v", err)
	}
	if string(content) != "cert-data" {
		t.Fatalf("unexpected uploaded content: %s", string(content))
	}
}
