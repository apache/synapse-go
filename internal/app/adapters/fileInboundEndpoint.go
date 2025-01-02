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

package adapters

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"strconv"

	"github.com/apache/synapse-go/internal/app/core/models"
	"github.com/apache/synapse-go/internal/app/core/ports"
	"github.com/apache/synapse-go/internal/app/core/services"
	"github.com/apache/synapse-go/internal/pkg/consolelogger"
	"github.com/apache/synapse-go/internal/pkg/synapsecontext"
)

type FileInboundEndpoint struct {
	Config    models.InboundConfig
	IsRunning bool
}

func (adapter *FileInboundEndpoint) Start(ctx context.Context) error {
	adapter.IsRunning = true
	waitgroup := ctx.Value("waitGroup").(*sync.WaitGroup)
	if adapter.Config.Protocol == "file" {
		var fileContent = "Hello World"
		interval, found := adapter.getIntervalParameterValue()
		if found {
			consolelogger.InfoLog(fmt.Sprintf("Polling file every %d milliseconds", interval))
		} else {
			consolelogger.ErrorLog("Interval parameter not found")
			return errors.New("interval parameter not found")
		}

		for {
			if !adapter.IsRunning {
				consolelogger.InfoLog("Cleaning up file polling")
				waitgroup.Done()
			}
			select {
			case <-ctx.Done():
				fmt.Println("Cleaning up file polling gracefully")
				consolelogger.InfoLog("Cleaning up file polling gracefully")
				waitgroup.Done()
				return nil
			default:
				consolelogger.DebugLog("Polling file")
				// Creating the new message context from file content
				var context = synapsecontext.SynapseContext{
					Properties: make(map[string]string),
					Message: synapsecontext.Message{
						RawPayload:  []byte(fileContent),
						ContentType: "text/plain",
					},
					Headers: make(map[string]string),
				}
				ports.InboundMessageMediator.MediateInboundMessage(services.GetMediationEngine(), adapter.Config, &context, ctx)
			}
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	} else {
		return errors.New("invalid protocol")
	}
}

func (adapter *FileInboundEndpoint) Stop() error {
	adapter.IsRunning = false
	return nil
}

func (adapter *FileInboundEndpoint) getIntervalParameterValue() (int, bool) {
	if adapter.Config.Parameters["interval"] != "" {
		interval, err := strconv.Atoi(adapter.Config.Parameters["interval"])
		if err != nil {
			return 0, false
		}
		return interval, true
	}
	return 0, false
}
