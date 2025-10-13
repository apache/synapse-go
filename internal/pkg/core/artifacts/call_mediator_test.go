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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/synapse-go/internal/pkg/core/synctx"
	"github.com/apache/synapse-go/internal/pkg/core/utils"
	"github.com/stretchr/testify/assert"
)

func TestCallMediator_Execute(t *testing.T) {
	// Create a test server that will handle the HTTP requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/test", r.URL.Path)

		// Verify request headers and body
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Return a response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"success"}`))
	}))
	defer server.Close()

	// Test cases
	tests := []struct {
		name           string
		mediator       CallMediator
		endpointRef    string
		endpointUrl    EndpointUrl
		msgContext     *synctx.MsgContext
		expectedResult bool
		expectedError  string
	}{
		{
			name: "successful_call",
			mediator: CallMediator{
				EndpointRef: "testEndpoint",
				Position:    Position{Hierarchy: "test.hierarchy"},
			},
			endpointRef: "testEndpoint",
			endpointUrl: EndpointUrl{
				Method:      "POST",
				URITemplate: server.URL + "/test",
			},
			msgContext: func() *synctx.MsgContext {
				ctx := synctx.CreateMsgContext()
				ctx.Message.RawPayload = []byte(`{"test":"data"}`)
				ctx.Message.ContentType = "application/json"
				return ctx
			}(),
			expectedResult: true,
			expectedError:  "",
		},
		{
			name: "empty_endpoint_ref",
			mediator: CallMediator{
				EndpointRef: "",
				Position:    Position{Hierarchy: "test.hierarchy"},
			},
			msgContext:     synctx.CreateMsgContext(),
			expectedResult: false,
			expectedError:  "endpoint reference not provided in call mediator at test.hierarchy",
		},
		{
			name: "endpoint_not_found",
			mediator: CallMediator{
				EndpointRef: "nonExistentEndpoint",
				Position:    Position{Hierarchy: "test.hierarchy"},
			},
			endpointRef:    "testEndpoint", // Different from mediator.EndpointRef
			endpointUrl:    EndpointUrl{},
			msgContext:     synctx.CreateMsgContext(),
			expectedResult: false,
			expectedError:  "endpoint not found with reference: nonExistentEndpoint at test.hierarchy",
		},
		{
			name: "empty_method",
			mediator: CallMediator{
				EndpointRef: "testEndpoint",
				Position:    Position{Hierarchy: "test.hierarchy"},
			},
			endpointRef: "testEndpoint",
			endpointUrl: EndpointUrl{
				Method:      "",
				URITemplate: server.URL + "/test",
			},
			msgContext:     synctx.CreateMsgContext(),
			expectedResult: false,
			expectedError:  "HTTP method not specified for endpoint: testEndpoint at test.hierarchy",
		},
		{
			name: "empty_url",
			mediator: CallMediator{
				EndpointRef: "testEndpoint",
				Position:    Position{Hierarchy: "test.hierarchy"},
			},
			endpointRef: "testEndpoint",
			endpointUrl: EndpointUrl{
				Method:      "POST",
				URITemplate: "",
			},
			msgContext:     synctx.CreateMsgContext(),
			expectedResult: false,
			expectedError:  "endpoint URL is empty for endpoint: testEndpoint at test.hierarchy",
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create config context
			configContext := &ConfigContext{
				EndpointMap: map[string]Endpoint{},
			}

			// Add endpoint if specified
			if tc.endpointRef != "" {
				configContext.EndpointMap[tc.endpointRef] = Endpoint{
					Name:        tc.endpointRef,
					EndpointUrl: tc.endpointUrl,
				}
			}

			// Create context with config context
			ctx := context.WithValue(context.Background(), utils.ConfigContextKey, configContext)

			// Execute the mediator
			result, err := tc.mediator.Execute(tc.msgContext, ctx)

			// Verify results
			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.True(t, result)

				// For successful cases, verify response was stored in message context
				if result {
					assert.NotEmpty(t, tc.msgContext.Message.RawPayload)
					assert.Equal(t, "application/json", tc.msgContext.Message.ContentType)
					assert.Contains(t, string(tc.msgContext.Message.RawPayload), "success")
				}
			} else {
				assert.EqualError(t, err, tc.expectedError)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

// TestCallMediatorWithInvalidConfigContext tests the behavior when an invalid config context is provided
func TestCallMediatorWithInvalidConfigContext(t *testing.T) {
	mediator := CallMediator{
		EndpointRef: "testEndpoint",
		Position:    Position{Hierarchy: "test.hierarchy"},
	}
	msgContext := synctx.CreateMsgContext()

	// Test with nil config context
	t.Run("nil_config_context", func(t *testing.T) {
		ctx := context.Background() // No value in context
		result, err := mediator.Execute(msgContext, ctx)
		assert.False(t, result)
		assert.EqualError(t, err, "config context not found in context at test.hierarchy")
	})

	// Test with wrong type in config context
	t.Run("invalid_config_context_type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), utils.ConfigContextKey, "not a ConfigContext")
		result, err := mediator.Execute(msgContext, ctx)
		assert.False(t, result)
		assert.EqualError(t, err, "invalid config context type")
	})
}

// TestCallMediatorServerErrors tests the behavior when the HTTP server returns errors
func TestCallMediatorServerErrors(t *testing.T) {
	// Create a test server that returns errors
	serverWithErrors := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a 500 error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer serverWithErrors.Close()

	mediator := CallMediator{
		EndpointRef: "testEndpoint",
		Position:    Position{Hierarchy: "test.hierarchy"},
	}

	// Create config context
	configContext := &ConfigContext{
		EndpointMap: map[string]Endpoint{
			"testEndpoint": {
				Name: "testEndpoint",
				EndpointUrl: EndpointUrl{
					Method:      "POST",
					URITemplate: serverWithErrors.URL + "/test",
				},
			},
		},
	}

	msgContext := synctx.CreateMsgContext()
	msgContext.Message.RawPayload = []byte(`{"test":"data"}`)
	msgContext.Message.ContentType = "application/json"

	// Create context with config context
	ctx := context.WithValue(context.Background(), utils.ConfigContextKey, configContext)

	// Execute the mediator
	result, err := mediator.Execute(msgContext, ctx)

	// Even with a 500 status code, we should get a successful execution
	// since the HTTP call itself succeeded
	assert.True(t, result)
	assert.NoError(t, err)

	// Verify response was stored in message context
	assert.NotEmpty(t, msgContext.Message.RawPayload)
	assert.Equal(t, "application/json", msgContext.Message.ContentType)
	assert.Contains(t, string(msgContext.Message.RawPayload), "error")
}
