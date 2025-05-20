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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORSConfig_IsOriginAllowed(t *testing.T) {
	tests := []struct {
		name     string
		config   CORSConfig
		origin   string
		expected bool
	}{
		{
			name: "CORS disabled",
			config: CORSConfig{
				Enabled:      false,
				AllowOrigins: []string{"https://example.com"},
			},
			origin:   "https://example.com",
			expected: false,
		},
		{
			name: "Wildcard origin",
			config: CORSConfig{
				Enabled:      true,
				AllowOrigins: []string{"*"},
			},
			origin:   "https://example.com",
			expected: true,
		},
		{
			name: "Exact origin match",
			config: CORSConfig{
				Enabled:      true,
				AllowOrigins: []string{"https://example.com"},
			},
			origin:   "https://example.com",
			expected: true,
		},
		{
			name: "Origin not allowed",
			config: CORSConfig{
				Enabled:      true,
				AllowOrigins: []string{"https://example.com"},
			},
			origin:   "https://different.com",
			expected: false,
		},
		{
			name: "Subdomain wildcard - matching",
			config: CORSConfig{
				Enabled:      true,
				AllowOrigins: []string{"*.example.com"},
			},
			origin:   "https://sub.example.com",
			expected: true,
		},
		{
			name: "Subdomain wildcard - non-matching",
			config: CORSConfig{
				Enabled:      true,
				AllowOrigins: []string{"*.example.com"},
			},
			origin:   "https://example.org",
			expected: false,
		},
		{
			name: "Multiple origins - matching one",
			config: CORSConfig{
				Enabled:      true,
				AllowOrigins: []string{"https://example.com", "https://example.org"},
			},
			origin:   "https://example.org",
			expected: true,
		},
		{
			name: "Multiple origins - not matching any",
			config: CORSConfig{
				Enabled:      true,
				AllowOrigins: []string{"https://example.com", "https://example.org"},
			},
			origin:   "https://example.net",
			expected: false,
		},
		{
			name: "Empty origin",
			config: CORSConfig{
				Enabled:      true,
				AllowOrigins: []string{"https://example.com"},
			},
			origin:   "",
			expected: false,
		},
		{
			name: "Empty allowed origins",
			config: CORSConfig{
				Enabled:      true,
				AllowOrigins: []string{},
			},
			origin:   "https://example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsOriginAllowed(tt.origin)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()

	// Verify default values
	assert.False(t, config.Enabled)
	assert.Equal(t, []string{"*"}, config.AllowOrigins)
	assert.Equal(t, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}, config.AllowMethods)
	assert.Equal(t, []string{"Origin", "Content-Type", "Accept", "Authorization"}, config.AllowHeaders)
	assert.Empty(t, config.ExposeHeaders)
	assert.False(t, config.AllowCredentials)
	assert.Equal(t, 86400, config.MaxAge)
}