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

type RespondMediator struct {
	XMLName  xml.Name `xml:"respond"`
}

func (respondMediator RespondMediator) Unmarshal(d *xml.Decoder, start xml.StartElement, position artifacts.Position) (artifacts.Mediator, error) {
	if err := d.DecodeElement(&respondMediator, &start); err != nil {
		return artifacts.RespondMediator{}, errors.New("error in unmarshalling respond mediator in " + position.FileName + " at line " + strconv.Itoa(position.LineNo))
	}
	position.Hierarchy = position.Hierarchy + "->respond"
	return artifacts.RespondMediator{
		Position: position,
	}, nil
}
