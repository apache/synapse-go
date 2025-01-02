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
	"errors"
	"strconv"

	"github.com/apache/synapse-go/internal/pkg/consolelogger"
	"github.com/apache/synapse-go/internal/pkg/synapsecontext"
)

type LogCategory string

const (
	Debug LogCategory = "DEBUG"
	Info  LogCategory = "INFO"
)

type LogMediator struct {
	Category string `xml:"category,attr"`
	Message  string `xml:"message"`
	Position ArtifactPosition
}

func (l *LogMediator) Execute(context *synapsecontext.SynapseContext) (bool, error) {
	if l.Message != "" {
		switch l.Category {
		case string(Debug):
			consolelogger.DebugLog(l.Message)
		case string(Info):
			consolelogger.InfoLog(l.Message)
		default:
			if l.Category == "" {
				consolelogger.InfoLog(l.Message)
			} else {
				return false, errors.New("log category not found")
			}
		}
	}
	return true, nil
}

func (logMediator *LogMediator) UnmarshalArtifact(d *xml.Decoder, start xml.StartElement, position ArtifactPosition) error {
	if err := d.DecodeElement(logMediator, &start); err != nil {
		return errors.New("error in unmarshalling log mediator in " + position.FileName + " at line " + strconv.Itoa(position.LineNo))
	}
	if logMediator.Category != "" && logMediator.Category != string(Debug) && logMediator.Category != string(Info) {
		return errors.New("invalid log category" + logMediator.Category + " in " + position.FileName + " at line " + strconv.Itoa(position.LineNo))
	}
	logMediator.Position = position
	return nil
}
