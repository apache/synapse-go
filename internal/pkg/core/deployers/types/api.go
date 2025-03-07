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

	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
)

type Resource struct {
	Methods       string             `xml:"methods,attr"`
	URITemplate   string             `xml:"uri-template,attr"`
	InSequence    artifacts.Sequence `xml:"inSequence"`
	FaultSequence artifacts.Sequence `xml:"faultSequence"`
}

type API struct {
	Context   string               `xml:"context,attr"`
	Name      string               `xml:"name,attr"`
	Resources []artifacts.Resource `xml:"resource"`
	Position  artifacts.Position
}

func (api *API) Unmarshal(xmlData string, position artifacts.Position) (artifacts.API, error) {
	decoder := xml.NewDecoder(strings.NewReader(xmlData))
	newAPI := artifacts.API{}
	newAPI.Position = position
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		switch elem := token.(type) {
		case xml.StartElement:
			switch elem.Name.Local {
			case "api":
				for _, attr := range elem.Attr {
					switch attr.Name.Local {
					case "context":
						newAPI.Context = attr.Value
					case "name":
						newAPI.Name = attr.Value
						newAPI.Position.Hierarchy = attr.Value
					}
				}
			case "resource":
				var resource = Resource{}
				res, err := resource.Unmarshal(decoder, elem, newAPI.Position)
				if err != nil {
					return artifacts.API{}, err
				}
				newAPI.Resources = append(api.Resources, res)
			default:
				// Skip unknown elements
				if err := decoder.Skip(); err != nil {
					return artifacts.API{}, err
				}
			}
		case xml.EndElement:
			// Stop when the </api> tag is encountered
			break
		}
	}
	return newAPI, nil
}

// implements custom unmarshaling for Resource
func (r *Resource) Unmarshal(decoder *xml.Decoder, start xml.StartElement, position artifacts.Position) (artifacts.Resource, error) {
	// Extract attributes from the <resource> element
	res := artifacts.Resource{}
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "methods":
			res.Methods = attr.Value
		case "uri-template":
			res.URITemplate = attr.Value
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
			case "inSequence", "faultSequence":
				seq, err := r.decodeSequence(decoder, position, elem.Name.Local, res)
				if err != nil {
					return artifacts.Resource{}, err
				}
				if elem.Name.Local == "inSequence" {
					res.InSequence = seq
				} else {
					res.FaultSequence = seq
				}
			default:
				// Skip unknown elements
				if err := decoder.Skip(); err != nil {
					return artifacts.Resource{}, err
				}
			}
		case xml.EndElement:
			// Stop when the </resource> tag is encountered
			break
		}
	}
	return res, nil
}

func (r *Resource) decodeSequence(decoder *xml.Decoder, position artifacts.Position, sequenceType string, res artifacts.Resource) (artifacts.Sequence, error) {

	for {
		token, err := decoder.Token()
		if err != nil {
			return artifacts.Sequence{}, err
		}
		if startElem, ok := token.(xml.StartElement); ok && startElem.Name.Local == "sequence" {
			break
		}
	}
	line, _ := decoder.InputPos()
	decodeSeq := Sequence{}
	seq, err := decodeSeq.unmarshal(decoder, artifacts.Position{
		FileName:  position.FileName,
		LineNo:    line,
		Hierarchy: position.Hierarchy + "->" + res.URITemplate + "->" + sequenceType,
	})
	if err != nil {
		return artifacts.Sequence{}, err
	}
	return seq, nil
}
