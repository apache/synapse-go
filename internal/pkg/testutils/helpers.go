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

package testutils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
	"github.com/apache/synapse-go/internal/pkg/core/synctx"
	"github.com/apache/synapse-go/internal/pkg/core/utils"
)

// TestConfigContext creates a test configuration context with common test endpoints
func CreateTestConfigContext() *artifacts.ConfigContext {
	return &artifacts.ConfigContext{
		ApiMap:      make(map[string]artifacts.API),
		EndpointMap: make(map[string]artifacts.Endpoint),
		SequenceMap: make(map[string]artifacts.Sequence),
		InboundMap:  make(map[string]artifacts.Inbound),
	}
}

// CreateTestCallMediator creates a test CallMediator with specified endpoint reference
func CreateTestCallMediator(endpointRef string) artifacts.CallMediator {
	return artifacts.CallMediator{
		EndpointRef: endpointRef,
		Position:    artifacts.Position{Hierarchy: "test.hierarchy"},
	}
}

// CreateTestMsgContext creates a test message context with JSON payload
func CreateTestMsgContext(payload []byte) *synctx.MsgContext {
	ctx := synctx.CreateMsgContext()
	ctx.Message.RawPayload = payload
	ctx.Message.ContentType = "application/json"
	return ctx
}

// CreateTestMsgContextWithHeaders creates a test message context with custom headers
func CreateTestMsgContextWithHeaders(payload []byte, headers map[string]string) *synctx.MsgContext {
	ctx := CreateTestMsgContext(payload)
	if ctx.Headers == nil {
		ctx.Headers = make(map[string]string)
	}
	for k, v := range headers {
		ctx.Headers[k] = v
	}
	return ctx
}

// CreateTestEndpoint creates a test endpoint with specified URL and method
func CreateTestEndpoint(name, method, url string) artifacts.Endpoint {
	return artifacts.Endpoint{
		Name: name,
		EndpointUrl: artifacts.EndpointUrl{
			Method:      method,
			URITemplate: url,
		},
	}
}

// CreateTestContext creates a context with test configuration
func CreateTestContext(configContext *artifacts.ConfigContext) context.Context {
	ctx := context.Background()
	if configContext != nil {
		ctx = context.WithValue(ctx, utils.ConfigContextKey, configContext)
	}
	return ctx
}

// CreateMockHTTPServer creates a mock HTTP server for testing
func CreateMockHTTPServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"method":    r.Method,
			"path":      r.URL.Path,
			"timestamp": time.Now().Format(time.RFC3339),
			"status":    "success",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
}

// CreateMockErrorServer creates a mock server that returns errors
func CreateMockErrorServer(statusCode int, errorMessage string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write([]byte(fmt.Sprintf(`{"error":"%s","code":%d}`, errorMessage, statusCode)))
	}))
}

// CreateSlowMockServer creates a mock server with configurable delay
func CreateSlowMockServer(delay time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"slow response","delay":"` + delay.String() + `"}`))
	}))
}

// AssertJSONEqual compares two JSON byte arrays for equality
func AssertJSONEqual(t interface {
	Errorf(format string, args ...interface{})
}, expected, actual []byte) {
	var expectedJSON, actualJSON interface{}

	if err := json.Unmarshal(expected, &expectedJSON); err != nil {
		t.Errorf("Failed to unmarshal expected JSON: %v", err)
		return
	}

	if err := json.Unmarshal(actual, &actualJSON); err != nil {
		t.Errorf("Failed to unmarshal actual JSON: %v", err)
		return
	}

	expectedStr, _ := json.Marshal(expectedJSON)
	actualStr, _ := json.Marshal(actualJSON)

	if string(expectedStr) != string(actualStr) {
		t.Errorf("JSON mismatch:\nExpected: %s\nActual: %s", string(expectedStr), string(actualStr))
	}
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(condition func() bool, timeout time.Duration, interval time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}
	return false
}

// CreateTestAPI creates a test API artifact
func CreateTestAPI(name, context, version string) artifacts.API {
	return artifacts.API{
		Name:    name,
		Context: context,
		Version: version,
		Resources: []artifacts.Resource{
			{
				Methods: []string{"GET", "POST"},
				URITemplate: artifacts.URITemplateInfo{
					FullTemplate: "/test",
					PathTemplate: "/test",
				},
				InSequence: artifacts.Sequence{
					Name: "testInSequence",
					MediatorList: []artifacts.Mediator{
						artifacts.LogMediator{
							Message: "Test log message",
						},
					},
				},
			},
		},
	}
}

// CreateTestSequence creates a test sequence artifact
func CreateTestSequence(name string, mediators []artifacts.Mediator) artifacts.Sequence {
	return artifacts.Sequence{
		Name:         name,
		MediatorList: mediators,
	}
}

// CreateTestInbound creates a test inbound endpoint
func CreateTestInbound(name, protocol, sequenceName string, parameters map[string]string) artifacts.Inbound {
	var params []artifacts.Parameter
	for key, value := range parameters {
		params = append(params, artifacts.Parameter{
			Name:  key,
			Value: value,
		})
	}

	return artifacts.Inbound{
		Name:       name,
		Protocol:   protocol,
		Sequence:   sequenceName,
		Parameters: params,
	}
}

// MockMediationEngine for testing
type MockMediationEngine struct {
	RegisteredSequences map[string]artifacts.Sequence
	ProcessedMessages   []string
}

func NewMockMediationEngine() *MockMediationEngine {
	return &MockMediationEngine{
		RegisteredSequences: make(map[string]artifacts.Sequence),
		ProcessedMessages:   make([]string, 0),
	}
}

func (m *MockMediationEngine) RegisterSequence(name string, sequence artifacts.Sequence) error {
	m.RegisteredSequences[name] = sequence
	return nil
}

func (m *MockMediationEngine) ProcessMessage(ctx context.Context, msgContext *synctx.MsgContext, sequenceName string) error {
	m.ProcessedMessages = append(m.ProcessedMessages, sequenceName)
	return nil
}

// MockRouterService for testing
type MockRouterService struct {
	RegisteredAPIs []artifacts.API
	ServerStarted  bool
}

func NewMockRouterService() *MockRouterService {
	return &MockRouterService{
		RegisteredAPIs: make([]artifacts.API, 0),
		ServerStarted:  false,
	}
}

func (m *MockRouterService) RegisterAPI(api artifacts.API) error {
	m.RegisteredAPIs = append(m.RegisteredAPIs, api)
	return nil
}

func (m *MockRouterService) Start(ctx context.Context) error {
	m.ServerStarted = true
	<-ctx.Done()
	m.ServerStarted = false
	return nil
}

func (m *MockRouterService) Stop() error {
	m.ServerStarted = false
	return nil
}

// PerformanceTimer for measuring test execution time
type PerformanceTimer struct {
	start time.Time
	name  string
}

func NewPerformanceTimer(name string) *PerformanceTimer {
	return &PerformanceTimer{
		start: time.Now(),
		name:  name,
	}
}

func (p *PerformanceTimer) Stop() time.Duration {
	duration := time.Since(p.start)
	fmt.Printf("Performance: %s took %v\n", p.name, duration)
	return duration
}

// TestDataGenerator helps generate test data
type TestDataGenerator struct {
	counter int
}

func NewTestDataGenerator() *TestDataGenerator {
	return &TestDataGenerator{counter: 0}
}

func (g *TestDataGenerator) GenerateJSONPayload(dataType string) []byte {
	g.counter++

	switch dataType {
	case "order":
		return []byte(fmt.Sprintf(`{
			"orderId": "ORD-%03d",
			"customerId": "CUST-%03d",
			"amount": %.2f,
			"timestamp": "%s",
			"items": [
				{"id": "ITEM-%03d", "quantity": %d}
			]
		}`, g.counter, g.counter, float64(g.counter)*10.50, time.Now().Format(time.RFC3339), g.counter, g.counter%5+1))

	case "customer":
		return []byte(fmt.Sprintf(`{
			"customerId": "CUST-%03d",
			"name": "Customer %d",
			"email": "customer%d@example.com",
			"phone": "+1-555-%04d"
		}`, g.counter, g.counter, g.counter, 1000+g.counter))

	default:
		return []byte(fmt.Sprintf(`{
			"id": %d,
			"type": "%s",
			"timestamp": "%s",
			"data": "Generated test data %d"
		}`, g.counter, dataType, time.Now().Format(time.RFC3339), g.counter))
	}
}
