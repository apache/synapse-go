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

package inbound

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/apache/synapse-go/internal/app/core/domain"
	"github.com/apache/synapse-go/internal/app/core/ports"
	"github.com/apache/synapse-go/internal/pkg/core/synctx"
	"github.com/apache/synapse-go/internal/pkg/core/utils"
)

type FileInboundEndpoint struct {
	Config    domain.InboundConfig
	IsRunning bool
}

func (inbound *FileInboundEndpoint) Start(ctx context.Context, mediator ports.InboundMessageMediator) error {
	inbound.IsRunning = true
	waitgroup := ctx.Value(utils.WaitGroupKey).(*sync.WaitGroup)
	if inbound.Config.Protocol == "file" {
		var fileContent = "Hello World"
		interval, found := inbound.getIntervalParameterValue()
		if found {
			fmt.Printf("Polling file every %d milliseconds\n", interval)
		} else {
			fmt.Println("Interval parameter not found")
			return errors.New("interval parameter not found")
		}

		for {
			if !inbound.IsRunning {
				fmt.Println("Cleaning up file polling")
				waitgroup.Done()
			}
			select {
			case <-ctx.Done():
				fmt.Println("Cleaning up file polling gracefully")
				waitgroup.Done()
				return nil
			default:
				fmt.Println("Polling file")
				// Creating the new message context from file content
				var context = synctx.MsgContext{
					Properties: make(map[string]string),
					Message: synctx.Message{
						RawPayload:  []byte(fileContent),
						ContentType: "text/plain",
					},
					Headers: make(map[string]string),
				}
				// Mediating the inbound message
				mediator.MediateInboundMessage(ctx, inbound.Config.SequenceName, &context)
			}
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	} else {
		return errors.New("invalid protocol")
	}
}

// call this using a channel
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
