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

func TestUnmarshalMediators(t *testing.T) {
	xmlData := `<sequence>
		<log category="DEBUG">
			<message>Debug message</message>
		</log>
	</sequence>`

	decoder := xml.NewDecoder(strings.NewReader(xmlData))
	position := artifacts.Position{FileName: "testfile.xml"}

	sequence := &Sequence{}
	newSeq, err := sequence.unmarshal(decoder, position)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(newSeq.MediatorList) != 1 {
		t.Fatalf("Expected 1 mediator, got %d", len(newSeq.MediatorList))
	}

	logMediator := newSeq.MediatorList[0]
	if logMediator, ok := logMediator.(artifacts.LogMediator); ok {
		assert.Equal(t, "sequence->log", logMediator.Position.Hierarchy)
		assert.Equal(t, 2, logMediator.Position.LineNo)
	} else {
		t.Fatalf("Expected LogMediator, got %T", newSeq.MediatorList[0])
	}
}

func TestUnmarshalMediatorsWithEmptySequence(t *testing.T) {
	xmlData := `<sequence></sequence>`

	decoder := xml.NewDecoder(strings.NewReader(xmlData))
	position := artifacts.Position{FileName: "testfile.xml"}

	sequence := &Sequence{}
	newSeq, err := sequence.unmarshal(decoder, position)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(newSeq.MediatorList) != 0 {
		t.Fatalf("Expected 0 mediators, got %d", len(newSeq.MediatorList))
	}
}

func TestUnmarshalMediatorsWithInvalidXML(t *testing.T) {
	xmlData := `<sequence><log></sequence>`

	decoder := xml.NewDecoder(strings.NewReader(xmlData))
	position := artifacts.Position{FileName: "testfile.xml"}

	sequence := &Sequence{}
	_, err := sequence.unmarshal(decoder, position)
	assert.NotNil(t, err)
}
