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

	"github.com/apache/synapse-go/internal/app/core/ports"
)

type InboundService struct {
	inbound ports.InboundEndpoint
}

func NewInboundService(inbound ports.InboundEndpoint) *InboundService {
	return &InboundService{
		inbound: inbound,
	}
}

func (s *InboundService) Start(ctx context.Context, mediator ports.InboundMessageMediator) error {
	return s.inbound.Start(ctx, mediator)
}

func (s *InboundService) Stop() error {
	return s.inbound.Stop()
}
