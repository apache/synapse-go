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

package artifacts

import (
	"encoding/xml"
	"errors"
	"strconv"

	"github.com/apache/synapse-go/internal/pkg/synapsecontext"
)

type VariableType string

const (
	String  VariableType = "STRING"
	Boolean VariableType = "BOOLEAN"
	Integer VariableType = "INTEGER"
	Long    VariableType = "LONG"
	Double  VariableType = "DOUBLE"
	Json    VariableType = "JSON"
	Xml     VariableType = "XML"
)

type VariableMediator struct {
	Name     string `xml:"name,attr"`
	Type     string `xml:"type,attr"`
	Value    string `xml:"value,attr"`
	Position ArtifactPosition
}

func (variable *VariableMediator) Execute(context *synapsecontext.SynapseContext) bool {
	context.Properties[variable.Name] = variable.Value
	return true
}

func (variable *VariableMediator) UnmarshalArtifact(d *xml.Decoder, start xml.StartElement, position ArtifactPosition) error {
	if err := d.DecodeElement(variable, &start); err != nil {
		return errors.New("error in unmarshalling variable mediator in " + position.FileName + " at line " + strconv.Itoa(position.LineNo))
	}
	validTypes := map[VariableType]bool{
		String:  true,
		Boolean: true,
		Integer: true,
		Long:    true,
		Double:  true,
		Json:    true,
		Xml:     true,
	}

	if _, isValid := validTypes[VariableType(variable.Type)]; !isValid {
		return errors.New("invalid variable type " + variable.Type + " in " + position.FileName + " at line " + strconv.Itoa(position.LineNo))
	}
	variable.Position = position
	return nil
}
