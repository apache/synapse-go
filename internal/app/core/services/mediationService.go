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

package services

import (
	"context"
	"sync"

	"github.com/apache/synapse-go/internal/app/core/models"
	"github.com/apache/synapse-go/internal/pkg/artifacts"
	"github.com/apache/synapse-go/internal/pkg/consolelogger"
	"github.com/apache/synapse-go/internal/pkg/synapsecontext"
)

type MediationService struct {
	*artifacts.ConfigContext
}

var instance *MediationService
var once sync.Once

func GetMediationEngine() *MediationService {
	once.Do(func() {
		instance = &MediationService{
			ConfigContext: artifacts.GetConfigContextInstance(),
		}
	})
	return instance
}

func (me *MediationService) MediateInboundMessage(inbound models.InboundConfig,
	msg *synapsecontext.SynapseContext, ctx context.Context) error {
	seqName := inbound.SequenceName
	waitgroup := ctx.Value("waitGroup").(*sync.WaitGroup)
	waitgroup.Add(1)
	go func() {
		defer waitgroup.Done()
		select {
		case <-ctx.Done():
			consolelogger.DebugLog("Mediation of sequence stopped since context is done")
			waitgroup.Done()
			return
		default:
			sequence, exists := me.ConfigContext.SequenceMap[seqName]
			if !exists {
				consolelogger.ErrorLog("Sequence " + seqName + " not found")
				return
			}
			sequence.Execute(msg)
		}
	}()
	return nil
}
