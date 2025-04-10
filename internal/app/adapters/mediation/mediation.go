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

package mediation

import (
	"context"
	"log/slog"
	"sync"
	

	"github.com/apache/synapse-go/internal/pkg/core/artifacts"
	"github.com/apache/synapse-go/internal/pkg/core/synctx"
	"github.com/apache/synapse-go/internal/pkg/core/utils"
	"github.com/apache/synapse-go/internal/pkg/loggerfactory"
)

const (
	componentName = "mediation"
)

type MediationEngine struct {
	logger *slog.Logger
}

func NewMediationEngine() *MediationEngine {
	m := &MediationEngine{}
	m.logger = loggerfactory.GetLogger(componentName, m)
	return m
}

func (m *MediationEngine) UpdateLogger() {
	m.logger = loggerfactory.GetLogger(componentName, m)
}

func (m *MediationEngine) MediateInboundMessage(ctx context.Context, seqName string, msg *synctx.MsgContext) error {
	waitgroup := ctx.Value(utils.WaitGroupKey).(*sync.WaitGroup)
	configContext := ctx.Value(utils.ConfigContextKey).(*artifacts.ConfigContext)
	waitgroup.Add(1)
	go func() {
		defer waitgroup.Done()
		select {
		case <-ctx.Done():
			m.logger.Info("Mediation of sequence stopped since context is done")
			waitgroup.Done()
			return
		default:
			sequence, exists := configContext.SequenceMap[seqName]
			if !exists {
				m.logger.Error("Sequence " + seqName + " not found")
				return
			}
			sequence.Execute(msg)
		}
	}()
	return nil

}
