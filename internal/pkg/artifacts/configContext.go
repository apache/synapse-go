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

package artifacts

import (
	"sync"
)

// ConfigContext struct which holds the deployed artifact details
type ConfigContext struct {
	ApiMap      map[string]API
	EndpointMap map[string]Endpoint
	SequenceMap map[string]NamedSequence
	InboundMap  map[string]Inbound
}

func (c *ConfigContext) AddAPI(api API) {
	c.ApiMap[api.Name] = api
}

func (c *ConfigContext) AddEndpoint(endpoint Endpoint) {
	c.EndpointMap[endpoint.Name] = endpoint
}

func (c *ConfigContext) AddSequence(sequence NamedSequence) {
	c.SequenceMap[sequence.Name] = sequence
}

func (c *ConfigContext) AddInbound(inbound Inbound) {
	c.InboundMap[inbound.Name] = inbound
}

var instance *ConfigContext

var once sync.Once

// singleton instance of the ConfigContext
func GetConfigContextInstance() *ConfigContext {
	once.Do(func() {
		instance = &ConfigContext{
			ApiMap:      make(map[string]API),
			EndpointMap: make(map[string]Endpoint),
			SequenceMap: make(map[string]NamedSequence),
			InboundMap:  make(map[string]Inbound),
		}
	})
	return instance
}
