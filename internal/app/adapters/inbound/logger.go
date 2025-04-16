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

package inbound

import "fmt"

type DefaultLogger struct{}

func NewDefaultLogger() *DefaultLogger {
    return &DefaultLogger{}
}

func (l *DefaultLogger) Info(msg string, fields ...Field) {
    fmt.Printf("INFO: %s %v\n", msg, fieldsToMap(fields))
}

func (l *DefaultLogger) Error(msg string, err error, fields ...Field) {
    fmt.Printf("ERROR: %s: %v %v\n", msg, err, fieldsToMap(fields))
}

func (l *DefaultLogger) Debug(msg string, fields ...Field) {
    fmt.Printf("DEBUG: %s %v\n", msg, fieldsToMap(fields))
}

func (l *DefaultLogger) Warn(msg string, fields ...Field) {
    fmt.Printf("WARN: %s %v\n", msg, fieldsToMap(fields))
}

func fieldsToMap(fields []Field) map[string]interface{} {
    m := make(map[string]interface{})
    for _, f := range fields {
        m[f.Key] = f.Value
    }
    return m
} 