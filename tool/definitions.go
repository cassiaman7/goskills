package tool

import (
	openai "github.com/sashabaranov/go-openai"
)

// GetBaseTools returns the list of base tools available to all skills.
func GetBaseTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "run_shell_code",
				Description: "Executes a shell code snippet and returns its combined stdout and stderr.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"code": map[string]interface{}{
							"type":        "string",
							"description": "The shell code snippet to execute.",
						},
						"args": map[string]interface{}{
							"type":        "object",
							"description": "A map of key-value pairs to pass to the code.",
						},
					},
					"required": []string{"code"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "run_shell_script",
				Description: "Executes a shell script and returns its combined stdout and stderr. Use this for general shell commands.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"scriptPath": map[string]interface{}{
							"type":        "string",
							"description": "The path to the shell script to execute.",
						},
						"args": map[string]interface{}{
							"type":        "array",
							"description": "A list of string arguments to pass to the script.",
							"items": map[string]interface{}{
								"type": "string",
							},
						},
					},
					"required": []string{"scriptPath"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "run_python_code",
				Description: "Executes a Python code snippet and returns its combined stdout and stderr.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"code": map[string]interface{}{
							"type":        "string",
							"description": "The Python code snippet to execute.",
						},
						"args": map[string]interface{}{
							"type":        "object",
							"description": "A map of key-value pairs to pass to the code.",
						},
					},
					"required": []string{"code"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "run_python_script",
				Description: "Executes a Python script and returns its combined stdout and stderr.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"scriptPath": map[string]interface{}{
							"type":        "string",
							"description": "The path to the Python script to execute.",
						},
						"args": map[string]interface{}{
							"type":        "array",
							"description": "A list of string arguments to pass to the script.",
							"items": map[string]interface{}{
								"type": "string",
							},
						},
					},
					"required": []string{"scriptPath"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "read_file",
				Description: "Reads the content of a file and returns it as a string.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"filePath": map[string]interface{}{
							"type":        "string",
							"description": "The path to the file to read.",
						},
					},
					"required": []string{"filePath"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "write_file",
				Description: "Writes the given content to a file. If the file does not exist, it will be created. If it exists, its content will be truncated.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"filePath": map[string]interface{}{
							"type":        "string",
							"description": "The path to the file to write.",
						},
						"content": map[string]interface{}{
							"type":        "string",
							"description": "The content to write to the file.",
						},
					},
					"required": []string{"filePath", "content"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "duckduckgo_search",
				Description: "Performs a DuckDuckGo search for the given query and returns a summary or related topics.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "The search query.",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "wikipedia_search",
				Description: "Performs a search on Wikipedia for the given query and returns a summary of the relevant entry.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "The search query for Wikipedia.",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		// {
		// 	Type: openai.ToolTypeFunction,
		// 	Function: &openai.FunctionDefinition{
		// 		Name:        "web_fetch",
		// 		Description: "Fetches the clean text content from a given URL. It automatically parses the HTML and returns only the readable text.",
		// 		Parameters: map[string]interface{}{
		// 			"type": "object",
		// 			"properties": map[string]interface{}{
		// 				"url": map[string]interface{}{
		// 					"type":        "string",
		// 					"description": "The full URL to fetch, including the protocol (e.g., 'https://example.com').",
		// 				},
		// 			},
		// 			"required": []string{"url"},
		// 		},
		// 	},
		// },
	}
}
