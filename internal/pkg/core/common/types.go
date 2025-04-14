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

package common

// ConfigProvider defines the interface for configuration operations
type ConfigProvider interface {
	IsSet(key string) bool
	Unmarshal(key string, out interface{}) error
	MustUnmarshal(key string, out interface{})
}

// ArtifactPosition locates the exact position of any deployed artifact
// This is uesd to log the exact position of artifacts in error handling
// hierarchy ex: API -> Resource -> InSequence -> Log
type Position struct {
	LineNo    int
	FileName  string
	Hierarchy string
}
