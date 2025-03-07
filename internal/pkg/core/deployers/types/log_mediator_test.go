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

func TestLogMediator_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		xmlData string
		wantErr bool
	}{
		{"Valid log mediator", `<log category="DEBUG"><message>Debug message</message></log>`, false},
		{"Invalid log mediator", `<lo category="INVALID"><message>Debug message</message></logs>`, true},
		{"Valid with default category", `<log><message>Debug message</message></log>`, false},
		{"Valid minimal", `<log/>`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := xml.NewDecoder(strings.NewReader(tt.xmlData))
			logMediator := &LogMediator{}
			token, _ := decoder.Token()
			startElement, ok := token.(xml.StartElement)
			if !ok {
				t.Errorf("Expected xml.StartElement but got %T", token)
				return
			}
			mediator, err := logMediator.Unmarshal(decoder, startElement, artifacts.Position{})
			if tt.wantErr {
				assert.NotNil(t, err)
			} else if (err != nil) != tt.wantErr {
				t.Errorf("LogMediator.UnmarshalArtifact() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				logMediator, ok := mediator.(artifacts.LogMediator)
				if !ok {
					t.Errorf("Expected artifacts.LogMediator but got %T", mediator)
					return
				}
				assert.Equal(t, "->log", logMediator.Position.Hierarchy)
			}
		})
	}
}
