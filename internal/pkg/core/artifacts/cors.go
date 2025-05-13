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
	"strings"
)

// CORSConfig represents CORS configuration for an API
type CORSConfig struct {
	Enabled          bool     // Whether CORS is enabled for this API
	AllowOrigins     []string // Allowed origins, can contain wildcards like "*"
	AllowMethods     []string // Allowed HTTP methods
	AllowHeaders     []string // Allowed HTTP headers
	ExposeHeaders    []string // Headers to expose to the client
	AllowCredentials bool     // Whether credentials are allowed
	MaxAge           int      // Preflight cache duration in seconds
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		Enabled:          false,
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// IsOriginAllowed checks if the origin is allowed based on the CORS configuration
func (c *CORSConfig) IsOriginAllowed(origin string) bool {
	if !c.Enabled {
		return false
	}

	for _, allowedOrigin := range c.AllowOrigins {
		if allowedOrigin == "*" {
			return true
		}
		if allowedOrigin == origin {
			return true
		}
		// Support for subdomain wildcards like "*.example.com"
		if strings.HasPrefix(allowedOrigin, "*.") {
			suffix := allowedOrigin[1:] // e.g., "*.example.com" -> "*.example.com"
			domain := suffix[1:]        // e.g., "*.example.com" -> "example.com"
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}
	return false
}