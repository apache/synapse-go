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
