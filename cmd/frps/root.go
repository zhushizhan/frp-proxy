// Copyright 2018 fatedier, fatedier@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/fatedier/frp/pkg/config"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/config/v1/validation"
	"github.com/fatedier/frp/pkg/policy/security"
	"github.com/fatedier/frp/pkg/util/log"
	"github.com/fatedier/frp/pkg/util/system"
	"github.com/fatedier/frp/pkg/util/version"
	"github.com/fatedier/frp/server"
)

var (
	cfgFile          string
	showVersion      bool
	strictConfigMode bool
	allowUnsafe      []string

	serverCfg v1.ServerConfig
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file of frps")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "version of frps")
	rootCmd.PersistentFlags().BoolVarP(&strictConfigMode, "strict_config", "", true, "strict config parsing mode, unknown fields will cause errors")
	rootCmd.PersistentFlags().StringSliceVarP(&allowUnsafe, "allow-unsafe", "", []string{},
		fmt.Sprintf("allowed unsafe features, one or more of: %s", strings.Join(security.ServerUnsafeFeatures, ", ")))

	config.RegisterServerConfigFlags(rootCmd, &serverCfg)
}

var rootCmd = &cobra.Command{
	Use:   "frps",
	Short: "frps is the server of frp (https://github.com/fatedier/frp)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println(version.Full())
			return nil
		}

		resolvedCfgFile, err := resolveOptionalServerConfigPath(cfgFile)
		if err != nil {
			fmt.Println(err)
			system.PauseBeforeExit()
			os.Exit(1)
		}
		cfgFile = resolvedCfgFile

		var (
			svrCfg         *v1.ServerConfig
			isLegacyFormat bool
		)
		if cfgFile != "" {
			svrCfg, isLegacyFormat, err = config.LoadServerConfig(cfgFile, strictConfigMode)
			if err != nil {
				fmt.Println(err)
				system.PauseBeforeExit()
				os.Exit(1)
			}
			if isLegacyFormat {
				fmt.Printf("WARNING: ini format is deprecated and the support will be removed in the future, " +
					"please use yaml/json/toml format instead!\n")
			}
		} else {
			if err := serverCfg.Complete(); err != nil {
				fmt.Printf("failed to complete server config: %v\n", err)
				system.PauseBeforeExit()
				os.Exit(1)
			}
			svrCfg = &serverCfg
		}

		unsafeFeatures := security.NewUnsafeFeatures(allowUnsafe)
		validator := validation.NewConfigValidator(unsafeFeatures)
		warning, err := validator.ValidateServerConfig(svrCfg)
		if warning != nil {
			fmt.Printf("WARNING: %v\n", warning)
		}
		if err != nil {
			fmt.Println(err)
			system.PauseBeforeExit()
			os.Exit(1)
		}

		if err := runServer(svrCfg); err != nil {
			fmt.Println(err)
			system.PauseBeforeExit()
			os.Exit(1)
		}
		return nil
	},
}

func Execute() {
	rootCmd.SetGlobalNormalizationFunc(config.WordSepNormalizeFunc)
	if err := rootCmd.Execute(); err != nil {
		system.PauseBeforeExit()
		os.Exit(1)
	}
}

func resolveOptionalServerConfigPath(explicitPath string) (string, error) {
	if strings.TrimSpace(explicitPath) != "" {
		return explicitPath, nil
	}

	for _, candidate := range defaultServerConfigCandidates() {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", nil
}

func resolveRequiredServerConfigPath(explicitPath string) (string, error) {
	if strings.TrimSpace(explicitPath) != "" {
		return explicitPath, nil
	}

	candidates := defaultServerConfigCandidates()
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("frps: no config file found, tried: %s", strings.Join(candidates, ", "))
}

func defaultServerConfigCandidates() []string {
	names := []string{
		"frps.toml",
		"frps.yaml",
		"frps.yml",
		"frps.json",
		"frps.ini",
	}

	paths := make([]string, 0, len(names)*2)
	seen := map[string]struct{}{}
	addDir := func(dir string) {
		if strings.TrimSpace(dir) == "" {
			return
		}
		for _, name := range names {
			path := filepath.Join(dir, name)
			if _, ok := seen[path]; ok {
				continue
			}
			seen[path] = struct{}{}
			paths = append(paths, path)
		}
	}

	if exePath, err := os.Executable(); err == nil {
		addDir(filepath.Dir(exePath))
	}
	if cwd, err := os.Getwd(); err == nil {
		addDir(cwd)
	}
	return paths
}

func runServer(cfg *v1.ServerConfig) (err error) {
	log.InitLogger(cfg.Log.To, cfg.Log.Level, int(cfg.Log.MaxDays), cfg.Log.DisablePrintColor)

	if cfgFile != "" {
		log.Infof("frps uses config file: %s", cfgFile)
	} else {
		log.Infof("frps uses command line arguments for config")
	}

	svr, err := server.NewService(cfg)
	if err != nil {
		return err
	}
	exePath, _ := os.Executable()
	workDir, _ := os.Getwd()
	svr.SetConfigManager(server.NewFileConfigManager(server.FileConfigManagerOptions{
		Service:        svr,
		ConfigFilePath: cfgFile,
		ExecutablePath: exePath,
		Args:           os.Args[1:],
		WorkDir:        workDir,
		UnsafeFeatures: security.NewUnsafeFeatures(allowUnsafe),
	}))
	log.Infof("frps started successfully")
	svr.Run(context.Background())
	return
}
