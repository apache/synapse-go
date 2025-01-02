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
