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
	"strings"

	"github.com/apache/synapse-go/internal/pkg/consolelogger"
	"github.com/apache/synapse-go/internal/pkg/synapsecontext"
	"github.com/apache/synapse-go/pkg/utils"
)

type PayloadMediator struct {
	Description string      `xml:"description,attr"`
	MediaType   string      `xml:"media-type,attr"`
	Format      FormatBlock `xml:"format"`
	LineNo      int
	FileName    string
}

// FormatBlock captures all content inside <format>...</format> as a raw string
type FormatBlock struct {
	XMLName  xml.Name `xml:"format"`
	InnerXML string   `xml:",innerxml"`
}

func (l *PayloadMediator) Execute(context *synapsecontext.SynapseContext) bool {
	if l.Format.InnerXML == "" {
		consolelogger.ErrorLog("Error occurred while creating the payload in payload mediator. Empty payload")
		return false
	}
	payload := strings.TrimSpace(l.Format.InnerXML)
	// check Format is a valid JSON
	if l.MediaType == "json" {
		if !utils.IsValidJSON(payload) {
			consolelogger.ErrorLog("Error occurred while creating the payload in payload mediator. Invalid JSON format")
			return false
		} else {
			context.Message.RawPayload = []byte(payload)
			context.Message.ContentType = "application/json"
			context.Headers["Content-Type"] = "application/json"
		}
	} else if l.MediaType == "xml" {
		if !utils.IsValidXML(payload) {
			consolelogger.ErrorLog("Error occurred while creating the payload in payload mediator. Invalid XML format")
			return false
		} else {
			context.Message.ContentType = "application/xml"
			context.Message.RawPayload = []byte(payload)
			context.Headers["Content-Type"] = "application/xml"
		}
	} else {
		consolelogger.ErrorLog("Error occurred while creating the payload in payload mediator. Invalid media type")
		return false
	}
	return true
}

func (l *PayloadMediator) SetFileName(fileName string) {
	l.FileName = fileName
}
