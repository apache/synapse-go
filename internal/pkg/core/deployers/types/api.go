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
	"regexp"
	"strings"

	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
)

// Regular expression to find path parameters in the format {paramName}
var pathParamRegex = regexp.MustCompile(`\{([^}]+)\}`)

type Resource struct {
	Methods       string                    `xml:"methods,attr"`
	URITemplate   artifacts.URITemplateInfo `xml:"uri-template,attr"`
	InSequence    artifacts.Sequence        `xml:"inSequence"`
	FaultSequence artifacts.Sequence        `xml:"faultSequence"`
}

type API struct {
	Context     string               `xml:"context,attr"`
	Name        string               `xml:"name,attr"`
	Version     string               `xml:"version,attr"`
	VersionType string               `xml:"version-type,attr"`
	Resources   []artifacts.Resource `xml:"resource"`
	Position    artifacts.Position
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
					case "version":
						newAPI.Version = attr.Value
					case "version-type":
						newAPI.VersionType = attr.Value
					}
				}
			case "resource":
				var resource = Resource{}
				res, err := resource.Unmarshal(decoder, elem, newAPI.Position)
				if err != nil {
					return artifacts.API{}, err
				}
				newAPI.Resources = append(newAPI.Resources, res)
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

	// Validation checks
	if newAPI.Context == "" {
		return artifacts.API{}, fmt.Errorf("API context is required")
	}

	// Check that context starts with '/'
	if newAPI.Context == "" || newAPI.Context[0] != '/' {
		return artifacts.API{}, fmt.Errorf("API context must begin with '/' character")
	}

	if newAPI.Name == "" {
		return artifacts.API{}, fmt.Errorf("API name is required")
	}

	// Check version and versionType consistency
	hasVersion := newAPI.Version != ""
	hasVersionType := newAPI.VersionType != ""

	if hasVersion != hasVersionType {
		return artifacts.API{}, fmt.Errorf("both version and version-type must be specified together")
	}

	// Validate versionType if specified
	if hasVersionType && newAPI.VersionType != "context" && newAPI.VersionType != "url" {
		return artifacts.API{}, fmt.Errorf("version-type must be either 'context' or 'url', got: %s", newAPI.VersionType)
	}

	return newAPI, nil
}

// implements custom unmarshaling for Resource
func (r *Resource) Unmarshal(decoder *xml.Decoder, start xml.StartElement, position artifacts.Position) (artifacts.Resource, error) {
	// Extract attributes from the <resource> element
	res := artifacts.Resource{}
	var methodsStr string
	var uriTemplate string

	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "methods":
			methodsStr = attr.Value
		case "uri-template":
			uriTemplate = attr.Value
		}
	}

	// Split the methods string into a slice (e.g., "GET POST PUT" -> ["GET", "POST", "PUT"])
	if methodsStr != "" {
		res.Methods = strings.Fields(methodsStr)
	}

	// Parse the URI template if provided
	if uriTemplate != "" {
		// Parse and validate the URI template
		parsedInfo, err := r.parseURITemplate(uriTemplate)
		if err != nil {
			return artifacts.Resource{}, fmt.Errorf("invalid URI template '%s': %w", uriTemplate, err)
		}
		// Store the parsed URI template info in the artifacts.Resource
		res.URITemplate = parsedInfo
	}

	// Process child elements - use a labeled loop for cleaner exiting
parsingLoop:
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
			if elem.Name.Local == "resource" {
				break parsingLoop
			}
		}
	}
	return res, nil
}

func (r *Resource) decodeSequence(decoder *xml.Decoder, position artifacts.Position, sequenceType string, res artifacts.Resource) (artifacts.Sequence, error) {
	line, _ := decoder.InputPos()

	position = artifacts.Position{
		FileName:  position.FileName,
		LineNo:    line,
		Hierarchy: position.Hierarchy + "->" + res.URITemplate.FullTemplate + "->" + sequenceType,
	}

	// Check if the next element is a sequence tag
	for {
		token, err := decoder.Token()
		if err != nil {
			return artifacts.Sequence{}, err
		}

		if startElem, ok := token.(xml.StartElement); ok {
			if startElem.Name.Local == "sequence" {
				// Handle nested sequence format
				decodeSeq := Sequence{}
				seq, err := decodeSeq.unmarshal(decoder, position)
				if err != nil {
					return artifacts.Sequence{}, err
				}
				return seq, nil
			} else {
				// Handle direct mediators format
				var mediatorList []artifacts.Mediator
				if position.Hierarchy == "" {
					position.Hierarchy = sequenceType
				}

				// Process the first element we found
				switch startElem.Name.Local {
				case "log":
					logMediator := LogMediator{}
					mediator, err := logMediator.Unmarshal(decoder, startElem, position)
					if err != nil {
						return artifacts.Sequence{}, err
					}
					mediatorList = append(mediatorList, mediator)
				case "call":
					callMediator := CallMediator{}
					mediator, err := callMediator.Unmarshal(decoder, startElem, position)
					if err != nil {
						return artifacts.Sequence{}, err
					}
					mediatorList = append(mediatorList, mediator)
				}
				// Continue processing other elements
			OuterLoop:
				for {
					token, err := decoder.Token()
					if err != nil {
						break
					}
					line, _ := decoder.InputPos()
					position := artifacts.Position{LineNo: line, FileName: position.FileName, Hierarchy: position.Hierarchy}
					switch element := token.(type) {
					case xml.StartElement:
						switch element.Name.Local {
						case "log":
							logMediator := LogMediator{}
							mediator, err := logMediator.Unmarshal(decoder, element, position)
							if err != nil {
								return artifacts.Sequence{}, err
							}
							mediatorList = append(mediatorList, mediator)
						case "call":
							callMediator := CallMediator{}
							mediator, err := callMediator.Unmarshal(decoder, element, position)
							if err != nil {
								return artifacts.Sequence{}, err
							}
							mediatorList = append(mediatorList, mediator)
						}
					case xml.EndElement:
						// Stop when the </sequence> tag is encountered
						if element.Name.Local == sequenceType {
							break OuterLoop
						}
					}
				}
				return artifacts.Sequence{MediatorList: mediatorList, Position: position}, nil
			}
		}
	}
}

func (r *Resource) parseURITemplate(uriTemplate string) (artifacts.URITemplateInfo, error) {
	parsedInfo := artifacts.URITemplateInfo{
		FullTemplate:  uriTemplate,
		PathParameters: []string{},
		QueryParameters: make(map[string]string),
	}

	// Split the URI template into path and query parts
	parts := strings.SplitN(uriTemplate, "?", 2)
	pathPart := parts[0]
	parsedInfo.PathTemplate = pathPart

	var queryPart string
	if len(parts) > 1 {
		queryPart = parts[1]
	}

	// Extract path parameters
	pathSegments := strings.Split(pathPart, "/")
	paramSet := make(map[string]bool) // To track duplicate parameters

	for _, segment := range pathSegments {
		matches := pathParamRegex.FindStringSubmatch(segment)
		if len(matches) > 1 {
			paramName := matches[1]
			if paramSet[paramName] {
				return artifacts.URITemplateInfo{}, fmt.Errorf("duplicate path parameter: %s in uri-template: %s", paramName, uriTemplate)
			}
			paramSet[paramName] = true
			parsedInfo.PathParameters = append(parsedInfo.PathParameters, paramName)
		} else if strings.Contains(segment, "{") || strings.Contains(segment, "}") {
			return artifacts.URITemplateInfo{}, fmt.Errorf("invalid path parameter format in segment: '%s' of uri-template: %s. Expected '{paramName}'", segment, uriTemplate)
		}
	}

	// Extract query parameters
	if queryPart != "" {
		queryPairs := strings.Split(queryPart, "&")
		for _, pair := range queryPairs {
			keyValue := strings.SplitN(pair, "=", 2)
			if len(keyValue) == 2 {
				key := keyValue[0]
				value := keyValue[1]
				if _, exists := parsedInfo.QueryParameters[key]; exists {
					return artifacts.URITemplateInfo{}, fmt.Errorf("duplicate query parameter: %s in uri-template: %s", key, uriTemplate)
				}

				// Check if value follows the format {value}
				if !strings.HasPrefix(value, "{") || !strings.HasSuffix(value, "}") {
					return artifacts.URITemplateInfo{}, fmt.Errorf("invalid query parameter value format: '%s' in uri-template: %s. Expected format 'queryparam={value}'", pair, uriTemplate)
				}

				// Remove curly braces from the value
				value = strings.TrimPrefix(value, "{")
				value = strings.TrimSuffix(value, "}")
				parsedInfo.QueryParameters[key] = value
			} else if len(keyValue) == 1 && keyValue[0] != "" {
				return artifacts.URITemplateInfo{}, fmt.Errorf("invalid query parameter format: '%s' in uri-template: %s. Expected 'key=value'", pair, uriTemplate)
			} else if len(keyValue) == 0 && pair != "" {
				return artifacts.URITemplateInfo{}, fmt.Errorf("invalid query parameter format: '%s' in uri-template: %s. Expected 'key=value'", pair, uriTemplate)
			}
		}
	}
	return parsedInfo, nil
}