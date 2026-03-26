package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveOptionalServerConfigPathExplicit(t *testing.T) {
	got, err := resolveOptionalServerConfigPath("custom.toml")
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}
	if got != "custom.toml" {
		t.Fatalf("unexpected config path: %s", got)
	}
}

func TestDefaultServerConfigCandidatesIncludesPreferredNames(t *testing.T) {
	candidates := defaultServerConfigCandidates()
	if len(candidates) == 0 {
		t.Fatal("expected at least one candidate")
	}

	expectedSuffixes := []string{
		"frps.toml",
		"frps.yaml",
		"frps.yml",
		"frps.json",
		"frps.ini",
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

func TestResolveRequiredServerConfigPathFindsConfigInWorkingDirectory(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "frps.toml")
	if err := os.WriteFile(configPath, []byte("bindPort = 7000\n"), 0o600); err != nil {
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

	got, err := resolveRequiredServerConfigPath("")
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}
	if got != configPath {
		t.Fatalf("unexpected config path: got %s want %s", got, configPath)
	}
}

func TestResolveOptionalServerConfigPathFallsBackToCommandLineMode(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	got, err := resolveOptionalServerConfigPath("")
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty path for command line mode, got %s", got)
	}
}
