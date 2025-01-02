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

package inboundfactory

import (
	"errors"

	"github.com/apache/synapse-go/internal/app/adapters"
	"github.com/apache/synapse-go/internal/app/core/models"
	"github.com/apache/synapse-go/internal/app/core/ports"
)

func GetInstance(inboundtype string, config models.InboundConfig) (ports.InboundEndpoint, error) {
	if inboundtype == "file" {
		return &adapters.FileInboundEndpoint{
			Config:    config,
			IsRunning: false,
		}, nil
	}
	return nil, errors.New("Inbound type not found")
}
