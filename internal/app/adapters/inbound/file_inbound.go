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

package inbound

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/apache/synapse-go/internal/app/core/domain"
	"github.com/apache/synapse-go/internal/app/core/ports"
	"github.com/apache/synapse-go/internal/pkg/core/synctx"
)

// FileInboundEndpoint handles file-based inbound operations
type FileInboundEndpoint struct {
	config          domain.InboundConfig
	isRunning       bool
	fs              FileSystem
	clock           Clock
	logger          Logger
	mediator        ports.InboundMessageMediator
	processingFiles sync.Map
}

// NewFileInboundEndpoint creates a new FileInboundEndpoint instance
func NewFileInboundEndpoint(
	config domain.InboundConfig,
	fs FileSystem,
	clock Clock,
	logger Logger,
	mediator ports.InboundMessageMediator,
) *FileInboundEndpoint {
	return &FileInboundEndpoint{
		config:   config,
		fs:       fs,
		clock:    clock,
		logger:   logger,
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
		f.logger.Error("invalid configuration", err)
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	f.mediator = mediator
	f.logger.Info("starting file inbound endpoint")
	
	// Start polling
	err := f.poll(ctx)

	// When context is cancelled, wait for all processing to complete
	f.logger.Info("waiting for in-progress file operations to complete")

	return err
}

// call this using a channel
func (adapter *FileInboundEndpoint) Stop() error {
	adapter.isRunning = false
	return nil
}

func (f *FileInboundEndpoint) poll(ctx context.Context) error {

	interval, _ := strconv.Atoi(f.config.Parameters["interval"])
	ticker := f.clock.NewTicker(time.Duration(interval) * time.Millisecond)
	defer ticker.Stop()

	// // Create a done channel to signal when all processing is complete
	// done := make(chan struct{})

	// // Start a goroutine to handle the polling
	// go func() {
	// 	defer close(done)

	processingWg := &sync.WaitGroup{}

	for {
		select {
		case <-ctx.Done():
			f.logger.Info("received shutdown signal, stopping file polling")
			// Wait for all processing to complete before returning
			processingWg.Wait()
			return ctx.Err()

		case <-ticker.C:
			processingWg.Add(1)
			go func() {
				defer processingWg.Done()
				if err := f.processingCycle(ctx); err != nil {
					if err != context.Canceled {
						f.logger.Error("error in processing cycle", err)
					}
				}
			}()
		}
	}
	// }()

	// // Wait for either context cancellation or done signal
	// select {
	// case <-ctx.Done():
	// 	// Context was canceled, wait for processing to complete
	// 	<-done
	// 	return ctx.Err()
	// case <-done:
	// 	// Processing completed on its own
	// 	return nil
	// }
}

func (f *FileInboundEndpoint) processingCycle(ctx context.Context) error {
	files, err := f.scanDirectoryWithPattern()
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	return f.processFiles(ctx, files)
}

// Process a given list of files sequentially or parallely.
// Calls ProcessFile function for each file in the list.
func (f *FileInboundEndpoint) processFiles(ctx context.Context, files []string) error {
	isSequential := f.isSequentialProcessing()
	// wg := ctx.Value(utils.WaitGroupKey).(*sync.WaitGroup)
	fileWg := &sync.WaitGroup{}

	for _, file := range files {
		select {
		case <-ctx.Done():
			f.logger.Info("cancelling remaining file processing")
			// Wait for all processing to complete before returning
			fileWg.Wait()
			return ctx.Err()
		default:
			// Check if file is already being processed
			if _, exists := f.processingFiles.LoadOrStore(file, true); exists {
				f.logger.Debug("skipping file - already being processed",
					Field{"file", file})
				continue
			}

			if isSequential {
				if err := f.processFile(ctx, file); err != nil {
					f.logger.Error("failed to process file", err)
				}
			} else {
				fileWg.Add(1)
				go func(fileName string) {
					defer fileWg.Done()
					if err := f.processFile(ctx, fileName); err != nil {
						f.logger.Error("failed to process file", err)
					}
				}(file)
			}
		}
	}
	return nil
}

// Check status of sequential service level parameter
func (f *FileInboundEndpoint) isSequentialProcessing() bool {
	if val, exists := f.config.Parameters["sequential"]; exists {
		sequential, err := strconv.ParseBool(val)
		if err != nil {
			f.logger.Warn("invalid sequential value, defaulting to false",
				Field{"value", val})
			return false
		}
		return sequential
	}
	return false
}

// tryLockFile attempts to create a lock file and returns true if successful
func (f *FileInboundEndpoint) tryLockFile(filePath string) (bool, error) {
	lockFile := filePath + ".lock"

	// Get auto lock release interval from config (default 20000ms)
	lockTimeout := 20000
	if val, exists := f.config.Parameters["transport.vfs.AutoLockReleaseInterval"]; exists {
		timeout, err := strconv.Atoi(val)
		if err == nil {
			lockTimeout = timeout
		}
	}

	// Attempt to create lock file
	file, err := f.fs.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			// Lock file exists, check if it's stale
			lockInfo, err := f.fs.Stat(lockFile)
			if err != nil {
				return false, fmt.Errorf("failed to check lock file: %w", err)
			}

			// Check if lock is stale based on AutoLockReleaseInterval
			if lockTimeout != -1 && time.Since(lockInfo.ModTime()) > time.Duration(lockTimeout)*time.Millisecond {
				// Remove stale lock
				if err := f.fs.Remove(lockFile); err != nil {
					return false, fmt.Errorf("failed to remove stale lock: %w", err)
				}
				// Try again
				return f.tryLockFile(filePath)
			}
			return false, nil // File is locked by another process
		}
		return false, fmt.Errorf("failed to create lock file: %w", err)
	}

	// Write process information to lock file
	hostname, _ := os.Hostname()
	lockInfo := fmt.Sprintf("pid:%d\nhost:%s\ntime:%s\n", os.Getpid(), hostname, time.Now().Format(time.RFC3339))
	if _, err := file.Write([]byte(lockInfo)); err != nil {
		file.Close()
		f.fs.Remove(lockFile)
		return false, fmt.Errorf("failed to write lock info: %w", err)
	}

	file.Close()
	return true, nil
}

// releaseLock removes the lock file
func (f *FileInboundEndpoint) releaseLock(filePath string) error {
	return f.fs.Remove(filePath + ".lock")
}

// processFile with distributed locking
func (f *FileInboundEndpoint) processFile(ctx context.Context, file string) error {
	defer f.processingFiles.Delete(file)

	// Convert URI to path for locking
	filePath, err := f.convertFileURIToPath(file)
	if err != nil {
		return fmt.Errorf("failed to convert file path: %w", err)
	}

	// Try to acquire lock
	locked, err := f.tryLockFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !locked {
		return nil // File is locked, skip processing
	}
	defer f.releaseLock(filePath)

	msgContext, err := f.readFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Check context before proceeding with mediation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:

		// Process the file through mediator
		if err := f.mediator.MediateInboundMessage(ctx, f.config.SequenceName, msgContext); err != nil {
			if err := f.handleFileAction(file, "Failure"); err != nil {
				return fmt.Errorf("failed to handle file after failure: %w", err)
			}
		}else{
			if err := f.handleFileAction(file, "Process"); err != nil {
				return fmt.Errorf("failed to handle file after process: %w", err)
			}
		}
		return nil
	}
}

// handleFileAction handles file operations based on the configured action (MOVE, DELETE)
func (f *FileInboundEndpoint) handleFileAction(file, actionType string) error {
	f.logger.Info("handling file action")
	actionKey := fmt.Sprintf("transport.vfs.ActionAfter%s", strings.Title(actionType))
	if action, exists := f.config.Parameters[actionKey]; exists {
		f.logger.Info(action)
		if action == "MOVE" {
			movePathKey := fmt.Sprintf("transport.vfs.MoveAfter%s", strings.Title(actionType))
			movePath, exists := f.config.Parameters[movePathKey]
			if !exists || movePath == "" {
				return fmt.Errorf("move path not specified for %s action", actionType)
			}
			return f.moveFileAfterAction(file, movePath)
		}
	}
	// Default to DELETE as per specification
	return f.deleteFile(file)
}

func (f *FileInboundEndpoint) moveFileAfterAction(file, destPath string) error {
	sourcePath, err := f.convertFileURIToPath(file)
	if err != nil {
		return fmt.Errorf("failed to convert source path: %w", err)
	}

	destURI, err := f.convertFileURIToPath(destPath)
	if err != nil {
		return fmt.Errorf("failed to convert destination path: %w", err)
	}

	// Create full destination path including filename
	fileName := filepath.Base(sourcePath)
	fullDestPath := filepath.Join(destURI, fileName)

	if err := f.moveFile(sourcePath, fullDestPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	f.logger.Info("file moved successfully",
		Field{"from", sourcePath},
		Field{"to", fullDestPath})
	return nil
}

func (f *FileInboundEndpoint) deleteFile(file string) error {
	filePath, err := f.convertFileURIToPath(file)
	if err != nil {
		return fmt.Errorf("failed to convert file path: %w", err)
	}

	if err := f.fs.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	f.logger.Info("file deleted successfully",
		Field{"file", filePath})
	return nil
}

// Helper methods for file operations
func (f *FileInboundEndpoint) readFile(fileURI string) (*synctx.MsgContext, error) {
	filePath, err := f.convertFileURIToPath(fileURI)
	if err != nil {
		return nil, fmt.Errorf("failed to convert file URI to path: %w", err)
	}

	file, err := f.fs.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Set up headers with file metadata
	headers := map[string]string{
		"FILE_LENGTH":   fmt.Sprintf("%d", info.Size()),
		"LAST_MODIFIED": fmt.Sprintf("%d", info.ModTime().Unix()),
		"FILE_URI":      fileURI,
		"FILE_PATH":     filePath,
		"FILE_NAME":     info.Name(),
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

	// Try to read the file content
	content, err := io.ReadAll(file)
	if err != nil {
		// Return the message context with metadata but nil payload
		return msgContext, fmt.Errorf("failed to read file content: %w", err)
	}

	// Set the payload if read was successful
	msgContext.Message.RawPayload = content
	return msgContext, nil
}

func (f *FileInboundEndpoint) validateConfig() error {
	// Check protocol
	if f.config.Protocol != "file" {
		return fmt.Errorf("invalid protocol: expected 'file', got '%s'", f.config.Protocol)
	}

	// Check required interval parameter
	interval, exists := f.config.Parameters["interval"]
	if !exists || interval == "" {
		return fmt.Errorf("missing required parameter: 'interval'")
	}
	if _, err := strconv.Atoi(interval); err != nil {
		return fmt.Errorf("invalid interval value: must be a valid integer, got '%s'", interval)
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
		return fmt.Errorf("missing required paramter: 'transport.vfs.ContentType'")
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

func (f *FileInboundEndpoint) scanDirectoryWithPattern() ([]string, error) {
	folderURI := f.config.Parameters["transport.vfs.FileURI"]
	pattern, exists := f.config.Parameters["transport.vfs.FileNamePattern"]
	if !exists {
		pattern = ""
	}
	folderPath, err := f.convertFileURIToPath(folderURI)
	if err != nil {
		return nil, fmt.Errorf("failed to convert URI to path: %w", err)
	}

	entries, err := f.fs.ReadDir(folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories
		}

		// If pattern is empty, include all files
		// If pattern is set, only include matching files
		if pattern == "" || matchesPattern(entry.Name(), pattern) {
			files = append(files, filepath.Join(folderURI, entry.Name()))
		}
	}
	return files, nil
}

// Helper function to match file patterns
func matchesPattern(filename, pattern string) bool {
	matched, err := filepath.Match(pattern, filename)
	if err != nil {
		// If pattern is invalid, treat it as no match
		return false
	}
	return matched
}

func (f *FileInboundEndpoint) moveFile(sourcePath, destPath string) error {
	destDir := filepath.Dir(destPath)
	if err := f.fs.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	srcInfo, err := f.fs.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	err = f.fs.Rename(sourcePath, destPath)
	if err != nil {
		var linkErr *os.LinkError
		if errors.As(err, &linkErr) && strings.Contains(linkErr.Error(), "invalid cross-device link") {
			// Cross-filesystem move, use copy and delete
			sourceFile, err := f.fs.Open(sourcePath)
			if err != nil {
				return fmt.Errorf("failed to open source file for copy: %w", err)
			}
			defer sourceFile.Close()

			destFile, err := f.fs.Create(destPath)
			if err != nil {
				return fmt.Errorf("failed to create destination file for copy: %w", err)
			}
			defer destFile.Close()

			_, err = io.Copy(destFile, sourceFile)
			if err != nil {
				return fmt.Errorf("failed to copy file: %w", err)
			}

			if err := f.fs.Remove(sourcePath); err != nil {
				// Log the error, but don't fail the move
				fmt.Printf("warning: failed to remove source file after copy: %v\n", err)
			}
		} else {
			return fmt.Errorf("failed to move file: %w", err)
		}
	}

	if err := f.fs.Chmod(destPath, srcInfo.Mode()); err != nil {
		fmt.Printf("warning: failed to set permissions: %v\n", err)
	}

	return nil
}

func (f *FileInboundEndpoint) convertFileURIToPath(fileURI string) (string, error) {
	parsedURI, err := url.Parse(fileURI)
	if err != nil {
		return "", fmt.Errorf("invalid file URI: %w", err)
	}

	if parsedURI.Scheme != "file" {
		return "", fmt.Errorf("unsupported URI scheme: %s", parsedURI.Scheme)
	}

	filePath := parsedURI.Path
	filePath = filepath.Clean(filePath)
	filePath = strings.ReplaceAll(filePath, "%20", " ")

	return filePath, nil
}
