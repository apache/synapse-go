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
	"encoding/xml"
	"strings"
	"testing"

	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
	"github.com/stretchr/testify/assert"
)

func TestEndpoint_Unmarshal_Success(t *testing.T) {
	// Test data: Valid endpoint XML with name and HTTP details
	xmlData := `<endpoint name="stockQuoteEndpoint">
		<http method="GET" uri-template="https://api.example.com/stock/{symbol}"/>
	</endpoint>`

	// Create position for testing
	position := artifacts.Position{
		LineNo:    10,
		FileName:  "test_endpoint.xml",
		Hierarchy: "",
	}

	// Create endpoint instance and unmarshal
	endpoint := &Endpoint{}
	result, err := endpoint.Unmarshal(xmlData, position)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "stockQuoteEndpoint", result.Name)
	assert.Equal(t, "stockQuoteEndpoint", result.Position.Hierarchy)
	assert.Equal(t, "GET", result.EndpointUrl.Method)
	assert.Equal(t, "https://api.example.com/stock/{symbol}", result.EndpointUrl.URITemplate)
	assert.Equal(t, position.LineNo, result.Position.LineNo)
	assert.Equal(t, position.FileName, result.Position.FileName)
}

func TestEndpoint_Unmarshal_MissingName(t *testing.T) {
	// Test data: Missing required name attribute
	xmlData := `<endpoint>
		<http method="GET" uri-template="https://api.example.com/stock/{symbol}"/>
	</endpoint>`

	position := artifacts.Position{
		LineNo:    20,
		FileName:  "test_endpoint.xml",
		Hierarchy: "",
	}

	endpoint := &Endpoint{}
	_, err := endpoint.Unmarshal(xmlData, position)

	// Should return an error for missing name
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Endpoint name is required")
}

func TestEndpoint_Unmarshal_HttpMethods(t *testing.T) {
	// Test different HTTP methods with table-driven tests
	testCases := []struct {
		name     string
		method   string
		template string
	}{
		{"getEndpoint", "GET", "https://api.example.com/get"},
		{"postEndpoint", "POST", "https://api.example.com/post"},
		{"putEndpoint", "PUT", "https://api.example.com/put"},
		{"deleteEndpoint", "DELETE", "https://api.example.com/delete"},
		{"patchEndpoint", "PATCH", "https://api.example.com/patch"},
		{"headEndpoint", "HEAD", "https://api.example.com/head"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			xmlData := `<endpoint name="` + tc.name + `">
				<http method="` + tc.method + `" uri-template="` + tc.template + `"/>
			</endpoint>`

			position := artifacts.Position{
				LineNo:    40,
				FileName:  "test_endpoint.xml",
				Hierarchy: "",
			}

			endpoint := &Endpoint{}
			result, err := endpoint.Unmarshal(xmlData, position)

			assert.NoError(t, err)
			assert.Equal(t, tc.name, result.Name)
			assert.Equal(t, tc.method, result.EndpointUrl.Method)
			assert.Equal(t, tc.template, result.EndpointUrl.URITemplate)
		})
	}
}

func TestEndpoint_Unmarshal_EmptyHttpAttributes(t *testing.T) {
	// Test empty HTTP attributes
	xmlData := `<endpoint name="emptyAttributesEndpoint">
		<http method="" uri-template=""/>
	</endpoint>`

	position := artifacts.Position{
		LineNo:    50,
		FileName:  "test_endpoint.xml",
		Hierarchy: "",
	}

	endpoint := &Endpoint{}
	result, err := endpoint.Unmarshal(xmlData, position)

	assert.NoError(t, err)
	assert.Equal(t, "emptyAttributesEndpoint", result.Name)
	assert.Equal(t, "", result.EndpointUrl.Method)
	assert.Equal(t, "", result.EndpointUrl.URITemplate)
}

func TestEndpointUrl_Unmarshal_Direct(t *testing.T) {
	// Test the EndpointUrl.Unmarshal method directly
	xmlContent := `<http method="POST" uri-template="https://api.test.com/users"/>`
	decoder := xml.NewDecoder(strings.NewReader(xmlContent))

	// Get the start element
	var startElement xml.StartElement
	for {
		token, _ := decoder.Token()
		if token == nil {
			t.Fatal("Could not find start element in XML")
			break
		}
		if se, ok := token.(xml.StartElement); ok {
			startElement = se
			break
		}
	}

	position := artifacts.Position{
		LineNo:    60,
		FileName:  "test_endpoint.xml",
		Hierarchy: "testEndpoint",
	}

	endpointUrl := &EndpointUrl{}
	result, err := endpointUrl.Unmarshal(decoder, startElement, position)

	assert.NoError(t, err)
	assert.Equal(t, "POST", result.Method)
	assert.Equal(t, "https://api.test.com/users", result.URITemplate)
}

func TestEndpoint_Unmarshal_MultipleElements(t *testing.T) {
	// Test with multiple elements including some that should be skipped
	xmlData := `<endpoint name="complexEndpoint">
		<unknown>Should be skipped</unknown>
		<http method="PUT" uri-template="https://api.example.com/resource"/>
		<anotherUnknown>Should also be skipped</anotherUnknown>
	</endpoint>`

	position := artifacts.Position{
		LineNo:    70,
		FileName:  "test_endpoint.xml",
		Hierarchy: "",
	}

	endpoint := &Endpoint{}
	result, err := endpoint.Unmarshal(xmlData, position)

	assert.NoError(t, err)
	assert.Equal(t, "complexEndpoint", result.Name)
	assert.Equal(t, "PUT", result.EndpointUrl.Method)
	assert.Equal(t, "https://api.example.com/resource", result.EndpointUrl.URITemplate)
}
