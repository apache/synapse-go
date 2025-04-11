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
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/apache/synapse-go/internal/app/core/domain"
	"github.com/c2fo/vfs/v7"
	"github.com/c2fo/vfs/v7/backend"
	"github.com/c2fo/vfs/v7/backend/ftp"
	"github.com/c2fo/vfs/v7/backend/sftp"
	"github.com/c2fo/vfs/v7/options"
	"github.com/c2fo/vfs/v7/vfssimple"
)

var (
	// ErrMissingAuthority is return when authority is missing from a URI
	ErrMissingAuthority = errors.New("unable to determine uri authority ([user@]host[:port]) for network-based scheme")
	// ErrMissingScheme is returned when Scheme is missing from a URI
	ErrMissingScheme = errors.New("unable to determine uri scheme")
	// ErrRegFsNotFound is returned when no registered backend matches the URI
	ErrRegFsNotFound = errors.New("no matching registered filesystem found")
	// ErrBlankURI is returned when the URI is blank
	ErrBlankURI = errors.New("uri is blank")
)

//implements protocol handler interface
type VFSProtocolHandler struct {
	fsLocation vfs.Location
}

type VFSProtocolHandlerFactory struct {}

func (h *VFSProtocolHandlerFactory) CreateHandler(config domain.InboundConfig) (*VFSProtocolHandler, error) {
	uri, exists :=config.Parameters["transport.vfs.FileURI"]
	if !exists {
		return nil, fmt.Errorf("missing transport.vfs.FileURI parameter")
	}
	// Check and add trailing slash if needed
	if !strings.HasSuffix(uri, "/") {
		uri += "/"
	}
	var longest string
	backends := backend.RegisteredBackends()
	for _, backendName := range backends {
		if strings.HasPrefix(uri, backendName) {
			// The first match always becomes the longest
			if longest == "" {
				longest = backendName
				continue
			}

			// we found a longer (more specific) backend prefix matching URI
			if len(backendName) > len(longest) {
				longest = backendName
			}
		}
	}
	if longest == "" {
		return nil, ErrRegFsNotFound
	}

	var fsLocation vfs.Location

	if longest == "file" {
		fsLocation, err := vfssimple.NewLocation(uri)	
		if err != nil {
			return nil, fmt.Errorf("failed to create VFS location: %w", err)
		}
		return &VFSProtocolHandler{
			fsLocation: fsLocation,
		}, nil
	}

	if longest == "ftp" || longest == "sftp" || longest == "ftps" {
		u, err := url.Parse(uri)
        if err != nil {
			return nil, fmt.Errorf("failed to parse URI: %w", err)
        }

        userInfo := u.User
        var username, password string
        if userInfo != nil {
                username = userInfo.Username()
                password, _ = userInfo.Password()
        }

        query := u.Query()
        queryParams := make(map[string]string)
        for key, values := range query {
                if len(values) > 0 {
                        queryParams[key] = values[0]
                }
        }

        // Construct the new URI
        u.User = url.User(username)
        u.RawQuery = "" // Remove query parameters

		// Convert ftp://username:password@host:port/path to ftps://username@host:port/path
        convertedURI := u.String()

		if longest == "sftp" {
			fsLocation, err = vfssimple.NewLocation(convertedURI)	
			if err != nil {
				return nil, fmt.Errorf("failed to create VFS location: %w", err)
			}
			sftp_fs,ok := fsLocation.FileSystem().(*sftp.FileSystem)
			if !ok {
				return nil, fmt.Errorf("failed to cast to sftp.FileSystem")
			}         
			sftp_options := sftp.Options{
				Password: password,
			}
			options.ApplyOptions(sftp_fs, sftp.WithOptions(sftp_options))
		} else {
			fsLocation, err = vfssimple.NewLocation(convertedURI)
			if err != nil {
				return nil, fmt.Errorf("failed to create VFS location: %w", err)
			}
			ftp_fs,ok := fsLocation.FileSystem().(*ftp.FileSystem)
			if !ok {
				return nil, fmt.Errorf("failed to cast to ftp.FileSystem")
			}
			ftp_options := ftp.Options{
				Password: password,
			}
			options.ApplyOptions(ftp_fs, ftp.WithOptions(ftp_options))
		}

	}

	return &VFSProtocolHandler{
		fsLocation: fsLocation,
	}, nil
}

func (v *VFSProtocolHandler) ListFiles(pattern string) ([]string, error) {
	exists, err := v.fsLocation.Exists()
	if err != nil  {
		return nil, fmt.Errorf("failed to check if location exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("location does not exist")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex pattern: %w", err)
	}
	files, err := v.fsLocation.ListByRegex(re)
	if err != nil {
		return nil, fmt.Errorf("failed to list files by regex: %w", err)
	}
	return files, nil
}

func (v *VFSProtocolHandler) ReadFile(filePath string) ([]byte, error) {
	file, err := v.fsLocation.NewFile(filePath)
	if err != nil {
			return nil, fmt.Errorf("failed to create VFS file: %w", err)
	}
	defer file.Close()

	result, err := io.ReadAll(file)
	if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return result, nil
}

func (v *VFSProtocolHandler) GetLastModified(fileURI string) (*time.Time, error) {
	file, err := v.fsLocation.NewFile(fileURI)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	return file.LastModified()
}

func (v *VFSProtocolHandler) GetSize(fileURI string) (uint64, error) {
	file, err := v.fsLocation.NewFile(fileURI)
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}
	return file.Size()
}

func (v *VFSProtocolHandler) GetPath(fileURI string) string {
	file, _ := v.fsLocation.NewFile(fileURI)
	return file.Path()
}

func (v *VFSProtocolHandler) GetName(fileURI string) string {
	file, _ := v.fsLocation.NewFile(fileURI)
	return file.Name()
}



func (v *VFSProtocolHandler) DeleteFile(filePath string) error {
	return v.fsLocation.DeleteFile(filePath)
}

func (v *VFSProtocolHandler) MoveFile(sourcePath string, destLocationURI string) error {
	file, err := v.fsLocation.NewFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to create VFS file: %w", err)
	}
	defer file.Close()
	relPath, err := filepath.Rel(v.fsLocation.URI(), destLocationURI)
	if err != nil {
		return fmt.Errorf("failed to get relative destination location: %w", err)
	}
	destLocation, err := v.fsLocation.NewLocation(relPath + "/")
	if err != nil {
		return fmt.Errorf("failed to create VFS location: %w", err)
	}
	_, err = file.MoveToLocation(destLocation)
	if err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}
	return nil
}
                                                                                                                                                                                                                                                            