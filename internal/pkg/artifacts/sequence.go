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
	"fmt"
	"strings"

	"github.com/apache/synapse-go/internal/pkg/synapsecontext"
)

type BaseSequence struct {
	MediatorList []Mediator
	Position     ArtifactPosition
}

// Collection of mediators resides in an unnamed sequence
type Sequence struct {
	BaseSequence
}

type NamedSequence struct {
	Name string
	BaseSequence
}

func (v *BaseSequence) Execute(context *synapsecontext.SynapseContext) bool {
	for _, mediator := range v.MediatorList {
		result, err := mediator.Execute(context)
		if !result {
			return false
		}
		if err != nil {
			fmt.Println(err)
		}
	}
	return true
}

func (seq *BaseSequence) UnmarshalArtifact(decoder *xml.Decoder, start xml.StartElement, position ArtifactPosition) error {
	seq.Position = position
	mediatorList, err := seq.unmarshalMediators(decoder, start)
	if err != nil {
		return err
	}
	seq.MediatorList = mediatorList
	return nil
}

func (seq *Sequence) UnmarshalArtifact(decoder *xml.Decoder, start xml.StartElement, position ArtifactPosition) error {
	return seq.BaseSequence.UnmarshalArtifact(decoder, start, position)
}

func (seq *NamedSequence) UnmarshalArtifact(fileContent string, position ArtifactPosition) error {
	decoder := xml.NewDecoder(strings.NewReader(fileContent))
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	// Skip the xml declaration
	if token.(xml.ProcInst).Target == "xml" {
		token, _ = decoder.Token()
	}
	// skip all token until the start element
	for {
		if start, ok := token.(xml.StartElement); !ok {
			token, _ = decoder.Token()
		} else {
			if start.Name.Local == "sequence" {
				break
			} else {
				return errors.New("error in unmarshalling Sequence in " + position.FileName)
			}
		}
	}
	start, ok := token.(xml.StartElement)
	if !ok {
		return errors.New("error in unmarshalling Sequence in " + position.FileName)
	}
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "name":
			seq.Name = attr.Value
		}
	}
	seq.Position = position
	seq.Position.Hierarchy = seq.Name
	return seq.BaseSequence.UnmarshalArtifact(decoder, start, seq.Position)
}

func (seq *BaseSequence) unmarshalMediators(decoder *xml.Decoder, start xml.StartElement) ([]Mediator, error) {
	var mediatorList []Mediator

OuterLoop:
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		line, _ := decoder.InputPos()
		switch element := token.(type) {
		case xml.StartElement:
			var mediator Mediator
			switch element.Name.Local {
			case "log":
				logMediator := &LogMediator{}
				position := ArtifactPosition{LineNo: line, FileName: seq.Position.FileName, Hierarchy: seq.Position.Hierarchy}
				if err := logMediator.UnmarshalArtifact(decoder, element, position); err != nil {
					return nil, err
				}
				mediator = logMediator
			}
			if mediator != nil {
				mediatorList = append(mediatorList, mediator)
			}
		case xml.EndElement:
			// Stop when the </sequence> tag is encountered
			if element.Name.Local == start.Name.Local {
				break OuterLoop
			}
		}
	}
	return mediatorList, nil
}
