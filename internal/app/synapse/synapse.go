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

package synapse

import (
	"context"
	"runtime"
	"sync"

	"time"

	"github.com/apache/synapse-go/internal/pkg/artifacts"
	"github.com/apache/synapse-go/internal/pkg/consolelogger"
	"github.com/apache/synapse-go/internal/pkg/deployer"
	"github.com/apache/synapse-go/internal/pkg/dispatcher"
)

func Run(ctx context.Context) error {

	// using all available cores
	runtime.GOMAXPROCS(runtime.NumCPU())
	start := time.Now()
	consolelogger.PrintWelcomeMessage()
	router := dispatcher.Router{}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	newCtx := context.WithValue(ctx, "waitGroup", &wg)
	defer cancel()

	// Configuration context to hold all artifacts
	artifactInfo := artifacts.GetConfigContextInstance()

	deployer.DeployArtifacts(&router, artifactInfo, newCtx)

	elapsed := time.Since(start)
	consolelogger.InfoLog("Server started in " + elapsed.String())

	// consolelogger.InfoLog("Access the server at http://localhost:8080")
	// // Start the server in port 8080
	// err := http.ListenAndServe(":8080", &router)
	// if err != nil {
	// 	consolelogger.ErrorLog("Error starting server: " + err.Error())
	// 	return err
	// }

	<-ctx.Done()
	wg.Wait()
	consolelogger.InfoLog("Server stopped")
	return nil
}
