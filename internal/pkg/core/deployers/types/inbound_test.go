/*
Copyright 2025 The Synapse Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package types

import (
	"testing"

	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
)

func TestUnmarshal(t *testing.T) {
	xmlData := `<inbound name="testInbound" sequence="testSequence" protocol="testProtocol" suspend="testSuspend" onError="testOnError">
					<parameters>
						<parameter name="param1">value1</parameter>
						<parameter name="param2">value2</parameter>
					</parameters>
				</inbound>`
	position := artifacts.Position{Hierarchy: "testHierarchy"}

	inbound := &Inbound{}
	result, err := inbound.Unmarshal(xmlData, position)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Name != "testInbound" {
		t.Errorf("Expected Name to be 'testInbound', got %s", result.Name)
	}
	if result.Sequence != "testSequence" {
		t.Errorf("Expected Sequence to be 'testSequence', got %s", result.Sequence)
	}
	if result.Protocol != "testProtocol" {
		t.Errorf("Expected Protocol to be 'testProtocol', got %s", result.Protocol)
	}
	if result.Suspend != "testSuspend" {
		t.Errorf("Expected Suspend to be 'testSuspend', got %s", result.Suspend)
	}
	if result.OnError != "testOnError" {
		t.Errorf("Expected OnError to be 'testOnError', got %s", result.OnError)
	}
	if len(result.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(result.Parameters))
	}
	if result.Parameters[0].Name != "param1" || result.Parameters[0].Value != "value1" {
		t.Errorf("Expected first parameter to be (param1, value1), got (%s, %s)", result.Parameters[0].Name, result.Parameters[0].Value)
	}
	if result.Parameters[1].Name != "param2" || result.Parameters[1].Value != "value2" {
		t.Errorf("Expected second parameter to be (param2, value2), got (%s, %s)", result.Parameters[1].Name, result.Parameters[1].Value)
	}
	if result.Position.Hierarchy != "testHierarchy" {
		t.Errorf("Expected Hierarchy to be 'testHierarchy', got %s", result.Position.Hierarchy)
	}
}

func TestUnmarshalEmptyParameters(t *testing.T) {
	xmlData := `<inbound name="testInbound" sequence="testSequence" protocol="testProtocol" suspend="testSuspend" onError="testOnError">
					<parameters>
					</parameters>
				</inbound>`
	position := artifacts.Position{Hierarchy: "testHierarchy"}

	inbound := &Inbound{}
	result, err := inbound.Unmarshal(xmlData, position)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Parameters) != 0 {
		t.Errorf("Expected 0 parameters, got %d", len(result.Parameters))
	}
}

func TestUnmarshalNoParameters(t *testing.T) {
	xmlData := `<inbound name="testInbound" sequence="testSequence" protocol="testProtocol" suspend="testSuspend" onError="testOnError">
				</inbound>`
	position := artifacts.Position{Hierarchy: "testHierarchy"}

	inbound := &Inbound{}
	result, err := inbound.Unmarshal(xmlData, position)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Parameters) != 0 {
		t.Errorf("Expected 0 parameters, got %d", len(result.Parameters))
	}
}
