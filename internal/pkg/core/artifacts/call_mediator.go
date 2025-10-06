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

package artifacts

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/apache/synapse-go/internal/pkg/core/synctx"
	"github.com/apache/synapse-go/internal/pkg/core/utils"
)

type CallMediator struct {
	EndpointRef string
	Position    Position
}

func (cm CallMediator) Execute(msgContext *synctx.MsgContext, ctx context.Context) (bool, error) {
	if cm.EndpointRef == "" {
		return false, fmt.Errorf("endpoint reference not provided in call mediator at %s", cm.Position.Hierarchy)
	}

	// Get the ConfigContext from the context
	configContextValue := ctx.Value(utils.ConfigContextKey)
	if configContextValue == nil {
		return false, fmt.Errorf("config context not found in context at %s", cm.Position.Hierarchy)
	}

	configContext, ok := configContextValue.(*ConfigContext)
	if !ok {
		return false, fmt.Errorf("invalid config context type")
	}

	// Find the endpoint in the ConfigContext's EndpointMap
	endpoint, exists := configContext.EndpointMap[cm.EndpointRef]
	if !exists {
		return false, fmt.Errorf("endpoint not found with reference: %s at %s", cm.EndpointRef, cm.Position.Hierarchy)
	}

	// Create an HTTP client for making requests
	client := &http.Client{}

	// Determine HTTP method
	method := endpoint.EndpointUrl.Method
	if method == "" {
		return false, fmt.Errorf("HTTP method not specified for endpoint: %s at %s", cm.EndpointRef, cm.Position.Hierarchy)
	}

	// Get the URL from the endpoint
	url := endpoint.EndpointUrl.URITemplate
	if url == "" {
		return false, fmt.Errorf("endpoint URL is empty for endpoint: %s at %s", cm.EndpointRef, cm.Position.Hierarchy)
	}

	// Create an io.Reader from the byte slice
	payloadReader := bytes.NewReader(msgContext.Message.RawPayload)

	// Create request
	req, err := http.NewRequest(method, url, payloadReader)
	if err != nil {
		return false, fmt.Errorf("failed to create request for endpoint %s: %v", cm.EndpointRef, err)
	}

	// Add content-type header from msgContext ContentType
	req.Header.Set("Content-Type", msgContext.Message.ContentType)

	// Execute the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to execute request for endpoint %s: %v", cm.EndpointRef, err)
	}
	defer resp.Body.Close()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body for endpoint %s: %v", cm.EndpointRef, err)
	}
	// Set the response body to the message context
	msgContext.Message.RawPayload = bodyBytes
	msgContext.Message.ContentType = resp.Header.Get("Content-Type")
	
	// Return true to continue mediation
	return true, nil
}
