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

package deployer

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/apache/synapse-go/internal/app/adapters/inboundfactory"
	"github.com/apache/synapse-go/internal/app/core/models"
	"github.com/apache/synapse-go/internal/pkg/artifacts"
	"github.com/apache/synapse-go/internal/pkg/consolelogger"
)

func DeployInbound(folderPath string, confContext *artifacts.ConfigContext, context context.Context) error {
	waitgroup := context.Value("waitGroup").(*sync.WaitGroup)
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".xml" {
			continue
		}

		xmlFile, err := os.Open(filepath.Join(folderPath, file.Name()))
		if err != nil {
			fmt.Println("Error opening file:", err)
			continue
		}
		defer xmlFile.Close()

		data, err := io.ReadAll(xmlFile)
		if err != nil {
			fmt.Println("Error reading file:", err)
			continue
		}

		consolelogger.InfoLog("Deploying inbound-endpoint from " + file.Name())

		var inbound artifacts.Inbound
		inbound.Hierarchy = file.Name()
		if err := xml.Unmarshal(data, &inbound); err != nil {
			fmt.Println("Error unmarshaling XML:", err)
			continue
		}

		_, seqExists := confContext.SequenceMap[inbound.Sequence]
		if !seqExists {
			consolelogger.ErrorLog("Sequence " + inbound.Sequence + " not found for Inbound-Endpoint " + inbound.Name)
			continue
		}

		_, ok := confContext.InboundMap[inbound.Name]
		if ok {
			consolelogger.ErrorLog("Inbound-Endpoint " + inbound.Name + " already deployed")
			continue
		}

		confContext.AddInbound(inbound)
		consolelogger.InfoLog("Inbound-Endpoint " + inbound.Name + " deployed successfully")

		var parameters = make(map[string]string)
		for _, param := range inbound.Parameters {
			parameters[param.Name] = param.Value
		}
		// starting the inbound
		fileInboundEndpoint, err := inboundfactory.GetInstance("file", models.InboundConfig{
			Name:         inbound.Name,
			Protocol:     "file",
			SequenceName: inbound.Sequence,
			Parameters:   parameters,
		})
		if err == nil {
			waitgroup.Add(1)
			go fileInboundEndpoint.Start(context)
		}
	}
	return nil
}
