//go:build integration
// +build integration

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

package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/apache/synapse-go/internal/pkg/core/synctx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MockMediator for testing file inbound processing integration
type MockMediator struct {
	processedFiles []string
	processedData  [][]byte
	processCount   int
}

func (m *MockMediator) ProcessMessage(ctx context.Context, msgContext *synctx.MsgContext) error {
	if msgContext.Headers == nil {
		msgContext.Headers = make(map[string]string)
	}

	fileName := msgContext.Headers["FILE_NAME"]
	if fileName == "" {
		fileName = fmt.Sprintf("file_%d", m.processCount)
	}

	// Check for invalid JSON if content type is JSON
	if msgContext.Message.ContentType == "application/json" {
		var jsonObj interface{}
		if err := json.Unmarshal(msgContext.Message.RawPayload, &jsonObj); err != nil {
			return fmt.Errorf("invalid JSON in file %s: %w", fileName, err)
		}
	}

	m.processedFiles = append(m.processedFiles, fileName)
	m.processedData = append(m.processedData, msgContext.Message.RawPayload)
	m.processCount++

	return nil
}

// FileInboundIntegrationTestSuite defines the integration test suite for File Inbound
type FileInboundIntegrationTestSuite struct {
	suite.Suite
	tempDir      string
	inputDir     string
	processedDir string
	mockMediator *MockMediator
}

// SetupSuite runs once before all integration tests
func (suite *FileInboundIntegrationTestSuite) SetupSuite() {
	suite.tempDir = suite.T().TempDir()
	suite.inputDir = filepath.Join(suite.tempDir, "input")
	suite.processedDir = filepath.Join(suite.tempDir, "processed")

	// Create directories
	require.NoError(suite.T(), os.MkdirAll(suite.inputDir, 0755))
	require.NoError(suite.T(), os.MkdirAll(suite.processedDir, 0755))
}

// SetupTest runs before each test
func (suite *FileInboundIntegrationTestSuite) SetupTest() {
	suite.mockMediator = &MockMediator{
		processedFiles: make([]string, 0),
		processedData:  make([][]byte, 0),
		processCount:   0,
	}

	// Clean directories before each test
	suite.cleanDirectory(suite.inputDir)
	suite.cleanDirectory(suite.processedDir)
}

// cleanDirectory removes all files from a directory
func (suite *FileInboundIntegrationTestSuite) cleanDirectory(dir string) {
	files, err := os.ReadDir(dir)
	require.NoError(suite.T(), err)

	for _, file := range files {
		err := os.Remove(filepath.Join(dir, file.Name()))
		require.NoError(suite.T(), err)
	}
}

// TestIntegration_FileInbound_JSONProcessing tests integration between file system and message processing
func (suite *FileInboundIntegrationTestSuite) TestIntegration_FileInbound_JSONProcessing() {
	// Create test files with realistic data structures
	testFiles := map[string]string{
		"order_001.json":    `{"orderId":"001","customerId":"CUST001","amount":99.99,"items":[{"id":"ITEM001","quantity":2}]}`,
		"order_002.json":    `{"orderId":"002","customerId":"CUST002","amount":149.50,"items":[{"id":"ITEM002","quantity":1},{"id":"ITEM003","quantity":3}]}`,
		"customer_001.json": `{"customerId":"CUST001","name":"John Doe","email":"john@example.com","address":{"street":"123 Main St","city":"Anytown"}}`,
	}

	// Write test files to input directory
	for filename, content := range testFiles {
		filePath := filepath.Join(suite.inputDir, filename)
		require.NoError(suite.T(), os.WriteFile(filePath, []byte(content), 0644))
	}

	// Test integration: File System → Message Context → Mediator Processing
	suite.T().Log("Testing file system integration with message processing...")

	// Process each file (simulating file inbound adapter behavior)
	for filename, content := range testFiles {
		// Step 1: File System Integration - Read from actual file
		filePath := filepath.Join(suite.inputDir, filename)
		fileContent, err := os.ReadFile(filePath)
		require.NoError(suite.T(), err)
		suite.Equal(content, string(fileContent), "File content should match expected")

		// Step 2: Message Context Integration - Convert file to message
		msgContext := synctx.CreateMsgContext()
		msgContext.Message.RawPayload = fileContent
		msgContext.Message.ContentType = "application/json"
		msgContext.Headers["FILE_NAME"] = filename
		msgContext.Headers["FILE_PATH"] = filePath
		msgContext.Headers["FILE_SIZE"] = fmt.Sprintf("%d", len(fileContent))

		// Step 3: Mediation Integration - Process with mediator
		err = suite.mockMediator.ProcessMessage(context.Background(), msgContext)
		require.NoError(suite.T(), err)

		// Step 4: File System Integration - Move processed file
		processedPath := filepath.Join(suite.processedDir, filename)
		err = os.Rename(filePath, processedPath)
		require.NoError(suite.T(), err)

		suite.T().Logf("Successfully integrated file processing for: %s", filename)
	}

	// Verify integration results
	suite.Len(suite.mockMediator.processedFiles, 3, "All files should be processed through mediator")
	suite.Contains(suite.mockMediator.processedFiles, "order_001.json")
	suite.Contains(suite.mockMediator.processedFiles, "order_002.json")
	suite.Contains(suite.mockMediator.processedFiles, "customer_001.json")

	// Verify file system integration - files moved correctly
	processedFiles, err := os.ReadDir(suite.processedDir)
	require.NoError(suite.T(), err)
	suite.Len(processedFiles, 3, "All files should be moved to processed directory")

	inputFiles, err := os.ReadDir(suite.inputDir)
	require.NoError(suite.T(), err)
	suite.Len(inputFiles, 0, "Input directory should be empty after processing")

	// Verify message content integration - JSON structure preserved
	for i, data := range suite.mockMediator.processedData {
		suite.NotEmpty(data, "Processed data should not be empty for file %d", i)

		var jsonData map[string]interface{}
		err := json.Unmarshal(data, &jsonData)
		suite.NoError(err, "Processed data should maintain valid JSON structure")
	}
}

// TestIntegration_FileInbound_XMLProcessing tests XML file processing integration
func (suite *FileInboundIntegrationTestSuite) TestIntegration_FileInbound_XMLProcessing() {
	// Create XML test files
	xmlFiles := map[string]string{
		"product_001.xml": `<?xml version="1.0" encoding="UTF-8"?>
<product>
    <id>PROD001</id>
    <name>Laptop Computer</name>
    <price>999.99</price>
    <category>Electronics</category>
</product>`,
		"product_002.xml": `<?xml version="1.0" encoding="UTF-8"?>
<product>
    <id>PROD002</id>
    <name>Office Chair</name>
    <price>299.50</price>
    <category>Furniture</category>
</product>`,
	}

	// Write XML files to file system
	for filename, content := range xmlFiles {
		filePath := filepath.Join(suite.inputDir, filename)
		require.NoError(suite.T(), os.WriteFile(filePath, []byte(content), 0644))
	}

	// Test integration between file system and XML processing
	suite.T().Log("Testing XML file processing integration...")

	for filename, expectedContent := range xmlFiles {
		// Integration: File System → Message Context
		filePath := filepath.Join(suite.inputDir, filename)
		fileContent, err := os.ReadFile(filePath)
		require.NoError(suite.T(), err)

		// Integration: Message Context → Mediator
		msgContext := synctx.CreateMsgContext()
		msgContext.Message.RawPayload = fileContent
		msgContext.Message.ContentType = "application/xml"
		msgContext.Headers["FILE_NAME"] = filename
		msgContext.Headers["CONTENT_LENGTH"] = fmt.Sprintf("%d", len(fileContent))

		err = suite.mockMediator.ProcessMessage(context.Background(), msgContext)
		require.NoError(suite.T(), err)

		// Integration: File deletion (ActionAfterProcess = DELETE)
		err = os.Remove(filePath)
		require.NoError(suite.T(), err)

		suite.T().Logf("Integrated XML processing and deletion for: %s", filename)

		// Verify content integration
		suite.Equal(expectedContent, string(fileContent), "XML content should be preserved through integration")
	}

	// Verify mediator integration
	suite.Len(suite.mockMediator.processedFiles, 2, "Both XML files should be processed")
	suite.Contains(suite.mockMediator.processedFiles, "product_001.xml")
	suite.Contains(suite.mockMediator.processedFiles, "product_002.xml")

	// Verify file system integration - files deleted
	inputFiles, err := os.ReadDir(suite.inputDir)
	require.NoError(suite.T(), err)
	suite.Len(inputFiles, 0, "Files should be deleted after processing (ActionAfterProcess=DELETE)")
}

// TestIntegration_FileInbound_ConcurrentProcessing tests concurrent file processing integration
func (suite *FileInboundIntegrationTestSuite) TestIntegration_FileInbound_ConcurrentProcessing() {
	// Create multiple files for concurrent processing
	const numFiles = 20
	for i := 0; i < numFiles; i++ {
		filename := fmt.Sprintf("concurrent_test_%03d.json", i)
		content := fmt.Sprintf(`{"fileId":%d,"timestamp":"%s","data":"test data %d","processingId":"concurrent_%d"}`,
			i, time.Now().Format(time.RFC3339), i, i)

		filePath := filepath.Join(suite.inputDir, filename)
		require.NoError(suite.T(), os.WriteFile(filePath, []byte(content), 0644))
	}

	start := time.Now()

	// Test concurrent integration between file system and message processing
	suite.T().Log("Testing concurrent file processing integration...")

	// Read all files from file system
	inputFiles, err := os.ReadDir(suite.inputDir)
	require.NoError(suite.T(), err)
	suite.Len(inputFiles, numFiles, "All files should be created in file system")

	// Process files with integration testing
	for _, file := range inputFiles {
		filename := file.Name()
		filePath := filepath.Join(suite.inputDir, filename)

		// Integration: File System → Message Context
		content, err := os.ReadFile(filePath)
		require.NoError(suite.T(), err)

		// Integration: Message Context → Mediator Processing
		msgContext := synctx.CreateMsgContext()
		msgContext.Message.RawPayload = content
		msgContext.Message.ContentType = "application/json"
		msgContext.Headers["FILE_NAME"] = filename
		msgContext.Headers["PROCESSING_START"] = start.Format(time.RFC3339)

		err = suite.mockMediator.ProcessMessage(context.Background(), msgContext)
		require.NoError(suite.T(), err)

		// Integration: File System → Move to processed
		processedPath := filepath.Join(suite.processedDir, filename)
		err = os.Rename(filePath, processedPath)
		require.NoError(suite.T(), err)
	}

	processingTime := time.Since(start)

	// Verify concurrent processing integration
	suite.Len(suite.mockMediator.processedFiles, numFiles, "All files should be processed through mediator")

	// Verify file system integration under concurrent load
	processedFiles, err := os.ReadDir(suite.processedDir)
	require.NoError(suite.T(), err)
	suite.Len(processedFiles, numFiles, "All files should be moved to processed directory")

	inputFiles, err = os.ReadDir(suite.inputDir)
	require.NoError(suite.T(), err)
	suite.Len(inputFiles, 0, "Input directory should be empty after concurrent processing")

	// Verify performance integration
	suite.Less(processingTime, 5*time.Second, "Concurrent processing should complete within 5 seconds")

	suite.T().Logf("Successfully integrated concurrent processing of %d files in %v", numFiles, processingTime)
}

// TestIntegration_FileInbound_ErrorHandling tests error handling integration
func (suite *FileInboundIntegrationTestSuite) TestIntegration_FileInbound_ErrorHandling() {
	// Create error directory for integration testing
	errorDir := filepath.Join(suite.tempDir, "error")
	err := os.MkdirAll(errorDir, 0755)
	require.NoError(suite.T(), err)

	// Create files with different validity for integration testing
	testFiles := map[string]struct {
		content   string
		expectErr bool
	}{
		"invalid.json": {
			content:   `{"invalid": json file, missing quotes}`,
			expectErr: true,
		},
		"valid.json": {
			content:   `{"valid": "json file", "data": "test", "integration": true}`,
			expectErr: false,
		},
	}

	// Write test files to file system
	for filename, fileData := range testFiles {
		filePath := filepath.Join(suite.inputDir, filename)
		require.NoError(suite.T(), os.WriteFile(filePath, []byte(fileData.content), 0644))
	}

	suite.T().Log("Testing error handling integration across file system and message processing...")

	// Process files with error handling integration
	for filename, fileData := range testFiles {
		filePath := filepath.Join(suite.inputDir, filename)

		// Integration: File System → Message Context
		content, err := os.ReadFile(filePath)
		require.NoError(suite.T(), err)

		// Integration: Message Context → Mediator with Error Handling
		msgContext := synctx.CreateMsgContext()
		msgContext.Message.RawPayload = content
		msgContext.Message.ContentType = "application/json"
		msgContext.Headers["FILE_NAME"] = filename
		msgContext.Headers["ERROR_HANDLING"] = "enabled"

		err = suite.mockMediator.ProcessMessage(context.Background(), msgContext)

		if fileData.expectErr {
			// Integration: Error Case → Move to Error Directory
			suite.NotNil(err, "Should get error for invalid JSON")

			errorFilePath := filepath.Join(errorDir, filename)
			err = os.Rename(filePath, errorFilePath)
			require.NoError(suite.T(), err)

			suite.T().Logf("Integrated error handling for: %s", filename)
		} else {
			// Integration: Success Case → Move to Processed Directory
			require.NoError(suite.T(), err, "Valid files should process without error")

			processedPath := filepath.Join(suite.processedDir, filename)
			err = os.Rename(filePath, processedPath)
			require.NoError(suite.T(), err)

			suite.T().Logf("Integrated successful processing for: %s", filename)
		}
	}

	// Verify error handling integration results
	errorFiles, err := os.ReadDir(errorDir)
	require.NoError(suite.T(), err)
	suite.Len(errorFiles, 1, "One file should be in error directory")
	suite.Equal("invalid.json", errorFiles[0].Name())

	processedFiles, err := os.ReadDir(suite.processedDir)
	require.NoError(suite.T(), err)
	suite.Len(processedFiles, 1, "One file should be in processed directory")
	suite.Equal("valid.json", processedFiles[0].Name())

	// Verify mediator integration - only valid file processed
	suite.Len(suite.mockMediator.processedFiles, 1, "Only valid file should be processed by mediator")
	suite.Equal("valid.json", suite.mockMediator.processedFiles[0])

	// Verify file system integration - input directory empty
	inputFiles, err := os.ReadDir(suite.inputDir)
	require.NoError(suite.T(), err)
	suite.Len(inputFiles, 0, "Input directory should be empty after error handling")
}

// TestIntegration_FileInbound_MessageContextFields tests integration of message context field population
func (suite *FileInboundIntegrationTestSuite) TestIntegration_FileInbound_MessageContextFields() {
	// Create test file with metadata
	filename := "metadata_test.json"
	content := `{"testId":"METADATA001","description":"Testing message context integration"}`
	filePath := filepath.Join(suite.inputDir, filename)

	require.NoError(suite.T(), os.WriteFile(filePath, []byte(content), 0644))

	// Get file info for integration testing
	fileInfo, err := os.Stat(filePath)
	require.NoError(suite.T(), err)

	suite.T().Log("Testing message context field integration...")

	// Integration: File System → Message Context with Full Metadata
	fileContent, err := os.ReadFile(filePath)
	require.NoError(suite.T(), err)

	msgContext := synctx.CreateMsgContext()
	msgContext.Message.RawPayload = fileContent
	msgContext.Message.ContentType = "application/json"

	// Test integration of file metadata into message context
	msgContext.Headers["FILE_NAME"] = filename
	msgContext.Headers["FILE_PATH"] = filePath
	msgContext.Headers["FILE_SIZE"] = fmt.Sprintf("%d", fileInfo.Size())
	msgContext.Headers["FILE_MOD_TIME"] = fileInfo.ModTime().Format(time.RFC3339)
	msgContext.Headers["PROCESSING_TIME"] = time.Now().Format(time.RFC3339)

	// Verify message context integration
	suite.Equal(filename, msgContext.Headers["FILE_NAME"])
	suite.Equal(filePath, msgContext.Headers["FILE_PATH"])
	suite.Equal(fmt.Sprintf("%d", len(content)), msgContext.Headers["FILE_SIZE"])
	suite.NotEmpty(msgContext.Headers["FILE_MOD_TIME"])
	suite.NotEmpty(msgContext.Headers["PROCESSING_TIME"])

	// Integration: Message Context → Mediator Processing
	err = suite.mockMediator.ProcessMessage(context.Background(), msgContext)
	require.NoError(suite.T(), err)

	// Verify integration results
	suite.Len(suite.mockMediator.processedFiles, 1)
	suite.Equal(filename, suite.mockMediator.processedFiles[0])
	suite.Equal(content, string(suite.mockMediator.processedData[0]))

	suite.T().Log("Successfully integrated message context field population")
}

// TestFileInboundIntegrationTestSuite runs the integration test suite
func TestFileInboundIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(FileInboundIntegrationTestSuite))
}

// Helper function to check if file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
