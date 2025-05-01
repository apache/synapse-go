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

package artifacts

import (
	"sync"

	"github.com/apache/synapse-go/internal/pkg/core/common"
)

// Use the Position from common package
type Position = common.Position

// Break the cyclic dependency between artifacts and configuration context
type EndpointProvider interface {
	GetEndpoint(epName string) *Endpoint
}

type ConfigContext struct {
	ApiMap       map[string]API
	EndpointMap  map[string]Endpoint
	SequenceMap  map[string]Sequence
	InboundMap   map[string]Inbound
	DeploymentConfig map[string]interface{}
}

func (c *ConfigContext) AddAPI(api API) {
	c.ApiMap[api.Name] = api
}

func (c *ConfigContext) AddEndpoint(endpoint Endpoint) {
	c.EndpointMap[endpoint.Name] = endpoint
}

func (c *ConfigContext) AddSequence(sequence Sequence) {
	c.SequenceMap[sequence.Name] = sequence
}

func (c *ConfigContext) AddInbound(inbound Inbound) {
	c.InboundMap[inbound.Name] = inbound
}

func (c *ConfigContext) AddDeploymentConfig(deploymentConfig map[string]interface{}) {
	c.DeploymentConfig = deploymentConfig
}

func (c *ConfigContext) GetEndpoint(epName string) Endpoint {
	endpoint, exists := c.EndpointMap[epName]
	if !exists {
		return Endpoint{}
	}
	return endpoint
}

var instance *ConfigContext

var once sync.Once

// singleton instance of the ConfigContext
func GetConfigContext() *ConfigContext {
	once.Do(func() {
		instance = &ConfigContext{
			ApiMap:      make(map[string]API),
			EndpointMap: make(map[string]Endpoint),
			SequenceMap: make(map[string]Sequence),
			InboundMap:  make(map[string]Inbound),
			DeploymentConfig: make(map[string]interface{}),
		}
	})
	return instance
}
