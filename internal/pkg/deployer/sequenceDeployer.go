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

package deployer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/apache/synapse-go/internal/pkg/artifacts"
	"github.com/apache/synapse-go/internal/pkg/consolelogger"
)

func DeploySequences(folderPath string, confContext *artifacts.ConfigContext) error {

	files, err := os.ReadDir(folderPath)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".xml" {
			continue
		}

		xmlFile, err := os.Open(filepath.Join(folderPath, file.Name()))
		if err != nil {
			fmt.Println("Error opening file:", err)
			continue
		}
		defer xmlFile.Close()

		data, err := io.ReadAll(xmlFile)
		if err != nil {
			fmt.Println("Error reading file:", err)
			continue
		}

		consolelogger.InfoLog("Deploying sequence from " + file.Name())

		var sequence artifacts.NamedSequence
		if err := sequence.UnmarshalArtifact(string(data), artifacts.ArtifactPosition{FileName: file.Name()}); err != nil {
			fmt.Println("Error unmarshaling XML:", err)
			continue
		}

		_, ok := confContext.SequenceMap[sequence.Name]
		if ok {
			consolelogger.ErrorLog("Sequence " + sequence.Name + " already deployed")
			continue
		}

		confContext.AddSequence(sequence)
		consolelogger.InfoLog("Sequence " + sequence.Name + " deployed successfully")
		return nil
	}
	return nil
}
