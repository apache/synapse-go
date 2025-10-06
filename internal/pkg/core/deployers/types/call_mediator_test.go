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

func TestCallMediator_Unmarshal(t *testing.T) {
	tests := []struct {
		name          string
		xmlData       string
		position      artifacts.Position
		wantErr       bool
		wantEndpoint  string
		wantHierarchy string
	}{
		{
			name:          "Valid call mediator with endpoint",
			xmlData:       `<call><endpoint key="testEndpoint"/></call>`,
			position:      artifacts.Position{FileName: "test.xml", LineNo: 10, Hierarchy: "sequence"},
			wantErr:       false,
			wantEndpoint:  "testEndpoint",
			wantHierarchy: "sequence->call",
		},
		{
			name:          "Valid call mediator without endpoint",
			xmlData:       `<call></call>`,
			position:      artifacts.Position{FileName: "test.xml", LineNo: 15, Hierarchy: "sequence"},
			wantErr:       false,
			wantEndpoint:  "",
			wantHierarchy: "sequence->call",
		},
		{
			name:          "Empty call mediator",
			xmlData:       `<call/>`,
			position:      artifacts.Position{FileName: "test.xml", LineNo: 20, Hierarchy: "sequence"},
			wantErr:       false,
			wantEndpoint:  "",
			wantHierarchy: "sequence->call",
		},
		{
			name:          "Invalid XML structure",
			xmlData:       `<call><endpoint key="testEndpoint"</call>`,
			position:      artifacts.Position{FileName: "test.xml", LineNo: 25, Hierarchy: "sequence"},
			wantErr:       true,
			wantEndpoint:  "",
			wantHierarchy: "",
		},
		{
			name:          "Call with empty endpoint key",
			xmlData:       `<call><endpoint key=""/></call>`,
			position:      artifacts.Position{FileName: "test.xml", LineNo: 30, Hierarchy: "sequence"},
			wantErr:       false,
			wantEndpoint:  "",
			wantHierarchy: "sequence->call",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := xml.NewDecoder(strings.NewReader(tt.xmlData))
			callMediator := &CallMediator{}
			token, _ := decoder.Token()
			startElement, ok := token.(xml.StartElement)
			if !ok {
				t.Errorf("Expected xml.StartElement but got %T", token)
				return
			}

			// Call the Unmarshal method
			mediator, err := callMediator.Unmarshal(decoder, startElement, tt.position)

			// Check error expectations
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)

				// Type assertion
				callMediatorResult, ok := mediator.(artifacts.CallMediator)
				if !ok {
					t.Errorf("Expected artifacts.CallMediator but got %T", mediator)
					return
				}

				// Check specific field values
				assert.Equal(t, tt.wantEndpoint, callMediatorResult.EndpointRef, "EndpointRef should match")
				assert.Equal(t, tt.wantHierarchy, callMediatorResult.Position.Hierarchy, "Hierarchy should match")
				assert.Equal(t, tt.position.FileName, callMediatorResult.Position.FileName, "FileName should match")
				assert.Equal(t, tt.position.LineNo, callMediatorResult.Position.LineNo, "LineNo should match")
			}
		})
	}
}
