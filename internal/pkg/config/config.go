/*
 *  Licensed to the Apache Software Foundation (ASF) under one
 *  or more contributor license agreements.  See the NOTICE file
 *  distributed with this work for additional information
 *  regarding copyright ownership.  The ASF licenses this file
 *  to you under the Apache License, Version 2.0 (the
 *  "License"); you may not use this file except in compliance
 *  with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *   * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied.  See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
	"github.com/apache/synapse-go/internal/pkg/core/utils"
	"github.com/apache/synapse-go/internal/pkg/loggerfactory"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config implements the common.ConfigProvider interface
type Config struct {
	koanf *koanf.Koanf
}

func ReadFile(filename string) (*Config, error) {
	k := koanf.New(".")
	f := file.Provider(filename)
	if err := k.Load(f, toml.Parser()); err != nil {
		return nil, err
	}
	cfg := &Config{
		koanf: k,
	}
	return cfg, nil
}

func (c *Config) IsSet(key string) bool {
	return c.koanf.Exists(key)
}

func (c *Config) Watch(ctx context.Context, filename string) {
	f := file.Provider(filename)

	f.Watch(func(event interface{}, err error) {
		if err != nil {
			log.Printf("watch error: %v", err)
			return
		}
		// Throw away the old config and load a fresh copy.
		log.Println("config changed. Reloading ...")
		new_k := koanf.New(".")
		if err := new_k.Load(f, toml.Parser()); err != nil {
			log.Printf("error loading new config: %v", err)
			return
		}
		// Update the config
		c.koanf = new_k

		// Update the logger configuration
		var levelMap map[string]string
		var slogHandlerConfig loggerfactory.SlogHandlerConfig

		c.MustUnmarshal("logger.level.packages", &levelMap)
		c.MustUnmarshal("logger.handler", &slogHandlerConfig)

		cm := loggerfactory.GetConfigManager()
		cm.SetLogLevelMap(&levelMap)
		cm.SetSlogHandlerConfig(slogHandlerConfig)
	})
}

func (c *Config) Unmarshal(key string, out interface{}) error {
	err := c.koanf.Unmarshal(key, out)
	if err != nil {
		return fmt.Errorf("cannot unmarshal config for key %q: %v", key, err)
	}
	return nil
}

func (c *Config) MustUnmarshal(key string, out interface{}) {
	err := c.Unmarshal(key, out)
	if err != nil {
		panic(err)
	}
}

func InitializeConfig(ctx context.Context, confFolderPath string) error {
	files, err := os.ReadDir(confFolderPath)
	configContext := ctx.Value(utils.ConfigContextKey).(*artifacts.ConfigContext)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no config files found in %s", confFolderPath)
	}
	// {"LoggerConfig", "deployment"}
	for _, configurationType := range []string{"LoggerConfig", "deployment"} {
		configFilePath := filepath.Join(confFolderPath, configurationType+".toml")
		cfg, err := ReadFile(configFilePath)
		if err != nil {
			return fmt.Errorf("cannot read config file: %w", err)
		}

		switch configurationType {
		case "LoggerConfig":
			var levelMap map[string]string
			var slogHandlerConfig loggerfactory.SlogHandlerConfig

			if cfg.IsSet("logger") {
				cfg.MustUnmarshal("logger.handler", &slogHandlerConfig)
				cfg.MustUnmarshal("logger.level.packages", &levelMap)
			}

			cm := loggerfactory.GetConfigManager()
			cm.SetLogLevelMap(&levelMap)
			cm.SetSlogHandlerConfig(slogHandlerConfig)

			// Start watching for config changes
			cfg.Watch(context.Background(), configFilePath)

		case "deployment":
			deploymentConfigMap := make(map[string]interface{})
			if cfg.IsSet("server") {
				var serverConfigMap map[string]string
				cfg.MustUnmarshal("server", &serverConfigMap)

				// Validate required hostname key
				hostname, exists := serverConfigMap["hostname"]
				if !exists {
					return fmt.Errorf("missing required server configuration key: hostname")
				}

				// Validate hostname value
				if hostname == "" {
					return fmt.Errorf("server hostname cannot be empty")
				}

				// Validate offset if it exists (optional)
				if offsetStr, hasOffset := serverConfigMap["offset"]; hasOffset && offsetStr != "" {
					offset, err := strconv.Atoi(offsetStr)
					if err != nil {
						return fmt.Errorf("invalid server offset value: %s, must be an integer", offsetStr)
					}
					if offset < 0 {
						return fmt.Errorf("server offset must be non-negative, got: %d", offset)
					}
				}
				deploymentConfigMap["server"] = serverConfigMap
			} else {
				return fmt.Errorf("server configuration section is required in deployment.toml")
			}

			configContext.AddDeploymentConfig(deploymentConfigMap)
		}
	}
	return nil
}
