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

package file

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/apache/synapse-go/internal/app/core/domain"
	"github.com/apache/synapse-go/internal/app/core/ports"
	"github.com/apache/synapse-go/internal/pkg/core/synctx"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// FileInboundEndpoint handles file-based inbound operations
type FileInboundEndpoint struct {
	config          domain.InboundConfig
	isRunning       bool
	clock           FileClock
	mediator        ports.InboundMessageMediator
	processingFiles sync.Map
	protocolHandler ProtocolHandler
}

// NewFileInboundEndpoint creates a new FileInboundEndpoint instance
func NewFileInboundEndpoint(
	config domain.InboundConfig,
	mediator ports.InboundMessageMediator,
) *FileInboundEndpoint {
	return &FileInboundEndpoint{
		config:   config,
		clock:    NewFileClock(),
		mediator: mediator,
	}
}

func (f *FileInboundEndpoint) Start(ctx context.Context, mediator ports.InboundMessageMediator) error {
	// Check if context is already canceled before proceeding
	select {
	case <-ctx.Done():
		// Context already canceled, don't decrement WaitGroup
		return ctx.Err()
	default:
		// Context still valid, proceed with normal operation
	}

	if err := f.validateConfig(); err != nil {
		slog.Error("invalid configuration", "error", err)
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	f.mediator = mediator
	vfsFactory := &VFSProtocolHandlerFactory{}
	handler, err := vfsFactory.CreateHandler(f.config)
	if err != nil {
		slog.Error("failed to create protocol handler", "error", err)
		return fmt.Errorf("failed to create protocol handler: %w", err)
	}
	f.protocolHandler = handler

	slog.Info("starting file inbound endpoint")

	// Start polling
	err = f.poll(ctx)

	// When context is cancelled, wait for all processing to complete
	slog.Info("waiting for in-progress file operations to complete")

	return err
}

// Call this using a channel
func (f *FileInboundEndpoint) Stop() error {
	slog.Info("stopping file inbound endpoint")
	f.isRunning = false
	return nil
}

func (f *FileInboundEndpoint) poll(ctx context.Context) error {
	interval, err := strconv.Atoi(f.config.Parameters["interval"])
	if err != nil {
		slog.Error("invalid interval value", "error", err)
		return fmt.Errorf("invalid interval value: %w", err)
	}
	ticker := f.clock.NewTicker(time.Duration(interval) * time.Millisecond)
	defer ticker.Stop()

	processingWg := &sync.WaitGroup{}

	for {
		select {
		case <-ctx.Done():
			slog.Info("received shutdown signal, stopping file polling")
			// Wait for all processing to complete before returning
			processingWg.Wait()
			return ctx.Err()

		case <-ticker.C:
			processingWg.Add(1)
			go func() {
				defer processingWg.Done()
				if err := f.processingCycle(ctx); err != nil {
					if err != context.Canceled {
						slog.Error("error in processing cycle", "error", err)
					}
				}
			}()
		}
	}
}

func (f *FileInboundEndpoint) processingCycle(ctx context.Context) error {

	pattern, exists := f.config.Parameters["transport.vfs.FileNamePattern"]
	if !exists {
		pattern = ".*"
	}

	files, err := f.protocolHandler.ListFiles(pattern)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	// Process the files
	return f.processFiles(ctx, files)
}

func (f *FileInboundEndpoint) processFiles(ctx context.Context, files []string) error {
	// Check if sequential processing is required
	sequential := false
	if val, exists := f.config.Parameters["sequential"]; exists {
		var err error
		sequential, err = strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid sequential value: must be true/false, got '%s', defaulting to false", val)
		}
	}

	fileWg := &sync.WaitGroup{}

	for _, file := range files {
		select {
		case <-ctx.Done():
			slog.Info("cancelling remaining file processing")
			// Wait for all processing to complete before returning
			fileWg.Wait()
			return ctx.Err()
		default:
			// Check if file is already being processed
			if _, exists := f.processingFiles.LoadOrStore(file, true); exists {
				slog.Debug("skipping file - already being processed", "file", file)
				continue
			}

			if sequential {
				if err := f.processFile(ctx, file); err != nil {
					slog.Error("failed to process file", "error", err)
				}
			} else {
				fileWg.Add(1)
				go func(fileName string) {
					defer fileWg.Done()
					if err := f.processFile(ctx, fileName); err != nil {
						slog.Error("failed to process file", "error", err)
					}
				}(file)
			}
		}
	}
	return nil
}

func (f *FileInboundEndpoint) processFile(ctx context.Context, fileURI string) error {

	defer f.processingFiles.Delete(fileURI)

	fileSize, err := f.protocolHandler.GetSize(fileURI)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}

	lastModified, err := f.protocolHandler.GetLastModified(fileURI)
	if err != nil {
		return fmt.Errorf("failed to get last modified: %w", err)
	}

	filePath := f.protocolHandler.GetPath(fileURI)
	fileName := f.protocolHandler.GetName(fileURI)

	// Set up headers with file metadata
	headers := map[string]string{
		"FILE_LENGTH":   fmt.Sprintf("%d", fileSize),
		"LAST_MODIFIED": fmt.Sprintf("%d", lastModified.Unix()),
		"FILE_URI":      fileURI,
		"FILE_PATH":     filePath,
		"FILE_NAME":     fileName,
	}

	properties := map[string]interface{}{
		"isInbound":            "true",
		"ARTIFACT_NAME":        "inboundendpointfile",
		"inboundEndpointName":  "file",
		"ClientApiNonBlocking": "true",
	}

	// Create a message context with metadata but no content yet
	msgContext := &synctx.MsgContext{
		Properties: properties,
		Message: synctx.Message{
			ContentType: f.config.Parameters["transport.vfs.ContentType"],
		},
		Headers: headers,
	}

	// Read the file content
	content, err := f.protocolHandler.ReadFile(fileURI)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Set the payload if read was successful
	msgContext.Message.RawPayload = content

	// Check context before proceeding with mediation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:

		// Process the file through mediator
		if err := f.mediator.MediateInboundMessage(ctx, f.config.SequenceName, msgContext); err != nil {
			if err := f.handleFileAction(fileURI, "Failure"); err != nil {
				return fmt.Errorf("failed to handle file after failure: %w", err)
			}
		} else {
			if err := f.handleFileAction(fileURI, "Process"); err != nil {
				return fmt.Errorf("failed to handle file after process: %w", err)
			}
		}
		return nil
	}
}

// handleFileAction handles file operations based on the configured action (MOVE, DELETE)
func (f *FileInboundEndpoint) handleFileAction(fileURI, actionType string) error {
	slog.Info("handling file action")
	titleCaser := cases.Title(language.English)
	actionKey := fmt.Sprintf("transport.vfs.ActionAfter%s", titleCaser.String(actionType))
	if action, exists := f.config.Parameters[actionKey]; exists {
		slog.Info(action)
		if action == "MOVE" {
			movePathKey := fmt.Sprintf("transport.vfs.MoveAfter%s", titleCaser.String(actionType))
			movePath, exists := f.config.Parameters[movePathKey]
			if !exists || movePath == "" {
				return fmt.Errorf("move path not specified for %s action", actionType)
			}
			return f.protocolHandler.MoveFile(fileURI, movePath)
		}
	}
	// Default to DELETE as per specification
	return f.protocolHandler.DeleteFile(fileURI)
}

func (f *FileInboundEndpoint) validateConfig() error {
	// Check required interval parameter
	interval, exists := f.config.Parameters["interval"]
	if !exists || interval == "" {
		return fmt.Errorf("missing required parameter: 'interval'")
	}

	intervalMs, err := strconv.Atoi(interval)
	if err != nil {
		return fmt.Errorf("invalid interval value: must be an integer, got '%s'", interval)
	}
	if intervalMs <= 0 {
		return fmt.Errorf("invalid interval value: must be positive, got '%d'", intervalMs)
	}

	// Check sequential parameter (optional, default false)
	if val, exists := f.config.Parameters["sequential"]; exists {
		_, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid sequential value: must be true/false, got '%s'", val)
		}
	}

	// Check coordination parameter (optional, default false)
	if val, exists := f.config.Parameters["coordination"]; exists {
		_, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid coordination value: must be true/false, got '%s'", val)
		}
	}

	// Check required FileURI parameter
	fileURI, exists := f.config.Parameters["transport.vfs.FileURI"]
	if !exists || fileURI == "" {
		return fmt.Errorf("missing required parameter: 'transport.vfs.FileURI'")
	}

	// Check required ContentType parameter
	contentType, exists := f.config.Parameters["transport.vfs.ContentType"]
	if !exists || contentType == "" {
		return fmt.Errorf("missing required parameter: 'transport.vfs.ContentType'")
	}

	// Check action after process (optional, default DELETE)
	if val, exists := f.config.Parameters["transport.vfs.ActionAfterProcess"]; exists && val != "" {
		if val == "MOVE" {
			if _, exists := f.config.Parameters["transport.vfs.MoveAfterProcess"]; !exists || f.config.Parameters["transport.vfs.MoveAfterProcess"] == "" {
				return fmt.Errorf("missing required parameter: 'transport.vfs.MoveAfterProcess' is required when ActionAfterProcess is 'MOVE'")
			}
		}
	}

	if val, exists := f.config.Parameters["transport.vfs.ActionAfterFailure"]; exists && val != "" {
		if val == "MOVE" {
			if _, exists := f.config.Parameters["transport.vfs.MoveAfterFailure"]; !exists || f.config.Parameters["transport.vfs.MoveAfterFailure"] == "" {
				return fmt.Errorf("missing required parameter: 'transport.vfs.MoveAfterFailure' is required when ActionAfterFailure is 'MOVE'")
			}
		}
	}

	// Validate AutoLockReleaseInterval if provided
	if val, exists := f.config.Parameters["transport.vfs.AutoLockReleaseInterval"]; exists {
		timeout, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid AutoLockReleaseInterval value: must be an integer, got '%s'", val)
		}
		// Allow -1 for never timeout, or positive values
		if timeout != -1 && timeout <= 0 {
			return fmt.Errorf("invalid AutoLockReleaseInterval value: must be -1 or positive integer, got '%d'", timeout)
		}
	}

	return nil
}
