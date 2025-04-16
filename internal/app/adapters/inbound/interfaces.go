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
	"io"
	"os"
	"time"
)

type FileSystem interface {
	Open(name string) (File, error)
	ReadDir(dirname string) ([]os.DirEntry, error)
	Rename(oldpath, newpath string) error
	Remove(name string) error
	Stat(name string) (os.FileInfo, error)
	Create(name string) (File, error)
	OpenFile(name string, flag int, perm os.FileMode) (File, error)
	MkdirAll(path string, perm os.FileMode) error
	Chmod(name string, mode os.FileMode) error
}

type File interface {
	io.ReadCloser
	io.WriteCloser
	Stat() (os.FileInfo, error)
	Name() string
	Fd() uintptr
}

type Clock interface {
	Now() time.Time
	Sleep(d time.Duration)
	NewTicker(d time.Duration) *time.Ticker
}

type Logger interface {
	Info(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
	Debug(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
}

type Field struct {
	Key   string
	Value interface{}
}
