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

package consolelogger

import (
	"fmt"
)

// ANSI escape codes for rainbow colors
var colors = []string{
	"\033[31m", // Red
	"\033[33m", // Yellow
	"\033[32m", // Green
	"\033[36m", // Cyan
	"\033[34m", // Blue
	"\033[35m", // Magenta
}

// Reset ANSI code
const reset = "\033[0m"

func PrintWelcomeMessage() {
	colors := []string{
		"\033[31m", // Red
		"\033[33m", // Yellow
		"\033[32m", // Green
		"\033[36m", // Cyan
		"\033[34m", // Blue
		"\033[35m", // Magenta
	}

	// ANSI code to reset color to default
	reset := "\033[0m"

	art := []string{
		"",
		"      _/_/_/                                                             ",
		"   _/        _/    _/  _/_/_/      _/_/_/  _/_/_/      _/_/_/    _/_/    ",
		"    _/_/    _/    _/  _/    _/  _/    _/  _/    _/  _/_/      _/_/_/_/   ",
		"       _/  _/    _/  _/    _/  _/    _/  _/    _/      _/_/  _/          ",
		"_/_/_/      _/_/_/  _/    _/    _/_/_/  _/_/_/    _/_/_/      _/_/_/     ",
		"               _/                      _/                                ",
		"          _/_/                        _/                                 ",
	}
	// Iterate over each line of the ASCII art
	for _, line := range art {
		// Iterate over each character in the line
		for i, char := range line {
			// Select color based on character position to create a gradient
			color := colors[i%len(colors)]
			// Print the colored character without adding a newline
			fmt.Printf("%s%c", color, char)
		}
		// Reset color at the end of each line and add a newline
		fmt.Println(reset)
	}
}

func InfoLog(message string) {
	fmt.Println(colors[4] + "INFO: " + message + reset)
}

func DebugLog(message string) {
	fmt.Println(colors[1] + "DEBUG: " + message + reset)
}

func ErrorLog(message string) {
	fmt.Println(colors[0] + "ERROR: " + message + reset)
}

func MediatorErrorLog(mediatorName string, fileName string, lineNo int, message string) {
	fmt.Println(colors[5] + "Error occurred in " + mediatorName + " mediator at " + fileName + ":" +
		fmt.Sprint(lineNo) + " " + message + reset)
}
