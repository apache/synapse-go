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
	"io"
	"os"
	"sync"
	"time"

	"github.com/apache/synapse-go/internal/pkg/core/synctx"
	"github.com/apache/synapse-go/internal/pkg/core/utils"
	"github.com/stretchr/testify/mock"
)

// MockFileSystem mocks the FileSystem interface
type MockFileSystem struct {
	mock.Mock
}

func (m *MockFileSystem) Open(name string) (File, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(File), args.Error(1)
}

func (m *MockFileSystem) ReadDir(dirname string) ([]os.DirEntry, error) {
	args := m.Called(dirname)
	if err := args.Error(1); err != nil {
		return nil, err // Return nil slice and error
	}
	return args.Get(0).([]os.DirEntry), nil // Return slice and nil error.
}

func (m *MockFileSystem) Rename(oldpath, newpath string) error {
	args := m.Called(oldpath, newpath)
	return args.Error(0)
}

func (m *MockFileSystem) Remove(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(os.FileInfo), args.Error(1)
}

func (m *MockFileSystem) Create(name string) (File, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(File), args.Error(1)
}

func (m *MockFileSystem) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	args := m.Called(name, flag, perm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(File), args.Error(1)
}

func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}

func (m *MockFileSystem) Chmod(name string, mode os.FileMode) error {
	args := m.Called(name, mode)
	return args.Error(0)
}

// MockFile mocks the File interface
type MockFile struct {
	mock.Mock
	contents []byte
	position int
	ReadFunc func(p []byte) (int, error)
}

func (m *MockFile) Read(p []byte) (n int, err error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(p)
	}

	if m.position >= len(m.contents) {
		return 0, io.EOF
	}

	n = copy(p, m.contents[m.position:])
	m.position += n
	return n, nil
}

func (m *MockFile) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockFile) Write(p []byte) (n int, err error) {
	if len(m.Mock.ExpectedCalls) == 0 {
		return len(p), nil
	}
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockFile) Stat() (os.FileInfo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(os.FileInfo), args.Error(1)
}

func (m *MockFile) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockFile) Fd() uintptr {
	args := m.Called()
	return uintptr(args.Int(0))
}

// MockClock mocks the Clock interface
type MockClock struct {
	mock.Mock
	currentTime time.Time
}

func (m *MockClock) Now() time.Time {
	if !m.currentTime.IsZero() {
		return m.currentTime
	}
	args := m.Called()
	return args.Get(0).(time.Time)
}

func (m *MockClock) Sleep(d time.Duration) {
	m.Called(d)
}

func (m *MockClock) NewTicker(d time.Duration) *time.Ticker {
	args := m.Called(d)
	return args.Get(0).(*time.Ticker)
}

// MockLogger mocks the Logger interface
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, fields ...Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, err error, fields ...Field) {
	m.Called(msg, err, fields)
}

func (m *MockLogger) Debug(msg string, fields ...Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...Field) {
	m.Called(msg, fields)
}

// MockMediator mocks the InboundMessageMediator interface
type MockMediator struct {
	mock.Mock
}

func (m *MockMediator) MediateInboundMessage(ctx context.Context, seqName string, msg *synctx.MsgContext) error {
	args := m.Called(ctx, seqName, msg)
	return args.Error(0)
}

// MockFileInfo mocks os.FileInfo
type MockFileInfo struct {
	mock.Mock
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
	sys     interface{}
}

func (m *MockFileInfo) Name() string {
	return m.name
}

func (m *MockFileInfo) Size() int64 {
	return m.size
}

func (m *MockFileInfo) Mode() os.FileMode {
	return m.mode
}

func (m *MockFileInfo) ModTime() time.Time {
	return m.modTime
}

func (m *MockFileInfo) IsDir() bool {
	return m.isDir
}

func (m *MockFileInfo) Sys() interface{} {
	return m.sys
}

// MockDirEntry mocks os.DirEntry
type MockDirEntry struct {
	mock.Mock
	name  string
	isDir bool
}

func (m *MockDirEntry) Name() string {
	return m.name
}

func (m *MockDirEntry) IsDir() bool {
	return m.isDir
}

func (m *MockDirEntry) Type() os.FileMode {
	args := m.Called()
	return args.Get(0).(os.FileMode)
}

func (m *MockDirEntry) Info() (os.FileInfo, error) {
	args := m.Called()
	return args.Get(0).(os.FileInfo), args.Error(1)
}

// TestHelper for creating context with waitgroup
func createTestContext() (context.Context, context.CancelFunc, *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	ctx = context.WithValue(ctx, utils.WaitGroupKey, wg)
	return ctx, cancel, wg
}

// Creates string reader for file content
func createMockFileWithContent(content string) *MockFile {
	mockFile := &MockFile{contents: []byte(content)}
	info := &MockFileInfo{
		name:    "test.txt",
		size:    int64(len(content)),
		mode:    0644,
		modTime: time.Now(),
		isDir:   false,
	}
	mockFile.On("Stat").Return(info, nil)
	mockFile.On("Close").Return(nil)
	mockFile.On("Name").Return("test.txt")
	mockFile.On("Write", mock.Anything).Return(len(content), nil).Maybe()
	return mockFile
}
