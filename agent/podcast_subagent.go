package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// PodcastSubagent generates a podcast from a report.
type PodcastSubagent struct {
	client  *openai.Client
	model   string
	verbose bool
}

// NewPodcastSubagent creates a new PodcastSubagent.
func NewPodcastSubagent(client *openai.Client, model string, verbose bool) *PodcastSubagent {
	return &PodcastSubagent{
		client:  client,
		model:   model,
		verbose: verbose,
	}
}

// Type returns the task type this subagent handles.
func (p *PodcastSubagent) Type() TaskType {
	return TaskTypePodcast
}

// DialogueLine represents a single line of dialogue in the podcast.
type DialogueLine struct {
	Speaker string `json:"speaker"`
	Text    string `json:"text"`
}

// Execute generates a podcast from the input content.
func (p *PodcastSubagent) Execute(ctx context.Context, task Task) (Result, error) {
	if p.verbose {
		fmt.Println("🎙️ Podcast Subagent")
	}

	// Get content from parameters or description
	content, ok := task.Parameters["content"].(string)
	if !ok {
		content = task.Description
	}

	if p.verbose {
		fmt.Println("  Generating dialogue script...")
	}

	// 1. Generate Dialogue Script
	script, err := p.generateScript(ctx, content)
	if err != nil {
		return Result{
			TaskType: TaskTypePodcast,
			Success:  false,
			Error:    fmt.Sprintf("failed to generate script: %v", err),
		}, err
	}

	if p.verbose {
		fmt.Printf("  ✓ Script generated (%d lines)\n", len(script))
	}

	// Convert script to JSON string for output
	scriptJSON, err := json.MarshalIndent(script, "", "  ")
	if err != nil {
		return Result{
			TaskType: TaskTypePodcast,
			Success:  false,
			Error:    fmt.Sprintf("failed to marshal script: %v", err),
		}, err
	}

	outputMsg := fmt.Sprintf("Podcast script generated successfully!\n\nPlease submit the following script to https://listenhub.ai/zh to generate the audio:\n\n%s", string(scriptJSON))

	return Result{
		TaskType: TaskTypePodcast,
		Success:  true,
		Output:   outputMsg,
		Metadata: map[string]interface{}{
			"script": script,
		},
	}, nil
}

func (p *PodcastSubagent) generateScript(ctx context.Context, content string) ([]DialogueLine, error) {
	systemPrompt := `You are a podcast producer. Your goal is to convert the provided input text (a report or article) into an engaging dialogue between two hosts:
- Host 1 (Male): Enthusiastic, curious, asks questions, and introduces topics.
- Host 2 (Female): Knowledgeable, calm, explains details, and provides insights.

The dialogue should be natural, conversational, and easy to listen to. It should cover the main points of the input text.
Output ONLY a JSON array of objects, where each object has "speaker" ("Host 1" or "Host 2") and "text" (the spoken line).
Example:
[
  {"speaker": "Host 1", "text": "Welcome back to the show! Today we're discussing..."},
  {"speaker": "Host 2", "text": "That's right. It's a fascinating topic..."}
]`

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("Convert this text into a podcast dialogue (输出中文):\n\n%s", content),
		},
	}

	req := openai.ChatCompletionRequest{
		Model:       p.model,
		Messages:    messages,
		Temperature: 0.7,
	}

	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	scriptContent := resp.Choices[0].Message.Content

	// Clean up markdown code blocks if present
	if idx := strings.Index(scriptContent, "```json"); idx != -1 {
		scriptContent = scriptContent[idx+7:]
	} else if idx := strings.Index(scriptContent, "```"); idx != -1 {
		scriptContent = scriptContent[idx+3:]
	}
	if idx := strings.LastIndex(scriptContent, "```"); idx != -1 {
		scriptContent = scriptContent[:idx]
	}
	scriptContent = strings.TrimSpace(scriptContent)

	var script []DialogueLine
	if err := json.Unmarshal([]byte(scriptContent), &script); err != nil {
		return nil, fmt.Errorf("failed to parse script JSON: %w", err)
	}

	return script, nil
}
