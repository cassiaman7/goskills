package goskills

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/smallnest/goskills/tool"
)

// RunnerConfig holds all the necessary configuration for the runner.
type RunnerConfig struct {
	APIKey           string
	APIBase          string
	Model            string
	SkillsDir        string
	Verbose          bool
	AutoApproveTools bool
	AllowedScripts   []string
}

// Run executes the main skill selection and execution logic.
func Run(ctx context.Context, userPrompt string, cfg RunnerConfig) (string, error) {
	// --- PRE-FLIGHT CHECK ---
	if cfg.APIKey == "" {
		return "", errors.New("API key is not set")
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4o" // Default model
	}

	openaiConfig := openai.DefaultConfig(cfg.APIKey)
	if cfg.APIBase != "" {
	openaiConfig.BaseURL = cfg.APIBase
	}
	client := openai.NewClientWithConfig(openaiConfig)

	// --- STEP 1: SKILL DISCOVERY ---
	if cfg.Verbose {
		fmt.Printf("🔎 Discovering available skills in %s...\n", cfg.SkillsDir)
	}
	availableSkills, err := discoverSkills(cfg.SkillsDir)
	if err != nil {
		return "", fmt.Errorf("failed to discover skills: %w", err)
	}
	if len(availableSkills) == 0 {
		return "", errors.New("no valid skills found")
	}
	if cfg.Verbose {
		fmt.Printf("✅ Found %d skills.\n\n", len(availableSkills))
	}

	// --- STEP 2: SKILL SELECTION ---
	if cfg.Verbose {
		fmt.Println("🧠 Asking LLM to select the best skill...")
	}
	selectedSkillName, err := selectSkill(ctx, client, cfg.Model, userPrompt, availableSkills)
	if err != nil {
		return "", fmt.Errorf("failed during skill selection: %w", err)
	}

	selectedSkill, ok := availableSkills[selectedSkillName]
	if !ok {
		return "", fmt.Errorf("⚠️ LLM selected a non-existent skill '%s'. Aborting", selectedSkillName)
	}
	if cfg.Verbose {
		fmt.Printf("✅ LLM selected skill: %s\n\n", selectedSkillName)
	}

	// --- STEP 3: SKILL EXECUTION (with Tool Calling) ---
	if cfg.Verbose {
		fmt.Println("🚀 Executing skill (with potential tool calls).")
		fmt.Println(strings.Repeat("-", 40))
	}

	finalOutput, err := executeSkillWithTools(ctx, client, userPrompt, selectedSkill, cfg)
	if err != nil {
		return "", fmt.Errorf("failed during skill execution: %w", err)
	}

	return finalOutput, nil
}

func discoverSkills(skillsRoot string) (map[string]SkillPackage, error) {
	packages, err := ParseSkillPackages(skillsRoot)
	if err != nil {
		return nil, err
	}

	skills := make(map[string]SkillPackage, len(packages))
	for _, pkg := range packages {
		if pkg != nil {
			skills[pkg.Meta.Name] = *pkg
		}
	}

	return skills, nil
}

func selectSkill(ctx context.Context, client *openai.Client, model, userPrompt string, skills map[string]SkillPackage) (string, error) {
	var sb strings.Builder
	sb.WriteString("User Request: " + "" + userPrompt + "" + "\n\n")
	sb.WriteString("Available Skills:\n")
	for name, skill := range skills {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", name, skill.Meta.Description))
	}
	sb.WriteString("\nBased on the user request, which single skill is the most appropriate to use? Respond with only the name of the skill.")

	req := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an expert assistant that selects the most appropriate skill to handle a user's request. Your response must be only the exact name of the chosen skill, with no other text or explanation.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: sb.String(),
			},
		},
		Temperature: 0,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	skillName := strings.TrimSpace(resp.Choices[0].Message.Content)
	skillName = strings.Trim(skillName, "'\"")

	return skillName, nil
}

func executeSkillWithTools(ctx context.Context, client *openai.Client, userPrompt string, skill SkillPackage, cfg RunnerConfig) (string, error) {
	var skillBody strings.Builder
	skillBody.WriteString(skill.Body)
	skillBody.WriteString("\n\n## SKILL CONTEXT\n")
	skillBody.WriteString(fmt.Sprintf("Skill Root Path: %s\n", skill.Path))
	// ... (rest of skill context)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: skillBody.String(),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userPrompt,
		},
	}

	availableTools, scriptMap := GenerateToolDefinitions(skill)
	var finalResponse strings.Builder

	for i := 0; i < 10; i++ { // Limit to 10 iterations to prevent infinite loops
		req := openai.ChatCompletionRequest{
			Model:    cfg.Model,
			Messages: messages,
			Tools:    availableTools,
		}

		resp, err := client.CreateChatCompletion(ctx, req)
		if err != nil {
			return "", fmt.Errorf("ChatCompletion error: %w", err)
		}

		msg := resp.Choices[0].Message
		messages = append(messages, msg)

		if msg.ToolCalls == nil {
			finalResponse.WriteString(msg.Content)
			return finalResponse.String(), nil
		}

		// Parallel execution of tool calls could be implemented here
		for _, tc := range msg.ToolCalls {
			if cfg.Verbose {
				fmt.Printf("⚙️ Calling tool: %s with args: %s\n", tc.Function.Name, tc.Function.Arguments)
			}

			// --- SECURITY CHECK ---
			if !cfg.AutoApproveTools {
				fmt.Print("⚠️  Allow this tool execution? [y/N]: ")
				var input string
				fmt.Scanln(&input)
				if strings.ToLower(input) != "y" {
					fmt.Println("❌ Tool execution denied by user.")
					messages = append(messages, openai.ChatCompletionMessage{
						Role:       openai.ChatMessageRoleTool,
						ToolCallID: tc.ID,
						Content:    "Error: User denied tool execution.",
					})
					continue
				}
			}

			toolOutput, err := executeToolCall(tc, scriptMap, skill.Path)
			if err != nil {
				fmt.Printf("❌ Tool call failed: %v\n", err)
				messages = append(messages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					ToolCallID: tc.ID,
					Content:    fmt.Sprintf("Error: %v", err),
				})
			} else {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					ToolCallID: tc.ID,
					Content:    toolOutput,
				})
			}
		}
	}
	return "", errors.New("exceeded maximum tool call iterations")
}

func executeToolCall(toolCall openai.ToolCall, scriptMap map[string]string, skillPath string) (string, error) {
	var toolOutput string
	var err error

	switch toolCall.Function.Name {
	case "run_shell_code":
		var params struct {
			Code string         `json:"code"`
			Args map[string]any `json:"args"`
		}
		if err = json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal run_shell_code arguments: %w", err)
		}
		shellTool := tool.ShellTool{}
		toolOutput, err = shellTool.Run(params.Args, params.Code)
	case "run_shell_script":
		var params struct {
			ScriptPath string   `json:"scriptPath"`
			Args       []string `json:"args"`
		}
		if err = json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal run_shell_script arguments: %w", err)
		}
		toolOutput, err = tool.RunShellScript(params.ScriptPath, params.Args)
	case "run_python_code":
		var params struct {
			Code string         `json:"code"`
			Args map[string]any `json:"args"`
		}
		if err = json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal run_python_code arguments: %w", err)
		}
		pythonTool := tool.PythonTool{}
		toolOutput, err = pythonTool.Run(params.Args, params.Code)
	case "run_python_script":
		var params struct {
			ScriptPath string   `json:"scriptPath"`
			Args       []string `json:"args"`
		}
		if err = json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal run_python_script arguments: %w", err)
		}
		toolOutput, err = tool.RunPythonScript(params.ScriptPath, params.Args)
	case "read_file":
		var params struct {
			FilePath string `json:"filePath"`
		}
		if err = json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal read_file arguments: %w", err)
		}
		path := params.FilePath
		if !filepath.IsAbs(path) && skillPath != "" {
			resolvedPath := filepath.Join(skillPath, path)
			if _, err := os.Stat(resolvedPath); err == nil {
				path = resolvedPath
			}
		}
		toolOutput, err = tool.ReadFile(path)
	case "write_file":
		var params struct {
			FilePath string `json:"filePath"`
			Content  string `json:"content"`
		}
		if err = json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal write_file arguments: %w", err)
		}
		err = tool.WriteFile(params.FilePath, params.Content)
		if err == nil {
			toolOutput = fmt.Sprintf("Successfully wrote to file: %s", params.FilePath)
		}
	case "duckduckgo_search":
		var params struct {
			Query string `json:"query"`
		}
		if err = json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal duckduckgo_search arguments: %w", err)
		}
		toolOutput, err = tool.DuckDuckGoSearch(params.Query)
	case "wikipedia_search":
		var params struct {
			Query string `json:"query"`
		}
		if err = json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal wikipedia_search arguments: %w", err)
		}
		toolOutput, err = tool.WikipediaSearch(params.Query)
	case "web_fetch":
		var params struct {
			URL string `json:"url"`
		}
		if err = json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal web_fetch arguments: %w", err)
		}
		toolOutput, err = tool.WebFetch(params.URL)
	default:
		if scriptPath, ok := scriptMap[toolCall.Function.Name]; ok {
			var params struct {
				Args []string `json:"args"`
			}
			if toolCall.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
					return "", fmt.Errorf("failed to unmarshal script arguments: %w", err)
				}
			}
			if strings.HasSuffix(scriptPath, ".py") {
				toolOutput, err = tool.RunPythonScript(scriptPath, params.Args)
			} else {
				toolOutput, err = tool.RunShellScript(scriptPath, params.Args)
			}
		} else {
			return "", fmt.Errorf("unknown tool: %s", toolCall.Function.Name)
		}
	}

	if err != nil {
		return "", fmt.Errorf("tool execution failed for %s: %w", toolCall.Function.Name, err)
	}
	return toolOutput, nil
}
