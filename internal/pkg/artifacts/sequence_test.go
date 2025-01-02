package artifacts

import (
	"bytes"
	"encoding/xml"
	"testing"

	"github.com/apache/synapse-go/internal/pkg/synapsecontext"
	"github.com/stretchr/testify/assert"
)

type MockMediator struct {
	ExecuteResult bool
	ExecuteError  error
}

func (m *MockMediator) Execute(context *synapsecontext.SynapseContext) (bool, error) {
	return m.ExecuteResult, m.ExecuteError
}

func TestBaseSequence_Execute(t *testing.T) {
	tests := []struct {
		name       string
		mediators  []Mediator
		wantResult bool
	}{
		{
			name: "All mediators execute successfully",
			mediators: []Mediator{
				&MockMediator{ExecuteResult: true},
				&MockMediator{ExecuteResult: true},
			},
			wantResult: true,
		},
		{
			name: "One mediator fails",
			mediators: []Mediator{
				&MockMediator{ExecuteResult: true},
				&MockMediator{ExecuteResult: false},
			},
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seq := &BaseSequence{
				MediatorList: tt.mediators,
			}
			context := &synapsecontext.SynapseContext{}
			if got := seq.Execute(context); got != tt.wantResult {
				t.Errorf("BaseSequence.Execute() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func TestBaseSequence_UnmarshalArtifact(t *testing.T) {
	xmlData := `<sequence><log/></sequence>`
	decoder := xml.NewDecoder(bytes.NewBufferString(xmlData))
	startElement := xml.StartElement{Name: xml.Name{Local: "sequence"}}
	position := ArtifactPosition{LineNo: 1, FileName: "test.xml", Hierarchy: "root"}

	seq := &Sequence{}
	err := seq.UnmarshalArtifact(decoder, startElement, position)
	if err != nil {
		t.Fatalf("UnmarshalArtifact() error = %v", err)
	}

	if len(seq.MediatorList) != 1 {
		t.Errorf("Expected 1 mediator, got %d", len(seq.MediatorList))
	}
}

func TestInvalid_UnmashalArtifact(t *testing.T) {
	xmlData := `<sequence>
		<log category="INVALID">
			<message>Debug message</message>
		</log>
	</sequence>`
	decoder := xml.NewDecoder(bytes.NewBufferString(xmlData))
	startElement := xml.StartElement{Name: xml.Name{Local: "sequence"}}
	position := ArtifactPosition{LineNo: 1, FileName: "test.xml", Hierarchy: "sequence1"}

	seq := &Sequence{}
	err := seq.UnmarshalArtifact(decoder, startElement, position)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid log category")
	assert.Contains(t, err.Error(), "test.xml")
	assert.Contains(t, err.Error(), "at line 2")
}

func TestNamedSequence_UnmarshalArtifact(t *testing.T) {
	xmlContent := `<sequence name="inboundSeq" trace="disable" xmlns="http://ws.apache.org/ns/synapse">
    	<log category="INFO">
        	<message>message from inbound</message>
    	</log>
	</sequence>`
	position := ArtifactPosition{FileName: "inboundSeq.xml"}
	seq := NamedSequence{}
	err := seq.UnmarshalArtifact(xmlContent, position)
	assert.NoError(t, err)
	assert.Equal(t, "inboundSeq", seq.Name)
	assert.Equal(t, "INFO", seq.MediatorList[0].(*LogMediator).Category)
	assert.Equal(t, "inboundSeq", seq.MediatorList[0].(*LogMediator).Position.Hierarchy)
	assert.Equal(t, 2, seq.MediatorList[0].(*LogMediator).Position.LineNo)
}
