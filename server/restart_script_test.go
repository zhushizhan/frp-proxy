package server

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEnsureFRPSServiceScriptWritesExpectedFile(t *testing.T) {
	exeName := "frps"
	if runtime.GOOS == "windows" {
		exeName = "frps.exe"
	}

	exePath := filepath.Join(t.TempDir(), exeName)
	scriptPath, err := ensureFRPSServiceScript(exePath)
	if err != nil {
		t.Fatalf("ensure script: %v", err)
	}

	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(scriptPath, "frps-service.ps1") {
			t.Fatalf("unexpected script path: %s", scriptPath)
		}
	} else {
		if !strings.HasSuffix(scriptPath, "frps-service.sh") {
			t.Fatalf("unexpected script path: %s", scriptPath)
		}
	}
}

func TestBuildFRPSServiceCommandIncludesConfigEnv(t *testing.T) {
	scriptPath := filepath.Join(t.TempDir(), "frps-service")
	exePath := filepath.Join(t.TempDir(), "frps")
	configPath := filepath.Join(t.TempDir(), "frps.toml")
	workDir := filepath.Dir(configPath)

	cmd, err := buildFRPSServiceCommand(scriptPath, 12345, exePath, configPath, workDir)
	if err != nil {
		t.Fatalf("build command: %v", err)
	}

	env := strings.Join(cmd.Env, "\n")
	for _, item := range []string{
		"FRPS_EXE_PATH=" + exePath,
		"FRPS_CONFIG_PATH=" + configPath,
		"FRPS_WORK_DIR=" + workDir,
		"FRPS_STOP_WAIT_SECONDS=30",
	} {
		if !strings.Contains(env, item) {
			t.Fatalf("expected env %q in command env", item)
		}
	}
}
