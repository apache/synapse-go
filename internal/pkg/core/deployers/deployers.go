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

package deployers

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/apache/synapse-go/internal/app/adapters/inbound"
	"github.com/apache/synapse-go/internal/app/core/domain"
	"github.com/apache/synapse-go/internal/app/core/ports"
	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
	"github.com/apache/synapse-go/internal/pkg/core/deployers/types"
	"github.com/apache/synapse-go/internal/pkg/core/utils"
)

type Deployer struct {
	inboundMediator ports.InboundMessageMediator
	basePath        string
}

// Synapse/
// ├─ bin/
// │  └─ synapse           (the compiled binary)
// └─ artifacts/
//    ├─ APIs/
//    |─ Endpoints/
//    |─ Sequences/
//    └─ Inbounds/

func NewDeployer(basePath string, inboundMediator ports.InboundMessageMediator) *Deployer {
	return &Deployer{basePath: basePath, inboundMediator: inboundMediator}
}

func (d *Deployer) Deploy(ctx context.Context) error {
	files, err := os.ReadDir(d.basePath)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}
	for _, artifactType := range []string{"Sequences", "APIs", "Inbounds"} {
		folderPath := filepath.Join(d.basePath, artifactType)
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
				return err
			}
			defer xmlFile.Close()
			data, err := io.ReadAll(xmlFile)
			if err != nil {
				fmt.Println("Error reading file:", err)
				continue
			}
			switch artifactType {
			case "APIs":
				d.DeployAPIs(ctx, file.Name(), string(data))
			case "Sequences":
				d.DeploySequences(ctx, file.Name(), string(data))
			case "Inbounds":
				d.DeployInbounds(ctx, file.Name(), string(data))
			}
		}
	}
	return nil
}

func (d *Deployer) DeploySequences(ctx context.Context, fileName string, xmlData string) {
	position := artifacts.Position{FileName: fileName}
	sequence := types.Sequence{}
	newSeq, err := sequence.Unmarshal(xmlData, position)
	if err != nil {
		fmt.Println("Error unmarshalling sequence:", err)
		return
	}
	configContext := ctx.Value(utils.ConfigContextKey).(*artifacts.ConfigContext)
	configContext.AddSequence(newSeq)
	fmt.Println("Deployed sequence: ", newSeq.Name)
}

func (d *Deployer) DeployAPIs(ctx context.Context, fileName string, xmlData string) {
	position := artifacts.Position{FileName: fileName}
	api := types.API{}
	newApi, err := api.Unmarshal(xmlData, position)
	if err != nil {
		fmt.Println("Error unmarshalling sequence:", err)
		return
	}
	configContext := ctx.Value(utils.ConfigContextKey).(*artifacts.ConfigContext)
	configContext.AddAPI(newApi)
	fmt.Println("Deployed API: ", newApi.Name)
}

func (d *Deployer) DeployInbounds(ctx context.Context, fileName string, xmlData string) {
	position := artifacts.Position{FileName: fileName}
	inboundEp := types.Inbound{}
	newInbound, err := inboundEp.Unmarshal(xmlData, position)
	if err != nil {
		fmt.Println("Error unmarshalling sequence:", err)
		return
	}
	configContext := ctx.Value(utils.ConfigContextKey).(*artifacts.ConfigContext)
	configContext.AddInbound(newInbound)
	fmt.Println("Deployed inbound: ", newInbound.Name)

	// Start the inbound endpoint
	parametersMap := make(map[string]string)
	for _, param := range newInbound.Parameters {
		parametersMap[param.Name] = param.Value
	}
	inboundEndpoint, eerr := inbound.NewInbound(domain.InboundConfig{
		SequenceName: newInbound.Sequence,
		Name:         newInbound.Name,
		Protocol:     newInbound.Protocol,
		Parameters:   parametersMap,
	})
	if eerr != nil {
		fmt.Println("Error creating inbound endpoint:", eerr)
		return
	}
	wg := ctx.Value(utils.WaitGroupKey).(*sync.WaitGroup)
	wg.Add(1)
	go func() {
		err = inboundEndpoint.Start(ctx, d.inboundMediator)
		if err != nil {
			fmt.Println("Error starting inbound endpoint:", err)
			wg.Done()
			return
		}
	}()
}
