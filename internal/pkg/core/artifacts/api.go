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
	"github.com/apache/synapse-go/internal/pkg/core/synctx"
)

type Resource struct {
	Methods       []string
	URITemplate   URITemplateInfo
	InSequence    Sequence
	FaultSequence Sequence
}

type URITemplateInfo struct {
	FullTemplate    string            // The original full URI template
	PathTemplate    string            // Just the path part (without query)
	PathParameters  []string          // List of path parameters
	QueryParameters map[string]string // Map of query param name to variable name
}

type API struct {
	Context     string
	Name        string
	Version     string
	VersionType string
	Resources   []Resource
	Position    Position
}

func (r *Resource) Mediate(context *synctx.MsgContext) bool {
	isSuccessInSeq := r.InSequence.Execute(context)
	if !isSuccessInSeq {
		isCompleteFaultSeq := r.FaultSequence.Execute(context)
		if !isCompleteFaultSeq {
			return false
		}
	}
	return true
}
