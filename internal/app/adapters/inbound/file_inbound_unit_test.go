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
	"os"
	"testing"
	"time"

	"github.com/apache/synapse-go/internal/app/core/domain"
	"github.com/apache/synapse-go/internal/pkg/core/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	

)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  domain.InboundConfig
		wantErr bool
	}{
		{
			name: "Valid configuration",
			config: domain.InboundConfig{
				Protocol: "file",
				Parameters: map[string]string{
					"interval":                         "1000",
					"transport.vfs.FileURI":            "file:///test",
					"transport.vfs.ContentType":        "text/plain",
					"transport.vfs.ActionAfterProcess": "DELETE",
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid protocol",
			config: domain.InboundConfig{
				Protocol: "http",
				Parameters: map[string]string{
					"interval":                  "1000",
					"transport.vfs.FileURI":     "file:///test",
					"transport.vfs.ContentType": "text/plain",
				},
			},
			wantErr: true,
		},
		{
			name: "Missing interval",
			config: domain.InboundConfig{
				Protocol: "file",
				Parameters: map[string]string{
					"transport.vfs.FileURI":     "file:///test",
					"transport.vfs.ContentType": "text/plain",
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid interval value",
			config: domain.InboundConfig{
				Protocol: "file",
				Parameters: map[string]string{
					"interval":                  "abc",
					"transport.vfs.FileURI":     "file:///test",
					"transport.vfs.ContentType": "text/plain",
				},
			},
			wantErr: true,
		},
		{
			name: "Missing file URI",
			config: domain.InboundConfig{
				Protocol: "file",
				Parameters: map[string]string{
					"interval":                  "1000",
					"transport.vfs.ContentType": "text/plain",
				},
			},
			wantErr: true,
		},
		{
			name: "Missing content type",
			config: domain.InboundConfig{
				Protocol: "file",
				Parameters: map[string]string{
					"interval":              "1000",
					"transport.vfs.FileURI": "file:///test",
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid sequential value",
			config: domain.InboundConfig{
				Protocol: "file",
				Parameters: map[string]string{
					"interval":                  "1000",
					"transport.vfs.FileURI":     "file:///test",
					"transport.vfs.ContentType": "text/plain",
					"sequential":                "notbool",
				},
			},
			wantErr: true,
		},
		{
			name: "MOVE action without move path",
			config: domain.InboundConfig{
				Protocol: "file",
				Parameters: map[string]string{
					"interval":                         "1000",
					"transport.vfs.FileURI":            "file:///test",
					"transport.vfs.ContentType":        "text/plain",
					"transport.vfs.ActionAfterProcess": "MOVE",
				},
			},
			wantErr: true,
		},
		{
			name: "Valid MOVE action with path",
			config: domain.InboundConfig{
				Protocol: "file",
				Parameters: map[string]string{
					"interval":                         "1000",
					"transport.vfs.FileURI":            "file:///test",
					"transport.vfs.ContentType":        "text/plain",
					"transport.vfs.ActionAfterProcess": "MOVE",
					"transport.vfs.MoveAfterProcess":   "file:///test/success",
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid lock timeout",
			config: domain.InboundConfig{
				Protocol: "file",
				Parameters: map[string]string{
					"interval":                              "1000",
					"transport.vfs.FileURI":                 "file:///test",
					"transport.vfs.ContentType":             "text/plain",
					"transport.vfs.AutoLockReleaseInterval": "abc",
				},
			},
			wantErr: true,
		},
		{
			name: "Negative lock timeout",
			config: domain.InboundConfig{
				Protocol: "file",
				Parameters: map[string]string{
					"interval":                              "1000",
					"transport.vfs.FileURI":                 "file:///test",
					"transport.vfs.ContentType":             "text/plain",
					"transport.vfs.AutoLockReleaseInterval": "-2",
				},
			},
			wantErr: true,
		},
		{
			name: "Valid -1 lock timeout",
			config: domain.InboundConfig{
				Protocol: "file",
				Parameters: map[string]string{
					"interval":                              "1000",
					"transport.vfs.FileURI":                 "file:///test",
					"transport.vfs.ContentType":             "text/plain",
					"transport.vfs.AutoLockReleaseInterval": "-1",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileInboundEndpoint{
				config: tt.config,
				logger: NewDefaultLogger(),
			}
			err := f.validateConfig()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

/*
func TestScanDirectoryWithPattern(t *testing.T) {
	tests := []struct {
		name         string
		folderURI    string
		pattern      string
		setupMocks   func(*MockFileSystem)
		expectedURIs []string
		wantErr      bool
	}{
		{
			name:      "Valid directory with no pattern",
			folderURI: "file:///test",
			pattern:   "",
			setupMocks: func(fs *MockFileSystem) {
				entries := []os.DirEntry{
					&MockDirEntry{name: "file1.txt", isDir: false},
					&MockDirEntry{name: "file2.log", isDir: false},
					&MockDirEntry{name: "subdir", isDir: true},
				}
				fs.On("ReadDir", "/test").Return(entries, nil)
				mockFile := &MockFile{}
				fs.On("OpenFile", "/test/file1.txt.lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(0644)).Return(mockFile, nil).Maybe()
				fs.On("Remove", "/test/file1.txt.lock").Return(nil).Maybe()
				fs.On("Open", "/test/file1.txt.lock").Return(mockFile, nil).Maybe()
				fs.On("Open", "/test/file1.txt").Return(mockFile, nil).Maybe()
				mockFile.On("Write", mock.Anything).Return(10, nil).Maybe()
				mockFile.On("Open", "/test/file1.txt").Return(mockFile, nil).Maybe()
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()
				
			},
			expectedURIs: []string{"file:/test/file1.txt", "file:/test/file2.log"},
			wantErr:      false,
		},
		{
			name:      "Valid directory with pattern",
			folderURI: "file:///test",
			pattern:   "*.txt",
			setupMocks: func(fs *MockFileSystem) {
				entries := []os.DirEntry{
					&MockDirEntry{name: "file1.txt", isDir: false},
					&MockDirEntry{name: "file2.log", isDir: false},
					&MockDirEntry{name: "subdir", isDir: true},
				}
				fs.On("ReadDir", "/test").Return(entries, nil)
				mockFile := &MockFile{}
				fs.On("OpenFile", "/test/file1.txt.lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(0644)).Return(mockFile, nil).Maybe()
				fs.On("Open", "/test/file1.txt.lock").Return(mockFile, nil).Maybe()
				fs.On("Open", "/test/file1.txt.lock").Return(mockFile, nil).Maybe()
				mockFile.On("Write", mock.Anything).Return(10, nil).Maybe()
				mockFile.On("Open", "/test/file1.txt").Return(mockFile, nil).Maybe()
				fs.On("Open", "/test/file1.txt").Return(mockFile, nil).Maybe()
				mockFile.On("Open", "/test/file1.txt").Return(mockFile, nil).Maybe()

				fs.On("Write", mock.Anything).Return(10, nil).Maybe()
			},
			expectedURIs: []string{"file:/test/file1.txt"},
			wantErr:      false,
		},
		{
			name:      "Invalid directory",
			folderURI: "file:///nonexistent",
			pattern:   "",
			setupMocks: func(fs *MockFileSystem) {
				fs.On("ReadDir", "/nonexistent").Return(nil, os.ErrNotExist)
				mockFile := &MockFile{}
				mockFile.On("Write", mock.Anything).Return(10, nil).Maybe()
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()
			},
			expectedURIs: nil,
			wantErr:      true,
		},
		{
			name:      "Empty directory",
			folderURI: "file:///empty",
			pattern:   "",
			setupMocks: func(fs *MockFileSystem) {
				fs.On("ReadDir", "/empty").Return([]os.DirEntry{}, nil)
				mockFile := &MockFile{}
				mockFile.On("Write", mock.Anything).Return(10, nil).Maybe()
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()
			},
			expectedURIs: []string{},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &MockFileSystem{}
			tt.setupMocks(fs)

			f := &FileInboundEndpoint{
				fs:     fs,
				logger: NewDefaultLogger(),
			}

			files, err := f.scanDirectoryWithPattern()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expectedURIs, files)
			}
			fs.AssertExpectations(t)
		})
	}
}
*/

func TestTryLockFile(t *testing.T) {
	tests := []struct {
		name       string
		filePath   string
		config     domain.InboundConfig
		setupMocks func(*MockFileSystem, *MockClock)
		expected   bool
		wantErr    bool
	}{
		{
			name:     "Successfully create lock",
			filePath: "/test/file.txt",
			config: domain.InboundConfig{
				Parameters: map[string]string{},
			},
			setupMocks: func(fs *MockFileSystem, clock *MockClock) {
				mockFile := &MockFile{}
				mockFile.On("Write", mock.Anything).Return(10, nil)
				mockFile.On("Close").Return(nil)
				fs.On("OpenFile", "/test/file.txt.lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(0644)).Return(mockFile, nil)
				mockFile.On("Write", mock.Anything).Return(10, nil).Maybe()
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()
			},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "Lock already exists - not stale",
			filePath: "/test/file.txt",
			config: domain.InboundConfig{
				Parameters: map[string]string{
					"transport.vfs.AutoLockReleaseInterval": "30000",
				},
			},
			setupMocks: func(fs *MockFileSystem, clock *MockClock) {
				// Lock exists
				fs.On("OpenFile", "/test/file.txt.lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(0644)).
					Return((*MockFile)(nil), os.ErrExist)

				// Not stale
				fileInfo := &MockFileInfo{
					modTime: time.Now().Add(-10 * time.Second),
				}
				fs.On("Stat", "/test/file.txt.lock").Return(fileInfo, nil)
				mockFile := &MockFile{}
				mockFile.On("Write", mock.Anything).Return(10, nil).Maybe()
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()

				// Current time for stale check
				clock.currentTime = time.Now()
			},
			expected: false,
			wantErr:  false,
		},
		{
			name:     "Lock already exists - stale",
			filePath: "/test/file.txt",
			config: domain.InboundConfig{
				Parameters: map[string]string{
					"transport.vfs.AutoLockReleaseInterval": "1000", // 1 second
				},
			},
			setupMocks: func(fs *MockFileSystem, clock *MockClock) {
				// First attempt - lock exists
				fs.On("OpenFile", "/test/file.txt.lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(0644)).
					Return((*MockFile)(nil), os.ErrExist).Once()

				// Stale check
				fileInfo := &MockFileInfo{
					modTime: time.Now().Add(-10 * time.Second), // 10 seconds old
				}
				fs.On("Stat", "/test/file.txt.lock").Return(fileInfo, nil)

				// Remove stale lock
				fs.On("Remove", "/test/file.txt.lock").Return(nil)

				// Second attempt - create lock
				mockFile := &MockFile{}
				mockFile.On("Write", mock.Anything).Return(10, nil)
				mockFile.On("Close").Return(nil)
				fs.On("OpenFile", "/test/file.txt.lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(0644)).
					Return(mockFile, nil).Once()
				mockFile.On("Write", mock.Anything).Return(10, nil).Maybe()
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()

				// Current time for stale check
				clock.currentTime = time.Now()
			},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "Error checking lock",
			filePath: "/test/file.txt",
			config: domain.InboundConfig{
				Parameters: map[string]string{},
			},
			setupMocks: func(fs *MockFileSystem, clock *MockClock) {
				fs.On("OpenFile", "/test/file.txt.lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(0644)).
					Return((*MockFile)(nil), os.ErrPermission)
				mockFile := &MockFile{}
				mockFile.On("Write", mock.Anything).Return(10, nil).Maybe()
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()
			},
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &MockFileSystem{}
			clock := &MockClock{}
			tt.setupMocks(fs, clock)

			f := &FileInboundEndpoint{
				config: tt.config,
				fs:     fs,
				clock:  clock,
				logger: NewDefaultLogger(),
			}

			locked, err := f.tryLockFile(tt.filePath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, locked)
			}
			fs.AssertExpectations(t)
			clock.AssertExpectations(t)
		})
	}
}

func TestProcessFile(t *testing.T) {
	tests := []struct {
		name       string
		file       string
		config     domain.InboundConfig
		setupMocks func(*MockFileSystem, *MockMediator, *MockLogger)
		wantErr    bool
	}{
		{
			name: "Successful file processing - DELETE",
			file: "file:///test/file.txt",
			config: domain.InboundConfig{
				SequenceName: "testSequence",
				Parameters: map[string]string{
					"transport.vfs.ContentType":        "text/plain",
					"transport.vfs.ActionAfterProcess": "DELETE",
				},
			},
			setupMocks: func(fs *MockFileSystem, mediator *MockMediator, logger *MockLogger) {

				// Try to lock file
				mockLockFile := &MockFile{}
				mockLockFile.On("Write", mock.Anything).Return(10, nil)
				mockLockFile.On("Close").Return(nil)
				fs.On("OpenFile", "/test/file.txt.lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(0644)).
					Return(mockLockFile, nil).Once()

				// Read file
				mockFile := createMockFileWithContent("test content")
				fs.On("Open", "/test/file.txt").Return(mockFile, nil).Once()

				// Mediate message
				mediator.On("MediateInboundMessage", mock.Anything, "testSequence", mock.Anything).Return(nil).Once()

				// Delete file after successful processing
				fs.On("Remove", "/test/file.txt").Return(nil).Once()

				// Release lock
				fs.On("Remove", "/test/file.txt.lock").Return(nil).Once()
				mockFile.On("Write", mock.Anything).Return(10, nil).Maybe()
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()

				// Logging
				logger.On("Info", mock.Anything, mock.Anything).Return().Maybe()
				logger.On("Debug", mock.Anything, mock.Anything).Return().Maybe()
			},
			wantErr: false,
		},

		{
			name: "Failed mediation",
			file: "file:///test/file.txt",
			config: domain.InboundConfig{
				SequenceName: "testSequence",
				Parameters: map[string]string{
					"transport.vfs.ContentType":        "text/plain",
					"transport.vfs.ActionAfterFailure": "DELETE",
				},
			},
			setupMocks: func(fs *MockFileSystem, mediator *MockMediator, logger *MockLogger) {
				// Try to lock file
				mockLockFile := &MockFile{}
				mockLockFile.On("Write", mock.Anything).Return(10, nil)
				mockLockFile.On("Close").Return(nil)
				fs.On("OpenFile", "/test/file.txt.lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(0644)).
					Return(mockLockFile, nil).Once()

				// Read file
				mockFile := createMockFileWithContent("test content")
				fs.On("Open", "/test/file.txt").Return(mockFile, nil).Once()

				// Mediation fails
				mediator.On("MediateInboundMessage", mock.Anything, "testSequence", mock.Anything).
					Return(errors.New("mediation error")).Once()

				// Delete file after failure
				fs.On("Remove", "/test/file.txt").Return(nil)

				// Release lock
				fs.On("Remove", "/test/file.txt.lock").Return(nil)
                mockFile.On("Write", mock.Anything).Return(10, nil).Maybe()
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()
				// Logging
				logger.On("Info", mock.Anything, mock.Anything).Return().Maybe()
				logger.On("Debug", mock.Anything, mock.Anything).Return().Maybe()
				logger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
			},
			wantErr: false, // No error returned because failure is handled internally
		},
		{
			name: "Failed file read",
			file: "file:///test/file.txt",
			config: domain.InboundConfig{
				SequenceName: "testSequence",
				Parameters: map[string]string{
					"transport.vfs.ContentType":        "text/plain",
					"transport.vfs.ActionAfterFailure": "DELETE",
				},
			},
			setupMocks: func(fs *MockFileSystem, mediator *MockMediator, logger *MockLogger) {
				// Try to lock file
				mockLockFile := &MockFile{}
				mockLockFile.On("Write", mock.Anything).Return(10, nil)
				mockLockFile.On("Close").Return(nil)
				fs.On("OpenFile", "/test/file.txt.lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(0644)).
					Return(mockLockFile, nil).Once()

				// Read file fails
				fs.On("Open", "/test/file.txt").Return((*MockFile)(nil), errors.New("read error")).Once()

				// Release lock
				fs.On("Remove", "/test/file.txt.lock").Return(nil).Once()
				mockLockFile.On("Write", mock.Anything).Return(10, nil).Maybe()
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()
				// Logging
				logger.On("Info", mock.Anything, mock.Anything).Return().Maybe()
				logger.On("Debug", mock.Anything, mock.Anything).Return().Maybe()
				logger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
			},
			wantErr: true,
		},

		{
			name: "Successful file processing - MOVE",
			file: "file:///test/file.txt",
			config: domain.InboundConfig{
				SequenceName: "testSequence",
				Parameters: map[string]string{
					"transport.vfs.ContentType":        "text/plain",
					"transport.vfs.ActionAfterProcess": "MOVE",
					"transport.vfs.MoveAfterProcess":   "file:///test/success",
				},
			},
			setupMocks: func(fs *MockFileSystem, mediator *MockMediator, logger *MockLogger) {
				// Convert URI to path
				fs.On("Stat", "/test/file.txt").Return(&MockFileInfo{
					name:    "file.txt",
					size:    10,
					modTime: time.Now(),
					mode:    0644,
				}, nil)

				// Try to lock file
				mockLockFile := &MockFile{}
				mockLockFile.On("Write", mock.Anything).Return(10, nil)
				mockLockFile.On("Close").Return(nil)
				fs.On("OpenFile", "/test/file.txt.lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(0644)).
					Return(mockLockFile, nil).Once()

				// Read file
				mockFile := createMockFileWithContent("test content")
				fs.On("Open", "/test/file.txt").Return(mockFile, nil).Once()

				// Mediate message
				mediator.On("MediateInboundMessage", mock.Anything, "testSequence", mock.Anything).Return(nil).Once()

				// For move operation, we need to mock the destination directory creation
				fs.On("MkdirAll", "/test/success", os.FileMode(0755)).Return(nil).Once()

				// Move file
				fs.On("Rename", "/test/file.txt", "/test/success/file.txt").Return(nil).Once()

				// Set permissions on destination
				fs.On("Chmod", "/test/success/file.txt", os.FileMode(0644)).Return(nil).Once()

				// Release lock
				fs.On("Remove", "/test/file.txt.lock").Return(nil).Once()
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()

				// Logging
				logger.On("Info", mock.Anything, mock.Anything).Return().Maybe()
				logger.On("Debug", mock.Anything, mock.Anything).Return().Maybe()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			fs := &MockFileSystem{}
			mediator := &MockMediator{}
			logger := &MockLogger{}
			tt.setupMocks(fs, mediator, logger)

			f := &FileInboundEndpoint{
				config:   tt.config,
				fs:       fs,
				logger:   logger,
				mediator: mediator,
			}

			// Create context
			ctx, cancel, wg := createTestContext()
			defer cancel()
			ctx = context.WithValue(ctx, utils.WaitGroupKey, wg)

			// Execute
			err := f.processFile(ctx, tt.file)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify expectations
			fs.AssertExpectations(t)
			mediator.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}

func TestHandleFileAction(t *testing.T) {
	tests := []struct {
		name       string
		file       string
		actionType string
		config     domain.InboundConfig
		setupMocks func(*MockFileSystem, *MockLogger)
		wantErr    bool
	}{
		{
			name:       "MOVE action with path",
			file:       "file:///test/file.txt",
			actionType: "Process",
			config: domain.InboundConfig{
				Parameters: map[string]string{
					"transport.vfs.ActionAfterProcess": "MOVE",
					"transport.vfs.MoveAfterProcess":   "file:///test/success",
				},
			},
			setupMocks: func(fs *MockFileSystem, logger *MockLogger) {
				// Path conversion
				fs.On("Stat", "/test/file.txt").Return(&MockFileInfo{
					name:    "file.txt",
					size:    10,
					modTime: time.Now(),
					mode:    0644,
				}, nil)

				// Move operation
				fs.On("Rename", "/test/file.txt", "/test/success/file.txt").Return(nil)

				// For move operation, we need to mock the destination directory creation
				fs.On("MkdirAll", "/test/success", os.FileMode(0755)).Return(nil).Once()

				// Set permissions on destination
				fs.On("Chmod", "/test/success/file.txt", os.FileMode(0644)).Return(nil).Once()
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()

				// Logging
				logger.On("Info", mock.Anything, mock.Anything).Return()
			},
			wantErr: false,
		},
		{
			name:       "DELETE action",
			file:       "file:///test/file.txt",
			actionType: "Process",
			config: domain.InboundConfig{
				Parameters: map[string]string{
					"transport.vfs.ActionAfterProcess": "DELETE",
				},
			},
			setupMocks: func(fs *MockFileSystem, logger *MockLogger) {
				// Delete operation
				fs.On("Remove", "/test/file.txt").Return(nil)
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()

				// Logging
				logger.On("Info", mock.Anything, mock.Anything).Return()
			},
			wantErr: false,
		},
		{
			name:       "Default to DELETE if no action specified",
			file:       "file:///test/file.txt",
			actionType: "Process",
			config: domain.InboundConfig{
				Parameters: map[string]string{},
			},
			setupMocks: func(fs *MockFileSystem, logger *MockLogger) {
				// Delete operation
				fs.On("Remove", "/test/file.txt").Return(nil)
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()

				// Logging
				logger.On("Info", mock.Anything, mock.Anything).Return()
			},
			wantErr: false,
		},
		{
			name:       "MOVE without path",
			file:       "file:///test/file.txt",
			actionType: "Process",
			config: domain.InboundConfig{
				Parameters: map[string]string{
					"transport.vfs.ActionAfterProcess": "MOVE",
				},
			},
			setupMocks: func(fs *MockFileSystem, logger *MockLogger) {
				// Logging
				logger.On("Info", mock.Anything, mock.Anything).Return()
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			fs := &MockFileSystem{}
			logger := &MockLogger{}
			tt.setupMocks(fs, logger)

			f := &FileInboundEndpoint{
				config: tt.config,
				fs:     fs,
				logger: logger,
			}

			// Execute
			err := f.handleFileAction(tt.file, tt.actionType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify expectations
			fs.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}

func TestConvertFileURIToPath(t *testing.T) {
	tests := []struct {
		name     string
		fileURI  string
		expected string
		wantErr  bool
	}{
		{
			name:     "Valid file URI",
			fileURI:  "file:///test/file.txt",
			expected: "/test/file.txt",
			wantErr:  false,
		},
		{
			name:     "Valid file URI with spaces",
			fileURI:  "file:///test/file%20with%20spaces.txt",
			expected: "/test/file with spaces.txt",
			wantErr:  false,
		},
		{
			name:     "Invalid scheme",
			fileURI:  "http://example.com/file.txt",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Invalid URI format",
			fileURI:  "not-a-uri",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileInboundEndpoint{}
			result, err := f.convertFileURIToPath(tt.fileURI)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestIsSequentialProcessing(t *testing.T) {
	tests := []struct {
		name     string
		config   domain.InboundConfig
		expected bool
	}{
		{
			name: "Sequential true",
			config: domain.InboundConfig{
				Parameters: map[string]string{
					"sequential": "true",
				},
			},
			expected: true,
		},
		{
			name: "Sequential false",
			config: domain.InboundConfig{
				Parameters: map[string]string{
					"sequential": "false",
				},
			},
			expected: false,
		},
		{
			name: "Sequential not set",
			config: domain.InboundConfig{
				Parameters: map[string]string{},
			},
			expected: false,
		},
		{
			name: "Sequential invalid value",
			config: domain.InboundConfig{
				Parameters: map[string]string{
					"sequential": "invalid",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileInboundEndpoint{
				config: tt.config,
				logger: NewDefaultLogger(),
			}
			result := f.isSequentialProcessing()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessFiles(t *testing.T) {
	tests := []struct {
		name       string
		files      []string
		config     domain.InboundConfig
		sequential bool
		setupMocks func(*MockFileSystem, *MockMediator, *MockLogger)
		wantErr    bool
	}{
		{
			name:  "Process multiple files sequentially",
			files: []string{"file:///test/file1.txt", "file:///test/file2.txt"},
			config: domain.InboundConfig{
				SequenceName: "testSequence",
				Parameters: map[string]string{
					"sequential":                "true",
					"transport.vfs.ContentType": "text/plain",
				},
			},
			sequential: true,
			setupMocks: func(fs *MockFileSystem, mediator *MockMediator, logger *MockLogger) {
				// For both files, setup mock behaviors
				for _, file := range []string{"file1.txt", "file2.txt"} {
					filePath := "/test/" + file

					// Lock acquisition
					mockLockFile := &MockFile{}
					mockLockFile.On("Write", mock.Anything).Return(10, nil)
					mockLockFile.On("Close").Return(nil)
					fs.On("OpenFile", filePath+".lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(0644)).
						Return(mockLockFile, nil)

					// File reading
					mockFile := createMockFileWithContent("content of " + file)
					fs.On("Open", filePath).Return(mockFile, nil)
					mockFile.On("Stat").Return(&MockFileInfo{
						name:    file,
						size:    int64(len("content of " + file)),
						modTime: time.Now(),
					}, nil)

					// Mediation
					mediator.On("MediateInboundMessage", mock.Anything, "testSequence", mock.Anything).Return(nil)

					// File handling (DELETE is default)
					fs.On("Remove", filePath).Return(nil)
					fs.On("Write", mock.Anything).Return(10, nil).Maybe()

					// Lock release
					fs.On("Remove", filePath+".lock").Return(nil)
					mockFile.On("Write", mock.Anything).Return(10, nil).Maybe()
				}

				// Logging
				logger.On("Info", mock.Anything, mock.Anything).Return()
			},
			wantErr: false,
		},
		{
			name:  "Skip file already being processed",
			files: []string{"file:///test/file1.txt", "file:///test/file1.txt"},
			config: domain.InboundConfig{
				SequenceName: "testSequence",
				Parameters: map[string]string{
					"transport.vfs.ContentType": "text/plain",
				},
			},
			sequential: false,
			setupMocks: func(fs *MockFileSystem, mediator *MockMediator, logger *MockLogger) {

				// Lock acquisition for first instance
				mockLockFile := &MockFile{}
				mockLockFile.On("Write", mock.Anything).Return(10, nil)
				mockLockFile.On("Close").Return(nil)

				// File reading
				mockFile := createMockFileWithContent("content of file1.txt")
				// fs.On("Open", filePath).Return(mockFile, nil)
				mockFile.On("Stat").Return(&MockFileInfo{
					name:    "file1.txt",
					size:    int64(len("content of file1.txt")),
					modTime: time.Now(),
				}, nil)
				fs.On("OpenFile", "/test/file1.txt.lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(0644)).Return(mockFile, nil).Maybe()
				fs.On("Remove", "/test/file1.txt.lock").Return(nil).Maybe()
				fs.On("Open", "/test/file1.txt.lock").Return(mockFile, nil).Maybe()
				logger.On("Debug", mock.Anything, mock.Anything).Return()
				mockLockFile.On("Write", mock.Anything).Return(10, nil).Maybe()
				mockFile.On("Open","/test/file1.txt.lock").Return(mockLockFile, nil).Maybe()
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()
			},
			wantErr: false,
		},
		{
			name:  "Empty file list",
			files: []string{},
			config: domain.InboundConfig{
				Parameters: map[string]string{
					"transport.vfs.ContentType": "text/plain",
				},
			},
			sequential: false,
			setupMocks: func(fs *MockFileSystem, mediator *MockMediator, logger *MockLogger) {
				// No operations expected
				mockLockFile := &MockFile{}
				mockLockFile.On("Write", mock.Anything).Return(10, nil).Maybe()
				fs.On("Write", mock.Anything).Return(10, nil).Maybe()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			fs := &MockFileSystem{}
			mediator := &MockMediator{}
			logger := &MockLogger{}
			tt.setupMocks(fs, mediator, logger)

			// Create endpoint with mocks
			f := &FileInboundEndpoint{
				config:   tt.config,
				fs:       fs,
				logger:   logger,
				mediator: mediator,
			}

			// Create context
			ctx, cancel, wg := createTestContext()
			defer cancel()
			ctx = context.WithValue(ctx, utils.WaitGroupKey, wg)

			// Execute
			err := f.processFiles(ctx, tt.files)

			// Wait for background tasks to complete
			cancel()

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify
			fs.AssertExpectations(t)
			mediator.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}


