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

package loggerfactory

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// ConfigManager manages logging configurations and registered components
type ConfigManager struct {
	mu                sync.RWMutex
	logLevelMap       *map[string]string
	slogHandlerConfig SlogHandlerConfig
	// Track components that have requested loggers
	registeredComponents map[string]LoggerUser
}

// LoggerUser is an interface for components that use loggers and need updates
type LoggerUser interface {
	UpdateLogger()
}

var (
	configManagerInstance *ConfigManager
	once                  sync.Once
)

func GetConfigManager() *ConfigManager {
	once.Do(func() {
		m := make(map[string]string)
		configManagerInstance = &ConfigManager{
			logLevelMap:          &m,
			registeredComponents: make(map[string]LoggerUser),
		}
	})
	return configManagerInstance
}

// SetLogLevelMap sets the log level map.
func (cm *ConfigManager) SetLogLevelMap(levelMap *map[string]string) {
	// Make a copy of registered components to avoid holding the lock during notification
	var componentsToNotify []LoggerUser

	cm.mu.Lock()
	cm.logLevelMap = levelMap

	// Create a copy of the components to notify
	for _, component := range cm.registeredComponents {
		componentsToNotify = append(componentsToNotify, component)
	}
	cm.mu.Unlock()

	// Notify components after releasing the lock
	for _, component := range componentsToNotify {
		component.UpdateLogger()
	}
}

func (cm *ConfigManager) GetLogLevelMap() *map[string]string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.logLevelMap
}

// SetSlogHandlerConfig sets the slog handler configuration.
func (cm *ConfigManager) SetSlogHandlerConfig(config SlogHandlerConfig) {
	// Make a copy of registered components to avoid holding the lock during notification
	var componentsToNotify []LoggerUser

	cm.mu.Lock()
	cm.slogHandlerConfig = config

	// Create a copy of the components to notify
	for _, component := range cm.registeredComponents {
		componentsToNotify = append(componentsToNotify, component)
	}
	cm.mu.Unlock()

	// Notify components after releasing the lock
	for _, component := range componentsToNotify {
		component.UpdateLogger()
	}
}

func (cm *ConfigManager) GetSlogHandlerConfig() SlogHandlerConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.slogHandlerConfig
}

// RegisterLoggerUser registers a component that uses a logger
func (cm *ConfigManager) RegisterLoggerUser(packageName string, component LoggerUser) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, ok := cm.registeredComponents[packageName]; !ok {
		cm.registeredComponents[packageName] = component
	}
}

// Type to extract and hold the slog handler related configurations from Config
// format : json/text
// outputpath: stdout/file/stderr
type SlogHandlerConfig struct {
	// json,text
	Format string `koanf:"format"`
	// stdout, file
	OutputPath string `koanf:"outputPath"`
}

// Intentionally put 'slog' in future we can introduce more abstract handlers. Every handler should implement slog.Handler interface
func GetSlogHandler(slogHandlerConfig SlogHandlerConfig) slog.Handler {
	format := slogHandlerConfig.Format
	outputPath := slogHandlerConfig.OutputPath
	var slogHandler slog.Handler
	switch format {
	case "json":
		switch outputPath {
		case "stdout":
			slogHandler = slog.NewJSONHandler(os.Stdout, nil)
		case "file":
			// l.Handler = slog.NewJSONHandler(slog.File(outputPath), slog.DefaultTimeFormat)
		}
	case "text":
		switch outputPath {
		case "stdout":
			slogHandler = slog.NewTextHandler(os.Stdout, nil)
		case "file":
			// l.Handler = slog.NewTextHandler(slog.File(outputPath), slog.DefaultTimeFormat)
		}
	}
	return slogHandler
}

// A LevelHandler wraps a Handler with an Enabled method
// that returns false for levels below a minimum.
type LevelHandler struct {
	level   slog.Leveler
	handler slog.Handler
}

// NewLevelHandler returns a LevelHandler with the given level.
// All methods except Enabled delegate to h.
func NewLevelHandler(level slog.Leveler, h slog.Handler) *LevelHandler {
	// Optimization: avoid chains of LevelHandlers.
	if lh, ok := h.(*LevelHandler); ok {
		h = lh.Handler()
	}
	return &LevelHandler{level, h}
}

// Enabled implements Handler.Enabled by reporting whether
// level is at least as large as h's level.
func (h *LevelHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle implements Handler.Handle.
func (h *LevelHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.handler.Handle(ctx, r)
}

// WithAttrs implements Handler.WithAttrs.
func (h *LevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewLevelHandler(h.level, h.handler.WithAttrs(attrs))
}

// WithGroup implements Handler.WithGroup.
func (h *LevelHandler) WithGroup(name string) slog.Handler {
	return NewLevelHandler(h.level, h.handler.WithGroup(name))
}

// Handler returns the Handler wrapped by h.
func (h *LevelHandler) Handler() slog.Handler {
	return h.handler
}

// LevelFromString converts a string representation of a log level to a slog.Leveler.
func LevelFromString(levelStr string) slog.Leveler {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		// Return default level (e.g., Info) or handle invalid input as needed.
		return slog.LevelInfo // Or return an error, or a custom level.
	}
}

// GetLogger returns a logger for the specified package name and automatically
// registers the component if it implements LoggerUser
func GetLogger(packageName string, component interface{}) *slog.Logger {
	cm := GetConfigManager()

	// If the component implements LoggerUser, register it
	if loggerUser, ok := component.(LoggerUser); ok {
		cm.RegisterLoggerUser(packageName, loggerUser)
	}

	levelMap := cm.GetLogLevelMap()
	slogHandlerConfig := cm.GetSlogHandlerConfig()

	levelStr, ok := (*levelMap)[packageName]
	if !ok {
		slog.Error("PackageName not found in LevelMap", "PackageName", packageName)
		return slog.New(NewLevelHandler(slog.LevelDebug, GetSlogHandler(slogHandlerConfig)))
	}
	return slog.New(NewLevelHandler(LevelFromString(levelStr), GetSlogHandler(slogHandlerConfig)))
}
