package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ResolveUploadTarget(configFilePath string, targetPath string, filename string) (string, string, error) {
	if strings.TrimSpace(filename) == "" {
		return "", "", fmt.Errorf("filename is required")
	}
	if strings.TrimSpace(configFilePath) == "" {
		return "", "", fmt.Errorf("config file path is required")
	}

	configDir := filepath.Dir(configFilePath)
	savedPath := strings.TrimSpace(targetPath)
	if savedPath == "" {
		savedPath = filepath.ToSlash(filepath.Join("certs", filepath.Base(filename)))
	}

	var absolutePath string
	if filepath.IsAbs(savedPath) {
		absolutePath = filepath.Clean(savedPath)
	} else {
		absolutePath = filepath.Clean(filepath.Join(configDir, filepath.FromSlash(savedPath)))
	}
	return savedPath, absolutePath, nil
}

func WriteUploadTarget(configFilePath string, targetPath string, filename string, content []byte) (string, error) {
	savedPath, absolutePath, err := ResolveUploadTarget(configFilePath, targetPath, filename)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return "", fmt.Errorf("create upload directory: %w", err)
	}
	if err := os.WriteFile(absolutePath, content, 0o600); err != nil {
		return "", fmt.Errorf("write upload file: %w", err)
	}
	return savedPath, nil
}
