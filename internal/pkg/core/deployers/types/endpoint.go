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
	"fmt"
	"strings"

	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
)

type Endpoint struct {
	Name        string               `xml:"name,attr"`
	EndpointUrl   artifacts.EndpointUrl `xml:"http"`
	Position    artifacts.Position
}

type EndpointUrl struct {
	Method       string  `xml:"method,attr"`
	URITemplate   string `xml:"uri-template,attr"`
}

func (endpoint *Endpoint) Unmarshal(xmlData string, position artifacts.Position) (artifacts.Endpoint, error) {
	decoder := xml.NewDecoder(strings.NewReader(xmlData))
	newEndpoint := artifacts.Endpoint{}
	newEndpoint.Position = position
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		switch elem := token.(type) {
		case xml.StartElement:
			switch elem.Name.Local {
			case "endpoint":
				for _, attr := range elem.Attr {
					switch attr.Name.Local {
					case "name":
						newEndpoint.Name = attr.Value
						newEndpoint.Position.Hierarchy = attr.Value
					}
				}
			case "http":
				var endpointUrl = EndpointUrl{}
				res, err := endpointUrl.Unmarshal(decoder, elem, newEndpoint.Position)
				if err != nil {
					return artifacts.Endpoint{}, err
				}
				newEndpoint.EndpointUrl = res
			default:
				// Skip unknown elements
				if err := decoder.Skip(); err != nil {
					return artifacts.Endpoint{}, err
				}
			}
		case xml.EndElement:
			// Stop when the </endpoint> tag is encountered
			break
		}
	}

	if newEndpoint.Name == "" {
		return artifacts.Endpoint{}, fmt.Errorf("Endpoint name is required")
	}

	return newEndpoint, nil
}

// implements custom unmarshaling for EndpointUrl
func (endpointUrl *EndpointUrl) Unmarshal(decoder *xml.Decoder, start xml.StartElement, position artifacts.Position) (artifacts.EndpointUrl, error) {
	// Extract attributes from the <http> element
	res := artifacts.EndpointUrl{}
	var method string
	var uriTemplate string

	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "method":
			method= attr.Value
		case "uri-template":
			uriTemplate = attr.Value
		}
	}
	res.Method = method
	res.URITemplate = uriTemplate
	return res, nil
}
