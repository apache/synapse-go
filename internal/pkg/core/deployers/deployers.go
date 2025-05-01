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
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/apache/synapse-go/internal/app/adapters/inbound"
	"github.com/apache/synapse-go/internal/app/core/domain"
	"github.com/apache/synapse-go/internal/app/core/ports"
	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
	"github.com/apache/synapse-go/internal/pkg/core/deployers/types"
	"github.com/apache/synapse-go/internal/pkg/core/router"
	"github.com/apache/synapse-go/internal/pkg/core/utils"
	"github.com/apache/synapse-go/internal/pkg/loggerfactory"
)

const (
	componentName = "deployers"
)

type Deployer struct {
	inboundMediator ports.InboundMessageMediator
	routerService   *router.RouterService
	basePath        string
	logger 			*slog.Logger
}

// Synapse/
// ├─ bin/
// │  └─ synapse           (the compiled binary)
// └─ artifacts/
//    ├─ APIs/
//    |─ Endpoints/
//    |─ Sequences/
//    └─ Inbounds/

func NewDeployer(basePath string, inboundMediator ports.InboundMessageMediator, routerService *router.RouterService) *Deployer {
	d := &Deployer{
		basePath:        basePath,
		inboundMediator: inboundMediator,
		routerService:   routerService,
	}
	d.logger = loggerfactory.GetLogger(componentName, d)
	return d
}

func (d *Deployer) UpdateLogger() {
	d.logger = loggerfactory.GetLogger(componentName,d)
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
				d.logger.Error("Error reading file:", "error", err)
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
		d.logger.Error("Error unmarshalling sequence:", "error", err)
		return
	}
	configContext := ctx.Value(utils.ConfigContextKey).(*artifacts.ConfigContext)
	configContext.AddSequence(newSeq)
	d.logger.Info("Deployed sequence: " + newSeq.Name)
}

func (d *Deployer) DeployAPIs(ctx context.Context, fileName string, xmlData string) {
	position := artifacts.Position{FileName: fileName}
	api := types.API{}
	newApi, err := api.Unmarshal(xmlData, position)
	if err != nil {
		d.logger.Error("Error unmarshalling api:", "error", err)
		return
	}
	configContext := ctx.Value(utils.ConfigContextKey).(*artifacts.ConfigContext)
	configContext.AddAPI(newApi)

	d.logger.Info("Deployed API: " + newApi.Name)

	// Register the API with the router service
	if err := d.routerService.RegisterAPI(ctx, newApi); err != nil {
		d.logger.Error("Error registering API with router service:", "error", err)
		return
	}
}

func (d *Deployer) DeployInbounds(ctx context.Context, fileName string, xmlData string) {
	position := artifacts.Position{FileName: fileName}
	inboundEp := types.Inbound{}
	newInbound, err := inboundEp.Unmarshal(xmlData, position)
	if err != nil {
		d.logger.Error("Error unmarshalling inbound:", "error", err)
		return
	}
	configContext := ctx.Value(utils.ConfigContextKey).(*artifacts.ConfigContext)
	configContext.AddInbound(newInbound)
	d.logger.Info("Deployed inbound: " + newInbound.Name)

	// Start the inbound endpoint
	parametersMap := make(map[string]string)
	for _, param := range newInbound.Parameters {
		parametersMap[param.Name] = param.Value
	}
	inboundEndpoint, err := inbound.NewInbound(domain.InboundConfig{
		SequenceName: newInbound.Sequence,
		Name:         newInbound.Name,
		Protocol:     newInbound.Protocol,
		Parameters:   parametersMap,
	})
	if err != nil {
		d.logger.Error("Error creating inbound endpoint:", "error", err)
		return
	}

	wg := ctx.Value(utils.WaitGroupKey).(*sync.WaitGroup)
	wg.Add(1)
	go func(endpoint ports.InboundEndpoint) {
		defer wg.Done()
		if err := endpoint.Start(ctx, d.inboundMediator); err != nil {
			d.logger.Error("Error starting inbound endpoint:", "error", err)
		}
	}(inboundEndpoint)
}
