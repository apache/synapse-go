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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/apache/synapse-go/internal/pkg/core/synctx"
	"gopkg.in/yaml.v2"
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
	CORSConfig  CORSConfig
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

// This function calculates the base path based on the API context and versioning type.
func (api *API) calculateBasePath() string {
	basePath := api.Context

	// Remove trailing slash from context if present
	if len(basePath) > 1 && strings.HasSuffix(basePath, "/") {
		basePath = basePath[:len(basePath)-1]
	}

	// Ensure basePath starts with a slash if not empty
	if basePath != "" && !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}

	// Handle versioning based on versionType
	if api.Version != "" && api.VersionType != "" {
		switch api.VersionType {
		case "url":
			// For URL type, add version as a path segment
			// Ensure no double slash if basePath is "/"
			if basePath == "/" {
				basePath = "/" + api.Version
			} else if basePath == "" {
				basePath = "/" + api.Version
			} else {
				basePath = basePath + "/" + api.Version
			}
		case "context":
			// For context type, replace {version} placeholder if it exists
			versionPattern := "{version}"
			basePath = strings.Replace(basePath, versionPattern, api.Version, 1)
		}
	}
	return basePath
}

// GenerateOpenAPISpec creates an OpenAPI 3.0.x specification document
// as a map[string]interface{} based on the input API definition.
// port = ":8290"
func (api *API) GenerateOpenAPISpec(hostname string, port string) (map[string]interface{}, error) {
	// --- 1. Basic OpenAPI Structure ---
	spec := make(map[string]interface{})
	spec["openapi"] = "3.0.3" // Specify OpenAPI version

	// --- 2. Info Object ---
	info := make(map[string]interface{})
	// Use API Name or a default if context is empty.
	title := api.Name
	if title == "" {
		title = "API Documentation"
	}
	info["title"] = title
	info["version"] = api.Version
	spec["info"] = info

	// --- 3. Servers Object ---
	basePath := api.calculateBasePath()
	serverURL := fmt.Sprintf("http://%s%s%s", hostname, port, basePath)
	// Validate the URL
	_, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("generated server URL is invalid '%s': %w", serverURL, err)
	}

	servers := []map[string]interface{}{
		{
			"url": serverURL,
		},
	}
	spec["servers"] = servers

	// --- 4. Paths Object ---
	paths := make(map[string]interface{})

	for _, resource := range api.Resources {
		pathTemplate := resource.URITemplate.PathTemplate
		// OpenAPI paths MUST start with a '/'
		if !strings.HasPrefix(pathTemplate, "/") {
			pathTemplate = "/" + pathTemplate
		}

		// Get or create the Path Item Object for this path
		pathItem, exists := paths[pathTemplate]
		if !exists {
			pathItem = make(map[string]interface{})
			paths[pathTemplate] = pathItem
		}
		pathItemMap := pathItem.(map[string]interface{}) // Type assertion

		// --- 5. Parameters (Path and Query) ---
		parameters := make([]interface{}, 0)

		// Path Parameters
		for _, paramName := range resource.URITemplate.PathParameters {
			parameters = append(parameters, map[string]interface{}{
				"name":        paramName,
				"in":          "path",
				"required":    true,                                         // Path parameters are always required
				"description": fmt.Sprintf("Path parameter: %s", paramName), // Optional description
				"schema": map[string]interface{}{
					"type": "string", // Defaulting to string.
				},
			})
		}

		// Query Parameters
		for queryParamName := range resource.URITemplate.QueryParameters {
			parameters = append(parameters, map[string]interface{}{
				"name":        queryParamName,
				"in":          "query",
				"required":    true,
				"description": fmt.Sprintf("Query parameter: %s", queryParamName), // Optional description
				"schema": map[string]interface{}{
					"type": "string", // Defaulting to string.
				},
			})
		}

		// --- 6. Operations (GET, POST, PUT, etc.) ---
		for _, method := range resource.Methods {
			httpMethod := strings.ToLower(method) // OpenAPI methods are lowercase (get, post, put)

			operation := make(map[string]interface{})
			// Add a basic summary
			operation["summary"] = fmt.Sprintf("%s operation for %s", strings.ToUpper(httpMethod), pathTemplate)

			if len(parameters) > 0 {
				operation["parameters"] = parameters
			}

			// --- 7. Request Body (Needed for POST, PUT, PATCH, etc.) ---
			if httpMethod == "post" || httpMethod == "put" || httpMethod == "patch" {
				requestBody := make(map[string]interface{})
				requestBody["description"] = "Request body payload" // Placeholder description
				requestBody["required"] = true                      // Typically true for POST/PUT

				content := make(map[string]interface{})

				content["application/json"] = map[string]interface{}{
					"schema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"message": map[string]interface{}{
								"type":    "string",
								"example": "Placeholder - Define actual schema based on API needs",
							},
						},
					},
				}
				requestBody["content"] = content
				operation["requestBody"] = requestBody
			}

			// --- 8. Responses ---
			// Adding minimal default responses.
			responses := make(map[string]interface{})
			responses["200"] = map[string]interface{}{ // Successful response
				"description": "OK", // Basic description
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"status": map[string]interface{}{
									"type":    "string",
									"example": "Success",
								},
							},
						},
					},
				},
			}

			responses["default"] = map[string]interface{}{ // Default response for errors
				"description": "Unexpected error",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"error": map[string]interface{}{
									"type":    "string",
									"example": "Error details",
								},
							},
						},
					},
				},
			}
			operation["responses"] = responses

			// Assign the operation to the correct method in the path item
			pathItemMap[httpMethod] = operation
		} // End methods loop
	} // End resources loop

	spec["paths"] = paths

	return spec, nil
}

// ServeSwaggerJSON generates the OpenAPI spec as JSON and writes it to the provided http.ResponseWriter.
// port = ":8290"
func (api *API) ServeSwaggerJSON(w http.ResponseWriter, hostname string, port string) error {
	spec, err := api.GenerateOpenAPISpec(hostname, port)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate OpenAPI spec: %v", err), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(spec); err != nil {
		// Error might occur if connection is closed prematurely
		http.Error(w, fmt.Sprintf("Failed to write JSON response: %v", err), http.StatusInternalServerError)
		return err
	}

	return nil
}

// ServeSwaggerYAML generates the OpenAPI spec as YAML and writes it to the provided http.ResponseWriter.
// port = ":8290"
func (api *API) ServeSwaggerYAML(w http.ResponseWriter, hostname string, port string) error {
	spec, err := api.GenerateOpenAPISpec(hostname, port)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate OpenAPI spec: %v", err), http.StatusInternalServerError)
		return err
	}

	// Marshal the map into YAML bytes
	yamlBytes, err := yaml.Marshal(spec)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal spec to YAML: %v", err), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// Write the YAML bytes to the response writer
	_, err = w.Write(yamlBytes)
	if err != nil {
		return err
	}
	return nil
}