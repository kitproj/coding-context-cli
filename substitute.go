package main

import (
	"os"
	"strings"
)

// substituteVariables replaces VS Code variable syntax ${var} and ${input:var} with their values
// Uses os.Expand for variable substitution
func substituteVariables(content string, params map[string]string) string {
	// First pass: convert ${input:varName} and ${input:varName:placeholder} to ${varName}
	// This needs to be done before os.Expand since os.Expand doesn't handle the input: prefix
	result := content
	for {
		start := strings.Index(result, "${input:")
		if start == -1 {
			break
		}
		
		// Find the closing }
		end := start + 8
		for end < len(result) && result[end] != '}' {
			end++
		}
		if end >= len(result) {
			break
		}
		
		// Extract the content between ${input: and }
		varPart := result[start+8 : end]
		// Split by : to get variable name (ignore placeholder if present)
		colonIdx := strings.Index(varPart, ":")
		var varName string
		if colonIdx >= 0 {
			varName = varPart[:colonIdx]
		} else {
			varName = varPart
		}
		
		// Replace ${input:varName...} with ${varName}
		result = result[:start] + "${" + varName + "}" + result[end+1:]
	}
	
	// Second pass: use os.Expand to substitute variables
	return os.Expand(result, func(varName string) string {
		// Return the parameter value, or empty string if not found
		return params[varName]
	})
}
