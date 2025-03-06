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

package types

import (
	"encoding/xml"
	"errors"
	"strconv"

	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
)

type LogMediator struct {
	XMLName  xml.Name `xml:"log"`
	Category string   `xml:"category,attr"`
	Message  string   `xml:"message"`
}

func (logMediator LogMediator) Unmarshal(d *xml.Decoder, start xml.StartElement, position artifacts.Position) (artifacts.Mediator, error) {
	if err := d.DecodeElement(&logMediator, &start); err != nil {
		return artifacts.LogMediator{}, errors.New("error in unmarshalling log mediator in " + position.FileName + " at line " + strconv.Itoa(position.LineNo))
	}
	position.Hierarchy = position.Hierarchy + "->log"
	return artifacts.LogMediator{
		Category: logMediator.Category,
		Message:  logMediator.Message,
		Position: position,
	}, nil
}
