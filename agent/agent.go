package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// PlanningAgent orchestrates task planning and subagent execution.
type PlanningAgent struct {
	client             *openai.Client
	config             AgentConfig
	messages           []openai.ChatCompletionMessage
	subagents          map[TaskType]Subagent
	interactionHandler InteractionHandler
}

// AgentConfig holds the configuration for the planning agent.
type AgentConfig struct {
	APIKey  string
	APIBase string
	Model   string
	Verbose bool
}

// NewPlanningAgent creates and initializes a new PlanningAgent.
func NewPlanningAgent(config AgentConfig, interactionHandler InteractionHandler) (*PlanningAgent, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if config.Model == "" {
		config.Model = "gpt-4o" // Default model
	}

	openaiConfig := openai.DefaultConfig(config.APIKey)
	if config.APIBase != "" {
		openaiConfig.BaseURL = config.APIBase
	}
	client := openai.NewClientWithConfig(openaiConfig)

	agent := &PlanningAgent{
		client:             client,
		config:             config,
		messages:           []openai.ChatCompletionMessage{},
		subagents:          make(map[TaskType]Subagent),
		interactionHandler: interactionHandler,
	}

	// Initialize subagents
	agent.subagents[TaskTypeSearch] = NewSearchSubagent(client, config.Model, config.Verbose, interactionHandler)
	agent.subagents[TaskTypeAnalyze] = NewAnalysisSubagent(client, config.Model, config.Verbose)
	agent.subagents[TaskTypeReport] = NewReportSubagent(client, config.Model, config.Verbose)
	agent.subagents[TaskTypeRender] = NewRenderSubagent(config.Verbose)
	agent.subagents[TaskTypePodcast] = NewPodcastSubagent(client, config.Model, config.Verbose)

	return agent, nil
}

// Plan decomposes a user request into subtasks.
func (a *PlanningAgent) Plan(ctx context.Context, userRequest string) (*Plan, error) {
	systemPrompt := `You are a planning agent that breaks down user requests into subtasks.
You have access to three types of subagents:
- SEARCH: Performs web searches to gather information
- ANALYZE: Analyzes and synthesizes gathered information
- REPORT: Generates formatted reports from analyzed data
- RENDER: Renders markdown content to terminal-friendly format

For the given user request, create a plan with a sequence of tasks.
Each task should have:
- type: one of SEARCH, ANALYZE, REPORT, or RENDER
- description: what the subagent should do
- parameters: optional parameters for the task (e.g., {"query": "search term"})

Return ONLY a valid JSON object with this structure:
{
  "description": "Overall plan description",
  "tasks": [
    {"type": "SEARCH", "description": "...", "parameters": {"query": "..."}},
    {"type": "ANALYZE", "description": "..."},
    {"type": "REPORT", "description": "..."},
    {"type": "RENDER", "description": "Render the report"}
  ]
}

Keep plans simple and focused. Typically 2-4 tasks are sufficient.`

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("Create a plan for this request: %s", userRequest),
		},
	}

	req := openai.ChatCompletionRequest{
		Model:       a.config.Model,
		Messages:    messages,
		Temperature: 0,
	}

	resp, err := a.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	content := resp.Choices[0].Message.Content

	// Clean up the content if it contains markdown code blocks
	if len(content) > 0 {
		// Remove ```json prefix if present
		if idx := strings.Index(content, "```json"); idx != -1 {
			content = content[idx+7:]
		} else if idx := strings.Index(content, "```"); idx != -1 {
			content = content[idx+3:]
		}

		// Remove closing ``` if present
		if idx := strings.LastIndex(content, "```"); idx != -1 {
			content = content[:idx]
		}

		content = strings.TrimSpace(content)
	}

	// Parse the JSON response
	var plan Plan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan JSON: %w\nResponse: %s", err, content)
	}

	if a.config.Verbose {
		fmt.Println("🧠 Planning Agent")
		fmt.Printf("📋 Plan: %s\n", plan.Description)
		for i, task := range plan.Tasks {
			fmt.Printf("  %d. [%s] %s\n", i+1, task.Type, task.Description)
		}
		fmt.Println()
	}

	return &plan, nil
}

// PlanWithReview creates a plan and optionally allows the user to review and modify it.
func (a *PlanningAgent) PlanWithReview(ctx context.Context, userRequest string) (*Plan, error) {
	// Create initial plan
	plan, err := a.Plan(ctx, userRequest)
	if err != nil {
		return nil, err
	}

	// If no interaction handler, return the plan as-is
	if a.interactionHandler == nil {
		return plan, nil
	}

	// Allow user to review and modify the plan
	for {
		modification, err := a.interactionHandler.ReviewPlan(plan)
		if err != nil {
			return nil, fmt.Errorf("plan review failed: %w", err)
		}

		// If no modification requested, use the current plan
		if modification == "" {
			break
		}

		// Re-plan with the user's modification
		if a.config.Verbose {
			fmt.Printf("🔄 Re-planning based on user feedback: %s\n\n", modification)
		}

		plan, err = a.Plan(ctx, modification)
		if err != nil {
			return nil, fmt.Errorf("re-planning failed: %w", err)
		}
	}

	return plan, nil
}

// Execute runs the plan by executing each task with the appropriate subagent.
func (a *PlanningAgent) Execute(ctx context.Context, plan *Plan) ([]Result, error) {
	if a.config.Verbose {
		fmt.Println("🔍 Executing plan...")
		fmt.Println()
	}

	results := make([]Result, 0, len(plan.Tasks))

	var contextData []string

	for i, task := range plan.Tasks {
		if a.config.Verbose {
			fmt.Printf("📍 Step %d/%d: [%s] %s\n", i+1, len(plan.Tasks), task.Type, task.Description)
		}

		// Inject context from previous tasks
		if len(contextData) > 0 {
			if task.Parameters == nil {
				task.Parameters = make(map[string]interface{})
			}
			// If context already exists in parameters, append to it
			if existingContext, ok := task.Parameters["context"].([]string); ok {
				task.Parameters["context"] = append(existingContext, contextData...)
			} else {
				task.Parameters["context"] = contextData
			}
		}

		subagent, ok := a.subagents[task.Type]
		if !ok {
			return nil, fmt.Errorf("unknown task type: %s", task.Type)
		}

		result, err := subagent.Execute(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("task %d failed: %w", i+1, err)
		}

		results = append(results, result)

		if result.Success {
			// Accumulate output for next tasks
			contextData = append(contextData, fmt.Sprintf("Output from %s task:\n%s", task.Type, result.Output))

			if a.config.Verbose {
				fmt.Printf("  ✓ Completed\n\n")
			}
		} else {
			if a.config.Verbose {
				fmt.Printf("  ✗ Failed: %s\n\n", result.Error)
			}
		}
	}

	return results, nil
}

// Run is the main entry point that plans and executes a user request.
func (a *PlanningAgent) Run(ctx context.Context, userRequest string) (string, error) {
	// Create a plan
	plan, err := a.Plan(ctx, userRequest)
	if err != nil {
		return "", err
	}

	// Execute the plan
	results, err := a.Execute(ctx, plan)
	if err != nil {
		return "", err
	}

	// Extract the final output (typically from the RENDER or REPORT task)
	var finalOutput string
	for i := len(results) - 1; i >= 0; i-- {
		if (results[i].TaskType == TaskTypeRender || results[i].TaskType == TaskTypeReport) && results[i].Success {
			finalOutput = results[i].Output
			break
		}
	}

	// If no report was generated, concatenate all outputs
	if finalOutput == "" {
		for _, result := range results {
			if result.Success {
				finalOutput += result.Output + "\n\n"
			}
		}
	}

	return finalOutput, nil
}

// RunInteractive is similar to Run but maintains conversation history.
func (a *PlanningAgent) RunInteractive(ctx context.Context, userRequest string) (string, error) {
	// Add user message to history
	a.messages = append(a.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userRequest,
	})

	// For interactive mode, we run the same planning logic
	result, err := a.Run(ctx, userRequest)
	if err != nil {
		return "", err
	}

	// Add assistant response to history
	a.messages = append(a.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: result,
	})

	return result, nil
}

// AddUserMessage adds a user message to the conversation history.
func (a *PlanningAgent) AddUserMessage(content string) {
	a.messages = append(a.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: content,
	})
}

// AddAssistantMessage adds an assistant message to the conversation history.
func (a *PlanningAgent) AddAssistantMessage(content string) {
	a.messages = append(a.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: content,
	})
}

// ClearHistory clears the conversation history.
func (a *PlanningAgent) ClearHistory() {
	a.messages = []openai.ChatCompletionMessage{}
}
