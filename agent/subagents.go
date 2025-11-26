package agent

import (
	"context"
	"fmt"

	markdown "github.com/MichaelMure/go-term-markdown"
	openai "github.com/sashabaranov/go-openai"
	"github.com/smallnest/goskills/tool"
)

// SearchSubagent performs web searches.
type SearchSubagent struct {
	client             *openai.Client
	model              string
	verbose            bool
	interactionHandler InteractionHandler
}

// NewSearchSubagent creates a new SearchSubagent.
func NewSearchSubagent(client *openai.Client, model string, verbose bool, interactionHandler InteractionHandler) *SearchSubagent {
	return &SearchSubagent{
		client:             client,
		model:              model,
		verbose:            verbose,
		interactionHandler: interactionHandler,
	}
}

// Type returns the task type this subagent handles.
func (s *SearchSubagent) Type() TaskType {
	return TaskTypeSearch
}

// Execute performs a web search based on the task.
func (s *SearchSubagent) Execute(ctx context.Context, task Task) (Result, error) {
	if s.verbose {
		fmt.Println("🌐 Web Search Subagent")
	}

	// Extract query from parameters
	query, ok := task.Parameters["query"].(string)
	if !ok {
		query = task.Description
	}

	if s.verbose {
		fmt.Printf("  Query: %q\n", query)
	}

	// Perform Tavily search
	searchResult, err := tool.TavilySearch(query)
	if err != nil {
		// Fallback to DuckDuckGo if Tavily fails (e.g. missing key)
		if s.verbose {
			fmt.Printf("  ⚠️ Tavily search failed: %v. Falling back to DuckDuckGo.\n", err)
		}
		searchResult, err = tool.DuckDuckGoSearch(query)
		if err != nil {
			return Result{
				TaskType: TaskTypeSearch,
				Success:  false,
				Error:    err.Error(),
			}, err
		}
	} else {
		// Human-in-the-loop: Ask if user wants more results
		if s.interactionHandler != nil {
			wantMore, err := s.interactionHandler.ReviewSearchResults(searchResult)
			if err == nil && wantMore {
				if s.verbose {
					fmt.Println("  🔄 User requested more results. Searching up to 100 results...")
				}
				moreResults, err := tool.TavilySearchWithLimit(query, 100)
				if err == nil {
					searchResult = moreResults
				} else {
					if s.verbose {
						fmt.Printf("  ⚠️ Failed to get more results: %v. Keeping original results.\n", err)
					}
				}
			}
		}
	}

	// Also try Wikipedia if results are sparse (optional, keeping existing logic)
	wikiResult, wikiErr := tool.WikipediaSearch(query)
	if wikiErr == nil && wikiResult != "" {
		searchResult = fmt.Sprintf("Web Search Results:\n%s\n\nWikipedia Results:\n%s", searchResult, wikiResult)
	}

	if s.verbose {
		fmt.Printf("  ✓ Retrieved information (%d bytes)\n", len(searchResult))
	}

	return Result{
		TaskType: TaskTypeSearch,
		Success:  true,
		Output:   searchResult,
		Metadata: map[string]interface{}{
			"query": query,
		},
	}, nil
}

// AnalysisSubagent analyzes and synthesizes information.
type AnalysisSubagent struct {
	client  *openai.Client
	model   string
	verbose bool
}

// NewAnalysisSubagent creates a new AnalysisSubagent.
func NewAnalysisSubagent(client *openai.Client, model string, verbose bool) *AnalysisSubagent {
	return &AnalysisSubagent{
		client:  client,
		model:   model,
		verbose: verbose,
	}
}

// Type returns the task type this subagent handles.
func (a *AnalysisSubagent) Type() TaskType {
	return TaskTypeAnalyze
}

// Execute analyzes information using the LLM.
func (a *AnalysisSubagent) Execute(ctx context.Context, task Task) (Result, error) {
	if a.verbose {
		fmt.Println("🔬 Analysis Subagent")
	}

	// Get context from parameters if available
	context_data, hasContext := task.Parameters["context"].(string)

	var prompt string
	if hasContext {
		prompt = fmt.Sprintf("Analyze the following information and %s:\n\n%s", task.Description, context_data)
	} else {
		prompt = task.Description
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are an analytical assistant that synthesizes and analyzes information. Provide clear, structured analysis.",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	req := openai.ChatCompletionRequest{
		Model:       a.model,
		Messages:    messages,
		Temperature: 0.3,
	}

	resp, err := a.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return Result{
			TaskType: TaskTypeAnalyze,
			Success:  false,
			Error:    err.Error(),
		}, err
	}

	analysis := resp.Choices[0].Message.Content

	if a.verbose {
		fmt.Printf("  ✓ Analysis complete (%d bytes)\n", len(analysis))
	}

	return Result{
		TaskType: TaskTypeAnalyze,
		Success:  true,
		Output:   analysis,
	}, nil
}

// ReportSubagent generates formatted reports.
type ReportSubagent struct {
	client  *openai.Client
	model   string
	verbose bool
}

// NewReportSubagent creates a new ReportSubagent.
func NewReportSubagent(client *openai.Client, model string, verbose bool) *ReportSubagent {
	return &ReportSubagent{
		client:  client,
		model:   model,
		verbose: verbose,
	}
}

// Type returns the task type this subagent handles.
func (r *ReportSubagent) Type() TaskType {
	return TaskTypeReport
}

// Execute generates a formatted report.
func (r *ReportSubagent) Execute(ctx context.Context, task Task) (Result, error) {
	if r.verbose {
		fmt.Println("📝 Report Subagent")
	}

	// Get context from parameters if available
	context_data, hasContext := task.Parameters["context"].(string)

	var prompt string
	if hasContext {
		prompt = fmt.Sprintf("Based on the following information, %s:\n\n%s", task.Description, context_data)
	} else {
		prompt = task.Description
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are a report writing assistant that creates well-formatted, clear, and comprehensive reports in Markdown format. Use appropriate headings, lists, and formatting to make the report easy to read.",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	req := openai.ChatCompletionRequest{
		Model:       r.model,
		Messages:    messages,
		Temperature: 0.5,
	}

	resp, err := r.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return Result{
			TaskType: TaskTypeReport,
			Success:  false,
			Error:    err.Error(),
		}, err
	}

	report := resp.Choices[0].Message.Content

	if r.verbose {
		fmt.Printf("  ✓ Report generated (%d bytes)\n", len(report))
	}

	return Result{
		TaskType: TaskTypeReport,
		Success:  true,
		Output:   report,
	}, nil
}

// RenderSubagent renders markdown to terminal-friendly format.
type RenderSubagent struct {
	verbose bool
}

// NewRenderSubagent creates a new RenderSubagent.
func NewRenderSubagent(verbose bool) *RenderSubagent {
	return &RenderSubagent{
		verbose: verbose,
	}
}

// Type returns the task type this subagent handles.
func (r *RenderSubagent) Type() TaskType {
	return TaskTypeRender
}

// Execute renders markdown content.
func (r *RenderSubagent) Execute(ctx context.Context, task Task) (Result, error) {
	if r.verbose {
		fmt.Println("🎨 Render Subagent")
	}

	// Get content from parameters or description
	content, ok := task.Parameters["content"].(string)
	if !ok {
		content = task.Description
	}

	if r.verbose {
		fmt.Printf("  Rendering %d bytes of content\n", len(content))
	}

	// Render markdown
	result := markdown.Render(content, 80, 6)

	return Result{
		TaskType: TaskTypeRender,
		Success:  true,
		Output:   string(result),
	}, nil
}
