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

type Sequence struct {
	MediatorList []artifacts.Mediator
}

// unmarshal decodes the XML data and populates the Sequence struct
func (seq *Sequence) unmarshal(decoder *xml.Decoder, position artifacts.Position) (artifacts.Sequence, error) {
	var mediatorList []artifacts.Mediator
	if position.Hierarchy == "" {
		position.Hierarchy = "sequence"
	} else {
		position.Hierarchy = position.Hierarchy + "->sequence"
	}
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
			case "respond":
				respondMediator := RespondMediator{}
				mediator, err := respondMediator.Unmarshal(decoder, element, position)
				if err != nil {
					return artifacts.Sequence{}, err
				}
				mediatorList = append(mediatorList, mediator)
			}
		case xml.EndElement:
			// Stop when the </sequence> tag is encountered
			if element.Name.Local == "sequence" {
				break OuterLoop
			}
		}
	}
	return artifacts.Sequence{MediatorList: mediatorList, Position: position}, nil
}

func (seq *Sequence) Unmarshal(xmlData string, position artifacts.Position) (artifacts.Sequence, error) {
	decoder := xml.NewDecoder(strings.NewReader(xmlData))
	for {
		token, err := decoder.Token()
		if err != nil {
			return artifacts.Sequence{}, err
		}
		if startElem, ok := token.(xml.StartElement); ok && startElem.Name.Local == "sequence" {
			for _, attr := range startElem.Attr {
				switch attr.Name.Local {
				case "name":
					position := artifacts.Position{LineNo: 1, FileName: position.FileName, Hierarchy: attr.Value}

					newSeq, err := seq.unmarshal(decoder, position)
					if err != nil {
						return artifacts.Sequence{}, err
					}
					newSeq.Name = attr.Value
					return newSeq, nil
				}
			}
			break
		}
	}
	return artifacts.Sequence{}, nil
}