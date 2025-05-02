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
	"fmt"
	
	"github.com/apache/synapse-go/internal/pkg/core/synctx"
)

// Dummy implementation of the RespondMediator
// This is a placeholder for the actual implementation
type RespondMediator struct {
	Position Position
}

func (rm RespondMediator) Execute(context *synctx.MsgContext) (bool, error) {
	// Set msg context header as a flag. If the header is set, it means the message has been processed and good to response
	context.Headers["http-response"] = "true"
	fmt.Println("Respond mediator")
	return true, nil
}
  