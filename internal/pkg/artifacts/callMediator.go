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
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/apache/synapse-go/internal/pkg/consolelogger"
	"github.com/apache/synapse-go/internal/pkg/synapsecontext"
)

type CallMediator struct {
	Endpoint CallEndpoint `xml:"endpoint"`
	LineNo   int
	FileName string
}

type CallEndpoint struct {
	Key string `xml:"key,attr"`
}

var sharedClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        1000,
		MaxConnsPerHost:     1000,
		MaxIdleConnsPerHost: 1000,
		IdleConnTimeout:     90 * time.Second,
	},
}

func (l *CallMediator) Execute(context *synapsecontext.SynapseContext) bool {

	ep, exists := GetConfigContextInstance().EndpointMap[l.Endpoint.Key]
	if !exists {
		consolelogger.ErrorLog("Endpoint not found: " + l.Endpoint.Key)
		return false
	}
	// Create a new HTTP request with the specified method, URL, and optional body
	req, err := http.NewRequest(strings.ToUpper(ep.EndpointUrl.Method),
		ep.EndpointUrl.URL, bytes.NewReader(context.Message.RawPayload))
	if err != nil {
		consolelogger.ErrorLog("failed to create request: " + err.Error())
		return false
	}
	req.Header.Set("Content-Type", context.Message.ContentType)

	// Send the request
	resp, err := sharedClient.Do(req)
	if err != nil {
		consolelogger.ErrorLog("failed to send request: " + err.Error())
		return false
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		consolelogger.ErrorLog("failed to read response body: " + err.Error())
		return false
	}
	context.Message.RawPayload = respBody
	context.Message.ContentType = resp.Header.Get("Content-Type")
	return true
}

func (l *CallMediator) SetFileName(fileName string) {
	l.FileName = fileName
}
