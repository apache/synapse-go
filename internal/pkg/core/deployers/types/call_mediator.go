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
	"errors"
	"strconv"

	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
)

type CallMediator struct {
	XMLName     xml.Name `xml:"call"`
	Endpoint    *struct {
		Key string `xml:"key,attr"`
	} `xml:"endpoint,omitempty"`
}

func (callMediator CallMediator) Unmarshal(d *xml.Decoder, start xml.StartElement, position artifacts.Position) (artifacts.Mediator, error) {
	var endpointRef string
	if err := d.DecodeElement(&callMediator, &start); err != nil {
		return nil, errors.New("error in unmarshalling call mediator in " + position.FileName + " at line " + strconv.Itoa(position.LineNo))
	}

	if callMediator.Endpoint != nil && callMediator.Endpoint.Key != "" {
		endpointRef = callMediator.Endpoint.Key
	}

	position.Hierarchy = position.Hierarchy + "->call"
	return artifacts.CallMediator{
		EndpointRef: endpointRef,
		Position:    position,
	}, nil
}
