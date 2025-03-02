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

package mediation

import (
	"context"
	"fmt"
	"sync"

	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
	"github.com/apache/synapse-go/internal/pkg/core/synctx"
)

type MediationEngine struct {
	ConfigContext *artifacts.ConfigContext
}

func NewMediationEngine(configContext *artifacts.ConfigContext) *MediationEngine {
	return &MediationEngine{
		ConfigContext: configContext,
	}
}

func (m *MediationEngine) MediateInboundMessage(ctx context.Context, seqName string, msg *synctx.MsgContext) error {
	waitgroup := ctx.Value("waitGroup").(*sync.WaitGroup)
	waitgroup.Add(1)
	go func() {
		defer waitgroup.Done()
		select {
		case <-ctx.Done():
			fmt.Println("Mediation of sequence stopped since context is done")
			waitgroup.Done()
			return
		default:
			sequence, exists := m.ConfigContext.SequenceMap[seqName]
			if !exists {
				fmt.Println("Sequence " + seqName + " not found")
				return
			}
			sequence.Execute(msg)
		}
	}()
	return nil

}
