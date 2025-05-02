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

package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/apache/synapse-go/internal/app/core/domain"
	"github.com/apache/synapse-go/internal/app/core/ports"
	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
	"github.com/apache/synapse-go/internal/pkg/core/synctx"
	"github.com/apache/synapse-go/internal/pkg/core/utils"
	"github.com/apache/synapse-go/internal/pkg/loggerfactory"
)

const (
	componentName = "http"
)

// HTTPInboundEndpoint handles http-based inbound operations
type HTTPInboundEndpoint struct {
	config    domain.InboundConfig
	mediator  ports.InboundMessageMediator
	IsRunning bool
	server    *http.Server
	router    *http.ServeMux
	logger    *slog.Logger
}

// NewHTTPInboundEndpoint creates a new HTTPInboundEndpoint instance
func NewHTTPInboundEndpoint(
	config domain.InboundConfig,
	mediator ports.InboundMessageMediator,
) *HTTPInboundEndpoint {
	h := &HTTPInboundEndpoint{
		config: config,
		router: http.NewServeMux(),
	}
	h.logger = loggerfactory.GetLogger(componentName, h)
	return h
}

func (h *HTTPInboundEndpoint) Start(ctx context.Context, mediator ports.InboundMessageMediator) error {
	// Check if context is already canceled before proceeding
	select {
	case <-ctx.Done():
		// Context already canceled, don't decrement WaitGroup
		return ctx.Err()
	default:
		// Context still valid, proceed with normal operation
	}

	h.IsRunning = true
	h.mediator = mediator

	// Set up the HTTP handler for the root path
	h.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Create message context
		msgContext := synctx.CreateMsgContext()

		// Set request into message context properties
		msgContext.Properties["http_request_body"] = r.Body

		// Mediate the inbound message
		if err := h.mediator.MediateInboundMessage(ctx, h.config.SequenceName, msgContext); err != nil {
			h.logger.Error("Error mediating inbound message", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Check if the http-response flag is set by the respond mediator
		responseFlag, exists := msgContext.Headers["http-response"]
		if !exists || responseFlag != "true" {
			// Send a simple 202 Accepted response when http-response flag is not properly set
			h.logger.Debug("http-response flag not set, sending 202 Accepted response")
			// Send 202 Accepted status
			w.WriteHeader(http.StatusAccepted)
			return
		}

		// Send the response back to the client
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Inbound Mediation successful"}`))
	})

	inboundPortStr := h.config.Parameters["inbound.http.port"]

	// Convert port to integer
	inboundServerPort, err := strconv.Atoi(inboundPortStr)
	if err != nil {
		h.logger.Error("Invalid port configuration", "port", inboundPortStr, "error", err)
		return errors.New("invalid inbound HTTP port configuration")
	}
	

	var hostname string
	// Get config context from the context
	configCtx := ctx.Value(utils.ConfigContextKey).(*artifacts.ConfigContext)

	if serverConfig, ok := configCtx.DeploymentConfig["server"].(map[string]string); ok {
		hostname = serverConfig["hostname"]
		if offsetStr, offsetExists := serverConfig["offset"]; offsetExists {
			if offsetInt, err := strconv.Atoi(offsetStr); err == nil {
				inboundServerPort += offsetInt
				h.logger.Info("Using port offset", "offset", offsetInt, "final_port", inboundServerPort)
			} else {
				h.logger.Warn("Warning: Invalid offset value, using default port", "offset", offsetStr)
			}
		}
	}

	// Convert the port to a string format expected by the HTTP server
	listenPort := fmt.Sprintf(":%d", inboundServerPort)

	//eg:- localhost:8290
	listenAddr := hostname + listenPort


	// Create a new HTTP server
	h.server = &http.Server{
		Addr:    listenAddr,
		Handler: h.router,
	}

	// Start the server in a goroutine
	go func() {
		h.logger.Info("Starting HTTP Inbound listener", "address", listenAddr)
		if err := h.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			h.logger.Error("HTTP Inbound listener error", "error", err)
		}
		h.logger.Info("HTTP inbound server stopped serving new connections")
	}()

	if err := h.Stop(ctx); err != nil {
		h.logger.Error("Error shutting down HTTP Inbound server", "error", err.Error())
	}
	h.logger.Info("HTTP Inbound server shut down gracefully")
	return nil
}

// Stops HTTP server gracefully
func (h *HTTPInboundEndpoint) Stop(ctx context.Context) error {
	<-ctx.Done()
	h.logger.Info("Shutting down HTTP Inbond server...")
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()
	if err := h.server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	h.IsRunning = false
	return nil
}

func (h *HTTPInboundEndpoint) UpdateLogger() {
	h.logger = loggerfactory.GetLogger(componentName, h)
}