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
	"errors"

	"github.com/apache/synapse-go/internal/app/adapters/inbound/file"
	"github.com/apache/synapse-go/internal/app/adapters/inbound/http"
	"github.com/apache/synapse-go/internal/app/core/domain"
	"github.com/apache/synapse-go/internal/app/core/ports"
)

var (
	ErrInboundTypeNotFound = errors.New("inbound type not found")
)

func NewInbound(config domain.InboundConfig) (ports.InboundEndpoint, error) {
	switch config.Protocol {
	case "file":
		return file.NewFileInboundEndpoint(
			config,
			nil,
		), nil

	case "http":
		return http.NewHTTPInboundEndpoint(
			config,
			nil,
		), nil
	default:
		return nil, ErrInboundTypeNotFound
	}
}