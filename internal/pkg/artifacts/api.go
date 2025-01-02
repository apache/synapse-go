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
	"net/http"
	"strings"

	"github.com/apache/synapse-go/internal/pkg/synapsecontext"
)

type Resource struct {
	Methods       string   `xml:"methods,attr"`
	URITemplate   string   `xml:"uri-template,attr"`
	InSequence    Sequence `xml:"inSequence"`
	FaultSequence Sequence `xml:"faultSequence"`
}

type API struct {
	Context   string     `xml:"context,attr"`
	Name      string     `xml:"name,attr"`
	Resources []Resource `xml:"resource"`
	Position  ArtifactPosition
}

func (api *API) UnmarshalArtifact(fileContent string, position ArtifactPosition) error {
	decoder := xml.NewDecoder(strings.NewReader(fileContent))
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	start, ok := token.(xml.StartElement)
	if !ok {
		return errors.New("error in unmarshalling API in " + position.FileName)
	}
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "context":
			api.Context = attr.Value
		case "name":
			api.Name = attr.Value
		}
	}
	api.Position = position
	api.Position.Hierarchy = api.Name

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch elem := token.(type) {
		case xml.StartElement:
			switch elem.Name.Local {
			case "resource":
				var resource = Resource{}
				if err := resource.UnmarshalArtifact(decoder, elem, api.Position); err != nil {
					return err
				}
				api.Resources = append(api.Resources, resource)
			default:
				// Skip unknown elements
				if err := decoder.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			// Stop when the </api> tag is encountered
			if elem.Name.Local == start.Name.Local {
				return nil
			}
		}
	}
	return nil
}

// implements custom unmarshaling for Resource
func (r *Resource) UnmarshalArtifact(decoder *xml.Decoder, start xml.StartElement, position ArtifactPosition) error {
	// Extract attributes from the <resource> element
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "methods":
			r.Methods = attr.Value
		case "uri-template":
			r.URITemplate = attr.Value
		}
	}

	// Process child elements
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch elem := token.(type) {
		case xml.StartElement:
			switch elem.Name.Local {
			case "inSequence":
				line, _ := decoder.InputPos()
				r.InSequence = Sequence{}
				if err := r.InSequence.UnmarshalArtifact(decoder, elem, ArtifactPosition{
					FileName:  position.FileName,
					LineNo:    line,
					Hierarchy: position.Hierarchy + "->" + r.URITemplate + "->" + "inSequence",
				}); err != nil {
					return err
				}
			case "faultSequence":
				line, _ := decoder.InputPos()
				r.FaultSequence = Sequence{}
				if err := r.FaultSequence.UnmarshalArtifact(decoder, elem, ArtifactPosition{
					FileName:  position.FileName,
					LineNo:    line,
					Hierarchy: position.Hierarchy + "->" + r.URITemplate + "->" + "faultSequence",
				}); err != nil {
					return err
				}
			default:
				// Skip unknown elements
				if err := decoder.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			// Stop when the </resource> tag is encountered
			if elem.Name.Local == start.Name.Local {
				return nil
			}
		}
	}
	return nil
}

func (resource *Resource) DispatchResource(w http.ResponseWriter, request *http.Request) {
	// Read transport headers
	var headers = make(map[string]string)
	for name, values := range request.Header {
		headers[name] = values[0]
	}

	// Creating the mssage context
	var context = synapsecontext.SynapseContext{
		Properties: make(map[string]string),
		Message:    synapsecontext.Message{},
		Headers:    headers,
	}

	// Execute the in-sequence
	resource.InSequence.Execute(&context)
}
