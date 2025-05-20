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

package router

import (
	"net/http"

	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
	"github.com/rs/cors"
)

// CORSMiddleware applies CORS headers based on the provided configuration using rs/cors package
func CORSMiddleware(handler http.Handler, config artifacts.CORSConfig) http.Handler {
	// Skip CORS handling if disabled
	if !config.Enabled {
		return handler
	}

	// Convert our config to rs/cors options
	options := cors.Options{
		AllowedOrigins:   config.AllowOrigins,
		AllowedMethods:   config.AllowMethods,
		AllowedHeaders:   config.AllowHeaders,
		ExposedHeaders:   config.ExposeHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	}

	// Create the cors handler
	corsHandler := cors.New(options)

	// Use the handler as middleware
	return corsHandler.Handler(handler)
}
