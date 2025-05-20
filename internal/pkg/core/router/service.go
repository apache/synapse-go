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

// Package router provides HTTP routing capabilities for Synapse APIs.
//
// The RouterService is the main component of this package, providing:
// - API registration with automatic route creation from resources
// - HTTP server lifecycle management with automatic start/stop
// - Request handling with conversion to/from Synapse message contexts
// - Method-based routing for RESTful APIs

package router

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"encoding/json"

	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
	"github.com/apache/synapse-go/internal/pkg/core/synctx"
	"github.com/apache/synapse-go/internal/pkg/loggerfactory"
)

const (
	componentName = "router"
)

// RouterService manages API routing and server lifecycle
type RouterService struct {
	server   *http.Server
	router   *http.ServeMux
	port     string // :8290
	hostname string
	logger   *slog.Logger
}

// NewRouterService creates a new router service with the given port and hostname
func NewRouterService(port string, hostname string) *RouterService {
	rs := &RouterService{
		router:   http.NewServeMux(),
		hostname: hostname,
		port:     port,
	}
	rs.logger = loggerfactory.GetLogger(componentName, rs)
	return rs
}

func (rs *RouterService) UpdateLogger() {
	rs.logger = loggerfactory.GetLogger(componentName, rs)
}

// RegisterAPI registers a new API with the router service
func (rs *RouterService) RegisterAPI(ctx context.Context, api artifacts.API) error {
	// Determine base path based on context and version
	basePath := api.Context

	// Remove trailing slash from context if present
	if len(basePath) > 1 && basePath[len(basePath)-1] == '/' {
		basePath = basePath[:len(basePath)-1]
	}

	// Handle versioning based on versionType
	if api.Version != "" && api.VersionType != "" {
		switch api.VersionType {
		case "url":
			// For URL type, add version as a path segment
			basePath = basePath + "/" + api.Version
		case "context":
			// For context type, replace {version} placeholder if it exists
			versionPattern := "{version}"
			basePath = strings.Replace(basePath, versionPattern, api.Version, 1)
		}
	}

	// Register swagger documentation handlers with appropriate versioning in URL
	// If version is not empty, register at /<API_NAME>:<API_VERSION>
	// If version is empty, register at /<API_NAME>
	swaggerBasePath := "/" + api.Name
	if api.Version != "" {
		swaggerBasePath = swaggerBasePath + ":" + api.Version
	}

	rs.router.HandleFunc(swaggerBasePath, func(w http.ResponseWriter, r *http.Request) {
		// Get the query parameters from the URL
		query := r.URL.Query()
			
		// Check for swagger file extension in the path
		if _, exists := query["swagger.yaml"]; exists {
			api.ServeSwaggerYAML(w, rs.hostname, rs.port)
			return
		}

		if _, exists := query["swagger.json"]; exists {
			api.ServeSwaggerJSON(w, rs.hostname, rs.port)
			return
		}
		http.NotFound(w, r)
	})

	// Create a subrouter for this API
	apiHandler := http.NewServeMux()

	// Register each resource in the API
	for _, resource := range api.Resources {
		// Register a handler for each HTTP method in the resource
		for _, method := range resource.Methods {
			// Construct the full pattern: "METHOD /path/to/resource"
			pattern := method + " " + resource.URITemplate.PathTemplate
			// Create a wrapper handler that checks query parameters before forwarding to the resource handler
			queryParamHandler := rs.createQueryParamMiddleware(resource, rs.createResourceHandler(resource, ctx))
			apiHandler.HandleFunc(pattern, queryParamHandler)
			rs.logger.Info("Registered route for API",
				slog.String("api_name", api.Name),
				slog.String("pattern", pattern))
			// No need to register explicit OPTIONS handlers when using rs/cors package
			// The CORSMiddleware already handles OPTIONS preflight requests automatically
		}
	}

	// Apply CORS middleware to the entire API subrouter if enabled
	var handler http.Handler = apiHandler
	if api.CORSConfig.Enabled {
		handler = CORSMiddleware(handler, api.CORSConfig)
	}

	// Register the API handler with the main router
	rs.router.Handle(basePath+"/", http.StripPrefix(basePath, apiHandler))
	return nil
}

// createHandlerFunc creates an HTTP handler function for the given API resource
func (rs *RouterService) createResourceHandler(resource artifacts.Resource, ctx context.Context) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Create message context
		msgContext := synctx.CreateMsgContext()

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
		  http.Error(w, "Error reading request body", http.StatusBadRequest)
		  return
		}
		r.Body.Close() // Properly close the body
		msgContext.Message.RawPayload = bodyBytes

		msgContext.Message.ContentType = r.Header.Get("Content-Type")

		// Set path parameters into message context properties
		pathParamsMap := make(map[string]string)
		for _, pathParam := range resource.URITemplate.PathParameters {
			pathParamsMap[pathParam] = r.PathValue(pathParam)
		}
		msgContext.Properties["uriParams"] = pathParamsMap

		// Set query parameters into message context properties
		queryParams := r.URL.Query()

		// If there are predefined query parameters, map each to their corresponding variable
		if len(resource.URITemplate.QueryParameters) > 0 {
			// Create a map to store the variable mappings
			queryVarMap := make(map[string]string)

			// Loop through each predefined query parameter
			for paramName, varName := range resource.URITemplate.QueryParameters {
				// Get the value from the request
				if values, exists := queryParams[paramName]; exists && len(values) > 0 {
					// Map the query parameter value to the variable name
					queryVarMap[varName] = values[0]
				}
			}

			// Store the variable mapping in the message context
			msgContext.Properties["queryParams"] = queryVarMap
		}

		// Process through mediation pipeline
		success := resource.Mediate(msgContext, ctx)

		// Write response
		if success {
			for name, value := range msgContext.Headers {
				w.Header().Set(name, value)
			}
			if msgContext.Message.RawPayload != nil {
				w.Write(msgContext.Message.RawPayload)
			}
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
	return handler
}

// createQueryParamMiddleware creates a middleware that validates query parameters against predefined parameters
func (rs *RouterService) createQueryParamMiddleware(resource artifacts.Resource, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If there are no predefined query parameters, just call the next handler
		if len(resource.URITemplate.QueryParameters) == 0 {
			next(w, r)
			return
		}

		// Get query parameters from the request
		queryParams := r.URL.Query()

		// Check if query parameter keys match exactly with the predefined keys
		// First, ensure all request query params exist in predefined params
		for key := range queryParams {
			if _, exists := resource.URITemplate.QueryParameters[key]; !exists {
				// Query parameter not defined in the template, reject the request
				http.Error(w, fmt.Sprintf("Unsupported query parameter: %s", key), http.StatusBadRequest)
				return
			}
		}

		// Now ensure all predefined query params exist in the request
		for key := range resource.URITemplate.QueryParameters {
			if !queryParams.Has(key) {
				// Required query parameter is missing, reject the request
				http.Error(w, fmt.Sprintf("Missing required query parameter: %s", key), http.StatusBadRequest)
				return
			}
		}

		// All parameters in the request are valid and all required parameters are present
		next(w, r)
	}
}

// startServer starts the HTTP server
func (rs *RouterService) StartServer(ctx context.Context) {
	//eg:- localhost:8290
	addr := rs.hostname + rs.port
	rs.server = &http.Server{
		Addr:    addr,
		Handler: rs.router,
	}

	// Register health/liveness endpoints
	rs.registerLivelinessEndpoint()
	rs.logger.Info("liveness endpoint registered")

	// Start the server in a goroutine
	go func() {
		rs.logger.Info("Starting HTTP server", "address", addr)
		if err := rs.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			rs.logger.Error("HTTP server error", slog.String("error", err.Error()))
		}
		rs.logger.Info("HTTP server stopped serving new connections")
	}()
}

func (rs *RouterService) StopServer() {
	if rs.server != nil {
		rs.logger.Info("Shutting down HTTP  server...")
		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownRelease()
		if err := rs.server.Shutdown(shutdownCtx); err != nil {
			rs.logger.Error("Error shutting down HTTP server", "error", err.Error())
		}
	}
}


// registerHealthEndpoints registers health and liveness endpoints
func (rs *RouterService) registerLivelinessEndpoint() {
	// liveliness probe endpoint
	rs.router.HandleFunc("/livez", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "UP",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})
}