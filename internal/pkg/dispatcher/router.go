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

package dispatcher

import (
	"net/http"
)

type Router struct {
	Routes map[string]map[string]http.HandlerFunc // path -> method -> handler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	if methodHandlers, ok := r.Routes[req.URL.Path]; ok {
		if handler, ok := methodHandlers[req.Method]; ok {
			handler(w, req)
			return
		}
	}
	http.NotFound(w, req)
}

// addRoute is a helper to add a new route
func (r *Router) AddRoute(method, pattern string, handler http.HandlerFunc) {
	if r.Routes == nil {
		r.Routes = make(map[string]map[string]http.HandlerFunc)
	}
	if r.Routes[pattern] == nil {
		r.Routes[pattern] = make(map[string]http.HandlerFunc)
	}
	r.Routes[pattern][method] = handler
}
