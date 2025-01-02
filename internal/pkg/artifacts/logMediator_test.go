package artifacts

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/apache/synapse-go/internal/pkg/synapsecontext"
)

func TestLogMediator_Execute(t *testing.T) {
	tests := []struct {
		category string
		message  string
		wantErr  bool
	}{
		{"DEBUG", "Debug message", false},
		{"INFO", "Info message", false},
		{"ERROR", "Error message", true},
		{"", "Without category message", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			logMediator := &LogMediator{
				Category: tt.category,
				Message:  tt.message,
			}
			context := &synapsecontext.SynapseContext{}
			_, err := logMediator.Execute(context)
			if (err != nil) != tt.wantErr {
				t.Errorf("LogMediator.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLogMediator_UnmarshalArtifact(t *testing.T) {
	tests := []struct {
		name    string
		xmlData string
		wantErr bool
	}{
		{"Valid log mediator", `<log category="DEBUG"><message>Debug message</message></log>`, false},
		{"Invalid log mediator", `<log category="INVALID"><message>Debug message</message></log>`, true},
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
			err := logMediator.UnmarshalArtifact(decoder, startElement, ArtifactPosition{})
			if (err != nil) != tt.wantErr {
				t.Errorf("LogMediator.UnmarshalArtifact() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
