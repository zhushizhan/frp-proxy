package sub

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveClientConfigPathExplicit(t *testing.T) {
	got, err := resolveClientConfigPath("custom.toml")
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}
	if got != "custom.toml" {
		t.Fatalf("unexpected config path: %s", got)
	}
}

func TestDefaultClientConfigCandidatesIncludesPreferredNames(t *testing.T) {
	candidates := defaultClientConfigCandidates()
	if len(candidates) == 0 {
		t.Fatal("expected at least one candidate")
	}

	expectedSuffixes := []string{
		"frpc.toml",
		"frpc.yaml",
		"frpc.yml",
		"frpc.json",
		"frpc.ini",
	}
	for _, suffix := range expectedSuffixes {
		found := false
		for _, candidate := range candidates {
			if filepath.Base(candidate) == suffix {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing candidate with suffix %s", suffix)
		}
	}
}

func TestResolveClientConfigPathFindsConfigInWorkingDirectory(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "frpc.toml")
	if err := os.WriteFile(configPath, []byte("serverAddr = \"127.0.0.1\"\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	got, err := resolveClientConfigPath("")
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}
	if got != configPath {
		t.Fatalf("unexpected config path: got %s want %s", got, configPath)
	}
}
