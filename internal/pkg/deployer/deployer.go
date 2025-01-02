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
	"os"
	"path/filepath"

	"github.com/apache/synapse-go/internal/pkg/artifacts"
	"github.com/apache/synapse-go/internal/pkg/consolelogger"
	"github.com/apache/synapse-go/internal/pkg/dispatcher"
)

func DeployArtifacts(router *dispatcher.Router, confContext *artifacts.ConfigContext, ctx context.Context) {
	exePath, err := os.Executable()
	if err != nil {
		consolelogger.ErrorLog("Failed to get executable path")
	}

	binDir := filepath.Dir(exePath)

	// Synapse/
	// ├─ bin/
	// │  └─ synapse           (the compiled binary)
	// └─ artifacts/
	//    ├─ APIs/
	//    |─ Endpoints/
	//    |─ Sequences/
	//    └─ Inbounds/

	artifactsPath := filepath.Join(binDir, "..", "artifacts")
	sequencePath := filepath.Join(artifactsPath, "Sequences")
	DeploySequences(sequencePath, confContext)
	apiPath := filepath.Join(artifactsPath, "APIs")
	DeployAPIs(apiPath, router, confContext)
	endpointsPath := filepath.Join(artifactsPath, "Endpoints")
	DeployEndpoints(endpointsPath, router, confContext)
	inboundsPath := filepath.Join(artifactsPath, "Inbounds")
	DeployInbound(inboundsPath, confContext, ctx)
}
