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
	"context"
	"testing"

	"github.com/apache/synapse-go/internal/pkg/core/synctx"
)

func TestLogMediator_Execute(t *testing.T) {
	tests := []struct {
		name     string
		category string
		message  string
		wantErr  bool
	}{
		{
			name:     "Test with INFO category",
			category: "INFO",
			message:  "This is an info message",
			wantErr:  false,
		},
		{
			name:     "Test with ERROR category",
			category: "ERROR",
			message:  "This is an error message",
			wantErr:  false,
		},
		{
			name:     "Test with empty category",
			category: "",
			message:  "This is a message with no category",
			wantErr:  false,
		},
		{
			name:     "Test with empty message",
			category: "DEBUG",
			message:  "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := &LogMediator{
				Category: tt.category,
				Message:  tt.message,
			}
			msgContext := synctx.MsgContext{}
			got, err := lm.Execute(&msgContext, context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("LogMediator.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != true {
				t.Errorf("LogMediator.Execute() = %v, want %v", got, true)
			}
		})
	}
}
