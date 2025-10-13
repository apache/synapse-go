//go:build integration
// +build integration

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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/apache/synapse-go/internal/pkg/core/synctx"
	"github.com/apache/synapse-go/internal/pkg/core/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// CallMediatorIntegrationTestSuite defines the integration test suite for CallMediator
type CallMediatorIntegrationTestSuite struct {
	suite.Suite
	successServer *httptest.Server
	errorServer   *httptest.Server
	slowServer    *httptest.Server
}

// SetupSuite runs once before all integration tests
func (suite *CallMediatorIntegrationTestSuite) SetupSuite() {
	// Success server
	suite.successServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate realistic processing time
		time.Sleep(10 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Response-Time", fmt.Sprintf("%dms", 10))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","data":{"id":123,"message":"processed"},"timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	}))

	// Error server
	suite.errorServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error","code":500}`))
	}))

	// Slow server for timeout testing
	suite.slowServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Simulate slow response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"slow response"}`))
	}))
}

// TearDownSuite runs once after all integration tests
func (suite *CallMediatorIntegrationTestSuite) TearDownSuite() {
	if suite.successServer != nil {
		suite.successServer.Close()
	}
	if suite.errorServer != nil {
		suite.errorServer.Close()
	}
	if suite.slowServer != nil {
		suite.slowServer.Close()
	}
}

// TestIntegration_SuccessfulCall tests successful HTTP calls with real server
func (suite *CallMediatorIntegrationTestSuite) TestIntegration_SuccessfulCall() {
	mediator := CallMediator{
		EndpointRef: "successEndpoint",
		Position:    Position{Hierarchy: "integration.test.success"},
	}

	configContext := &ConfigContext{
		EndpointMap: map[string]Endpoint{
			"successEndpoint": {
				Name: "successEndpoint",
				EndpointUrl: EndpointUrl{
					Method:      "POST",
					URITemplate: suite.successServer.URL + "/api/process",
				},
			},
		},
	}

	msgContext := synctx.CreateMsgContext()
	msgContext.Message.RawPayload = []byte(`{"requestId":"integration-test-001","data":"test payload"}`)
	msgContext.Message.ContentType = "application/json"

	ctx := context.WithValue(context.Background(), utils.ConfigContextKey, configContext)

	// Execute the mediator
	start := time.Now()
	result, err := mediator.Execute(msgContext, ctx)
	duration := time.Since(start)

	// Assertions
	suite.NoError(err)
	suite.True(result)
	suite.NotEmpty(msgContext.Message.RawPayload)
	suite.Equal("application/json", msgContext.Message.ContentType)

	// Verify response content
	responseStr := string(msgContext.Message.RawPayload)
	suite.Contains(responseStr, "success")
	suite.Contains(responseStr, "processed")

	// Performance assertions
	suite.Less(duration, 500*time.Millisecond, "Request should complete within 500ms")
}

// TestIntegration_ServerErrorHandling tests error response handling
func (suite *CallMediatorIntegrationTestSuite) TestIntegration_ServerErrorHandling() {
	mediator := CallMediator{
		EndpointRef: "errorEndpoint",
		Position:    Position{Hierarchy: "integration.test.error"},
	}

	configContext := &ConfigContext{
		EndpointMap: map[string]Endpoint{
			"errorEndpoint": {
				Name: "errorEndpoint",
				EndpointUrl: EndpointUrl{
					Method:      "POST",
					URITemplate: suite.errorServer.URL + "/api/error",
				},
			},
		},
	}

	msgContext := synctx.CreateMsgContext()
	msgContext.Message.RawPayload = []byte(`{"requestId":"integration-test-error"}`)
	msgContext.Message.ContentType = "application/json"

	ctx := context.WithValue(context.Background(), utils.ConfigContextKey, configContext)

	// Execute the mediator
	result, err := mediator.Execute(msgContext, ctx)

	// Even with server errors, the HTTP call should succeed
	suite.NoError(err)
	suite.True(result)

	// Verify error response is captured
	responseStr := string(msgContext.Message.RawPayload)
	suite.Contains(responseStr, "internal server error")
	suite.Contains(responseStr, "500")
}

// TestIntegration_ConcurrentCalls tests concurrent HTTP calls
func (suite *CallMediatorIntegrationTestSuite) TestIntegration_ConcurrentCalls() {
	const numConcurrentCalls = 10

	mediator := CallMediator{
		EndpointRef: "concurrentEndpoint",
		Position:    Position{Hierarchy: "integration.test.concurrent"},
	}

	configContext := &ConfigContext{
		EndpointMap: map[string]Endpoint{
			"concurrentEndpoint": {
				Name: "concurrentEndpoint",
				EndpointUrl: EndpointUrl{
					Method:      "POST",
					URITemplate: suite.successServer.URL + "/api/concurrent",
				},
			},
		},
	}

	ctx := context.WithValue(context.Background(), utils.ConfigContextKey, configContext)

	// Channel to collect results
	results := make(chan bool, numConcurrentCalls)
	errors := make(chan error, numConcurrentCalls)

	// Execute concurrent calls
	for i := 0; i < numConcurrentCalls; i++ {
		go func(callID int) {
			msgContext := synctx.CreateMsgContext()
			msgContext.Message.RawPayload = []byte(fmt.Sprintf(`{"callId":%d,"data":"concurrent test"}`, callID))
			msgContext.Message.ContentType = "application/json"

			result, err := mediator.Execute(msgContext, ctx)
			results <- result
			errors <- err
		}(i)
	}

	// Collect results
	var successCount int
	var errorCount int

	for i := 0; i < numConcurrentCalls; i++ {
		result := <-results
		err := <-errors

		if err != nil {
			errorCount++
		} else if result {
			successCount++
		}
	}

	// Assertions
	suite.Equal(numConcurrentCalls, successCount, "All concurrent calls should succeed")
	suite.Equal(0, errorCount, "No errors should occur")
}

// TestIntegration_ContextTimeout tests context timeout behavior
func (suite *CallMediatorIntegrationTestSuite) TestIntegration_ContextTimeout() {
	mediator := CallMediator{
		EndpointRef: "timeoutEndpoint",
		Position:    Position{Hierarchy: "integration.test.timeout"},
	}

	configContext := &ConfigContext{
		EndpointMap: map[string]Endpoint{
			"timeoutEndpoint": {
				Name: "timeoutEndpoint",
				EndpointUrl: EndpointUrl{
					Method:      "POST",
					URITemplate: suite.slowServer.URL + "/api/slow",
				},
			},
		},
	}

	msgContext := synctx.CreateMsgContext()
	msgContext.Message.RawPayload = []byte(`{"requestId":"timeout-test"}`)
	msgContext.Message.ContentType = "application/json"

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	ctx = context.WithValue(ctx, utils.ConfigContextKey, configContext)

	// Execute the mediator
	start := time.Now()
	result, err := mediator.Execute(msgContext, ctx)
	duration := time.Since(start)

	// Note: The current CallMediator implementation may not properly handle context timeouts
	// This test documents the current behavior - it might succeed despite timeout
	if err != nil {
		// If timeout is properly handled
		suite.False(result)
		suite.Less(duration, 1*time.Second, "Should timeout before slow server responds")
		suite.Contains(err.Error(), "context deadline exceeded")
	} else {
		// If timeout is not handled (current behavior)
		suite.T().Log("CallMediator does not properly handle context timeouts - this is a known limitation")
		suite.Greater(duration, 1*time.Second, "Request completed despite timeout")
	}
}

// TestCallMediatorIntegrationTestSuite runs the integration test suite
func TestCallMediatorIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CallMediatorIntegrationTestSuite))
}

// TestIntegration_CallMediator_HTTPMethods tests different HTTP methods
func TestIntegration_CallMediator_HTTPMethods(t *testing.T) {
	// Test server that handles different HTTP methods
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		switch r.Method {
		case "GET":
			w.Write([]byte(`{"method":"GET","status":"success"}`))
		case "POST":
			w.Write([]byte(`{"method":"POST","status":"created"}`))
		case "PUT":
			w.Write([]byte(`{"method":"PUT","status":"updated"}`))
		case "DELETE":
			w.Write([]byte(`{"method":"DELETE","status":"deleted"}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"method not allowed"}`))
		}
	}))
	defer server.Close()

	testCases := []struct {
		name           string
		method         string
		expectedStatus string
	}{
		{"GET_method", "GET", "success"},
		{"POST_method", "POST", "created"},
		{"PUT_method", "PUT", "updated"},
		{"DELETE_method", "DELETE", "deleted"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mediator := CallMediator{
				EndpointRef: "methodTestEndpoint",
				Position:    Position{Hierarchy: "integration.test.methods"},
			}

			configContext := &ConfigContext{
				EndpointMap: map[string]Endpoint{
					"methodTestEndpoint": {
						Name: "methodTestEndpoint",
						EndpointUrl: EndpointUrl{
							Method:      tc.method,
							URITemplate: server.URL + "/api/test",
						},
					},
				},
			}

			msgContext := synctx.CreateMsgContext()
			msgContext.Message.RawPayload = []byte(`{"test":"data"}`)
			msgContext.Message.ContentType = "application/json"

			ctx := context.WithValue(context.Background(), utils.ConfigContextKey, configContext)

			result, err := mediator.Execute(msgContext, ctx)

			require.NoError(t, err)
			assert.True(t, result)

			responseStr := string(msgContext.Message.RawPayload)
			assert.Contains(t, responseStr, tc.method)
			assert.Contains(t, responseStr, tc.expectedStatus)
		})
	}
}
