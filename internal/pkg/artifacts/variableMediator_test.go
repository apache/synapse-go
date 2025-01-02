package artifacts

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/apache/synapse-go/internal/pkg/synapsecontext"
)

func TestVariableMediator_Execute(t *testing.T) {
	context := &synapsecontext.SynapseContext{
		Properties: make(map[string]string),
	}
	variable := &VariableMediator{
		Name:  "testVariable",
		Type:  "STRING",
		Value: "testValue",
	}

	result := variable.Execute(context)
	if !result {
		t.Errorf("Expected Execute to return true, got %v", result)
	}

	if context.Properties["testVariable"] != "testValue" {
		t.Errorf("Expected context property 'testVariable' to be 'testValue', got %v", context.Properties["testVariable"])
	}
}

func TestVariableMediator_UnmarshalArtifact_ValidType(t *testing.T) {
	tests := []struct {
		name    string
		xmlData string
		wantErr bool
	}{
		{"Valid variable mediator", `<variable type="STRING" value="Hello World"></variable>`, false},
		{"Valid variable mediator", `<variable type="STRING" value="Hello World"/>`, false},
		{"Invalid variable mediator", `<variable type="ABC" value="Hello World"/>`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := xml.NewDecoder(strings.NewReader(tt.xmlData))
			variableMediator := &VariableMediator{}
			token, _ := decoder.Token()
			startElement, ok := token.(xml.StartElement)
			if !ok {
				t.Errorf("Expected xml.StartElement but got %T", token)
				return
			}
			err := variableMediator.UnmarshalArtifact(decoder, startElement, ArtifactPosition{})
			if (err != nil) != tt.wantErr {
				t.Errorf("VariableMediator.UnmarshalArtifact() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
