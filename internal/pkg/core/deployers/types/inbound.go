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

	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
)

type Inbound struct {
	Name       string      `xml:"name,attr"`
	Sequence   string      `xml:"sequence,attr"`
	Protocol   string      `xml:"protocol,attr"`
	Suspend    string      `xml:"suspend,attr"`
	OnError    string      `xml:"onError,attr"`
	Parameters []Parameter `xml:"parameters>parameter"`
	Hierarchy  string
}

type Parameter struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

func (inbound *Inbound) Unmarshal(xmlData string, position artifacts.Position) (artifacts.Inbound, error) {
	newInbound := artifacts.Inbound{}
	newInbound.Position = position
	if err := xml.Unmarshal([]byte(xmlData), &inbound); err != nil {
		return artifacts.Inbound{}, err
	}
	newInbound.Name = inbound.Name
	newInbound.Sequence = inbound.Sequence
	newInbound.Protocol = inbound.Protocol
	newInbound.Suspend = inbound.Suspend
	newInbound.OnError = inbound.OnError
	for _, parameter := range inbound.Parameters {
		newInbound.Parameters = append(newInbound.Parameters, artifacts.Parameter{Name: parameter.Name, Value: parameter.Value})
	}
	newInbound.Position.Hierarchy = position.Hierarchy
	return newInbound, nil
}
