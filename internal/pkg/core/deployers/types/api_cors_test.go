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
package types

import (
	"testing"

	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
	"github.com/stretchr/testify/assert"
)

func TestAPI_Unmarshal_WithCORS(t *testing.T) {
	xmlData := `
	<api context="/test" name="TestAPI">
		<cors enabled="true" 
			allow-origins="https://example.com,https://app.example.com" 
			allow-methods="GET,POST,PUT" 
			allow-headers="Content-Type,Authorization" 
			expose-headers="X-Response-Time" 
			allow-credentials="true" 
			max-age="3600" />
		<resource methods="GET" uri-template="/resource1">
			<inSequence>
				<log level="full"/>
			</inSequence>
		</resource>
	</api>`

	position := artifacts.Position{
		FileName: "testfile.xml",
		LineNo:   1,
	}

	api := &API{}
	result, err := api.Unmarshal(xmlData, position)
	assert.NoError(t, err)

	// Verify CORS configuration
	assert.True(t, result.CORSConfig.Enabled)
	assert.Equal(t, []string{"https://example.com", "https://app.example.com"}, result.CORSConfig.AllowOrigins)
	assert.Equal(t, []string{"GET", "POST", "PUT"}, result.CORSConfig.AllowMethods)
	assert.Equal(t, []string{"Content-Type", "Authorization"}, result.CORSConfig.AllowHeaders)
	assert.Equal(t, []string{"X-Response-Time"}, result.CORSConfig.ExposeHeaders)
	assert.True(t, result.CORSConfig.AllowCredentials)
	assert.Equal(t, 3600, result.CORSConfig.MaxAge)
}

func TestAPI_Unmarshal_WithDisabledCORS(t *testing.T) {
	xmlData := `
	<api context="/test" name="TestAPI">
		<cors enabled="false" 
			allow-origins="https://example.com" 
			allow-methods="GET" />
		<resource methods="GET" uri-template="/resource1">
			<inSequence>
				<log level="full"/>
			</inSequence>
		</resource>
	</api>`

	position := artifacts.Position{
		FileName: "testfile.xml",
		LineNo:   1,
	}

	api := &API{}
	result, err := api.Unmarshal(xmlData, position)
	assert.NoError(t, err)

	// Verify CORS is disabled but other configs are still parsed
	assert.False(t, result.CORSConfig.Enabled)
	assert.Equal(t, []string{"https://example.com"}, result.CORSConfig.AllowOrigins)
	assert.Equal(t, []string{"GET"}, result.CORSConfig.AllowMethods)
}

func TestAPI_Unmarshal_WithPartialCORSConfig(t *testing.T) {
	xmlData := `
	<api context="/test" name="TestAPI">
		<cors enabled="true" 
			allow-origins="https://example.com,*.example.org" />
		<resource methods="GET" uri-template="/resource1">
			<inSequence>
				<log level="full"/>
			</inSequence>
		</resource>
	</api>`

	position := artifacts.Position{
		FileName: "testfile.xml",
		LineNo:   1,
	}

	api := &API{}
	result, err := api.Unmarshal(xmlData, position)
	assert.NoError(t, err)

	// Verify partial CORS config with defaults for unspecified values
	assert.True(t, result.CORSConfig.Enabled)
	assert.Equal(t, []string{"https://example.com", "*.example.org"}, result.CORSConfig.AllowOrigins)

	// These should have default values
	assert.Equal(t, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}, result.CORSConfig.AllowMethods)
	assert.Equal(t, []string{"Origin", "Content-Type", "Accept", "Authorization"}, result.CORSConfig.AllowHeaders)
	assert.Empty(t, result.CORSConfig.ExposeHeaders)
	assert.False(t, result.CORSConfig.AllowCredentials)
	assert.Equal(t, 86400, result.CORSConfig.MaxAge)
}

func TestAPI_Unmarshal_WithInvalidMaxAge(t *testing.T) {
	xmlData := `
	<api context="/test" name="TestAPI">
		<cors enabled="true" 
			max-age="invalid" />
		<resource methods="GET" uri-template="/resource1">
			<inSequence>
				<log level="full"/>
			</inSequence>
		</resource>
	</api>`

	position := artifacts.Position{
		FileName: "testfile.xml",
		LineNo:   1,
	}

	api := &API{}
	result, err := api.Unmarshal(xmlData, position)
	assert.NoError(t, err)

	// Should use default max age when invalid value is provided
	assert.Equal(t, 86400, result.CORSConfig.MaxAge)
}

func TestAPI_Unmarshal_WithEmptyCORS(t *testing.T) {
	xmlData := `
	<api context="/test" name="TestAPI">
		<cors />
		<resource methods="GET" uri-template="/resource1">
			<inSequence>
				<log level="full"/>
			</inSequence>
		</resource>
	</api>`

	position := artifacts.Position{
		FileName: "testfile.xml",
		LineNo:   1,
	}

	api := &API{}
	result, err := api.Unmarshal(xmlData, position)
	assert.NoError(t, err)

	// Should use all defaults when empty CORS element is provided
	assert.False(t, result.CORSConfig.Enabled)
	assert.Equal(t, []string{"*"}, result.CORSConfig.AllowOrigins)
	assert.Equal(t, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}, result.CORSConfig.AllowMethods)
	assert.Equal(t, []string{"Origin", "Content-Type", "Accept", "Authorization"}, result.CORSConfig.AllowHeaders)
	assert.Empty(t, result.CORSConfig.ExposeHeaders)
	assert.False(t, result.CORSConfig.AllowCredentials)
	assert.Equal(t, 86400, result.CORSConfig.MaxAge)
}

func TestAPI_Unmarshal_WithoutCORS(t *testing.T) {
	xmlData := `
	<api context="/test" name="TestAPI">
		<resource methods="GET" uri-template="/resource1">
			<inSequence>
				<log level="full"/>
			</inSequence>
		</resource>
	</api>`

	position := artifacts.Position{
		FileName: "testfile.xml",
		LineNo:   1,
	}

	api := &API{}
	result, err := api.Unmarshal(xmlData, position)
	assert.NoError(t, err)

	// Should use default CORS config when no CORS element is provided
	assert.False(t, result.CORSConfig.Enabled)
	assert.Equal(t, []string{"*"}, result.CORSConfig.AllowOrigins)
	assert.Equal(t, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}, result.CORSConfig.AllowMethods)
	assert.Equal(t, []string{"Origin", "Content-Type", "Accept", "Authorization"}, result.CORSConfig.AllowHeaders)
	assert.Empty(t, result.CORSConfig.ExposeHeaders)
	assert.False(t, result.CORSConfig.AllowCredentials)
	assert.Equal(t, 86400, result.CORSConfig.MaxAge)
}

func TestAPI_Unmarshal_WithWhitespaceInCORSValues(t *testing.T) {
	xmlData := `
	<api context="/test" name="TestAPI">
		<cors enabled="true" 
			allow-origins=" https://example.com , https://app.example.com " 
			allow-methods=" GET , POST , PUT " 
			allow-headers=" Content-Type , Authorization " />
		<resource methods="GET" uri-template="/resource1">
			<inSequence>
				<log level="full"/>
			</inSequence>
		</resource>
	</api>`

	position := artifacts.Position{
		FileName: "testfile.xml",
		LineNo:   1,
	}

	api := &API{}
	result, err := api.Unmarshal(xmlData, position)
	assert.NoError(t, err)

	// Verify whitespace is trimmed from CORS values
	assert.Equal(t, []string{"https://example.com", "https://app.example.com"}, result.CORSConfig.AllowOrigins)
	assert.Equal(t, []string{"GET", "POST", "PUT"}, result.CORSConfig.AllowMethods)
	assert.Equal(t, []string{"Content-Type", "Authorization"}, result.CORSConfig.AllowHeaders)
}

func TestAPI_Unmarshal_WithWildcardOrigin(t *testing.T) {
	xmlData := `
	<api context="/test" name="TestAPI">
		<cors enabled="true" 
			allow-origins="*" />
		<resource methods="GET" uri-template="/resource1">
			<inSequence>
				<log level="full"/>
			</inSequence>
		</resource>
	</api>`

	position := artifacts.Position{
		FileName: "testfile.xml",
		LineNo:   1,
	}

	api := &API{}
	result, err := api.Unmarshal(xmlData, position)
	assert.NoError(t, err)

	// Verify wildcard origin is correctly parsed
	assert.Equal(t, []string{"*"}, result.CORSConfig.AllowOrigins)
}

func TestAPI_Unmarshal_WithUnexpectedCORSAttributes(t *testing.T) {
	// This XML has CORS with some unexpected attributes that should be ignored
	xmlData := `
	<api context="/test" name="TestAPI">
		<cors enabled="true" 
			allow-origins="https://example.com" 
			unexpected-attribute="should-be-ignored">
		</cors>
		<resource methods="GET" uri-template="/resource1">
			<inSequence>
				<log level="full"/>
			</inSequence>
		</resource>
	</api>`

	position := artifacts.Position{
		FileName: "testfile.xml",
		LineNo:   1,
	}

	api := &API{}
	result, err := api.Unmarshal(xmlData, position)
	assert.NoError(t, err)

	// Verify CORS configuration was properly parsed despite unexpected attributes
	assert.True(t, result.CORSConfig.Enabled)
	assert.Equal(t, []string{"https://example.com"}, result.CORSConfig.AllowOrigins)

	// Verify other defaults are still there
	assert.Equal(t, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}, result.CORSConfig.AllowMethods)
	assert.Equal(t, []string{"Origin", "Content-Type", "Accept", "Authorization"}, result.CORSConfig.AllowHeaders)
}