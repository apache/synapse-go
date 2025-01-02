package artifacts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalArtifact(t *testing.T) {
	xmlContent := `<api context="/test" name="TestAPI">
		<resource methods="GET" uri-template="/test">
			<inSequence>
				<log category="INFO"><message>Debug message</message></log>
			</inSequence>
			<faultSequence>
			</faultSequence>
		</resource>
		<resource methods="POST" uri-template="/test2">
			<inSequence>
				<log category="DEBUG"><message>Debug message</message></log>
			</inSequence>
			<faultSequence>
				<log category="INFO"><message>Debug message</message></log>
			</faultSequence>
		</resource>
	</api>`
	position := ArtifactPosition{FileName: "test.xml"}
	api := API{}
	err := api.UnmarshalArtifact(xmlContent, position)
	assert.NoError(t, err)
	assert.Equal(t, "/test", api.Context)
	assert.Equal(t, "TestAPI", api.Name)
	assert.Len(t, api.Resources, 2)
	assert.Equal(t, "GET", api.Resources[0].Methods)
	assert.Equal(t, "/test", api.Resources[0].URITemplate)
	assert.Equal(t, "TestAPI->/test->inSequence", api.Resources[0].InSequence.MediatorList[0].(*LogMediator).Position.Hierarchy)
	assert.Equal(t, 4, api.Resources[0].InSequence.MediatorList[0].(*LogMediator).Position.LineNo)
	assert.Equal(t, "POST", api.Resources[1].Methods)
	assert.Equal(t, "/test2", api.Resources[1].URITemplate)
	assert.Equal(t, "TestAPI->/test2->inSequence", api.Resources[1].InSequence.MediatorList[0].(*LogMediator).Position.Hierarchy)
	assert.Equal(t, 14, api.Resources[1].FaultSequence.MediatorList[0].(*LogMediator).Position.LineNo)
}

func TestUnmarshalArtifact_InvalidXML(t *testing.T) {
	xmlContent := `<api context="/test" name="TestAPI">
		<resource methods="GET" uri-template="/test">
			<inSequence><log category="INVALID"><message>Debug message</message></log></inSequence>
			<faultSequence></faultSequence>
		</resource>
	`
	position := ArtifactPosition{FileName: "test.xml"}
	api := API{}
	err := api.UnmarshalArtifact(xmlContent, position)
	assert.Error(t, err)
}

func TestUnmarshalArtifact_UnknownElement(t *testing.T) {
	xmlContent := `<api context="/test" name="TestAPI">
		<resource methods="GET" uri-template="/test">
			<unknownElement></unknownElement>
		</resource>
	</api>`
	position := ArtifactPosition{FileName: "test.xml"}
	api := API{}
	err := api.UnmarshalArtifact(xmlContent, position)
	assert.NoError(t, err)
	assert.Equal(t, "/test", api.Context)
	assert.Equal(t, "TestAPI", api.Name)
	assert.Len(t, api.Resources, 1)
	assert.Equal(t, "GET", api.Resources[0].Methods)
	assert.Equal(t, "/test", api.Resources[0].URITemplate)
}
