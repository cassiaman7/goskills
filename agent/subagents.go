package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/smallnest/goskills/tool"

	markdown "github.com/MichaelMure/go-term-markdown"
	gomarkdown "github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	openai "github.com/sashabaranov/go-openai"
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
		fmt.Println("🌐 网络搜索子Agent")
	}
	if s.interactionHandler != nil {
		s.interactionHandler.Log(fmt.Sprintf("> 网络搜索子Agent: %s", task.Description))
	}

	// Extract query from parameters
	query, ok := task.Parameters["query"].(string)
	if !ok {
		query = task.Description
	}

	if s.verbose {
		fmt.Printf("  查询: %q\n", query)
	}

	// Perform Tavily search
	searchResult, err := tool.TavilySearch(query)
	if err != nil {
		// Fallback to DuckDuckGo if Tavily fails (e.g. missing key)
		if s.verbose {
			fmt.Printf("  ⚠️ Tavily 搜索失败: %v。回退到 DuckDuckGo。\n", err)
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
					fmt.Println("  🔄 用户请求更多结果。正在搜索最多 50 条结果...")
				}
				moreResults, err := tool.TavilySearchWithLimit(query, 50)
				if err == nil {
					searchResult = moreResults
					if s.verbose {
						preview := moreResults
						if len(preview) > 500 {
							preview = preview[:500] + "..."
						}
						fmt.Printf("  🔎 新结果预览:\n%s\n", preview)
					}
				} else {
					if s.verbose {
						fmt.Printf("  ⚠️ 获取更多结果失败: %v。保留原始结果。\n", err)
					}
				}
			}
		}
	}

	// Also try Wikipedia if results are sparse (optional, keeping existing logic)
	wikiResult, wikiErr := tool.WikipediaSearch(query)
	if wikiErr == nil && wikiResult != "" {
		searchResult = fmt.Sprintf("网络搜索结果:\n%s\n\n维基百科结果:\n%s", searchResult, wikiResult)
	}

	if s.verbose {
		fmt.Printf("\n  ✓ 已检索信息 (%d 字节)\n", len(searchResult))
	}
	if s.interactionHandler != nil {
		s.interactionHandler.Log(fmt.Sprintf("✓ 已检索信息 (%d 字节)", len(searchResult)))
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
	client             *openai.Client
	model              string
	verbose            bool
	interactionHandler InteractionHandler
}

// NewAnalysisSubagent creates a new AnalysisSubagent.
func NewAnalysisSubagent(client *openai.Client, model string, verbose bool, interactionHandler InteractionHandler) *AnalysisSubagent {
	return &AnalysisSubagent{
		client:             client,
		model:              model,
		verbose:            verbose,
		interactionHandler: interactionHandler,
	}
}

// Type returns the task type this subagent handles.
func (a *AnalysisSubagent) Type() TaskType {
	return TaskTypeAnalyze
}

// Execute analyzes information using the LLM.
func (a *AnalysisSubagent) Execute(ctx context.Context, task Task) (Result, error) {
	if a.verbose {
		fmt.Println("🔬 分析子Agent")
	}
	if a.interactionHandler != nil {
		a.interactionHandler.Log(fmt.Sprintf("> 分析子Agent: %s", task.Description))
	}

	// Get context from parameters if available
	contextData, hasContext := task.Parameters["context"].([]string)

	var prompt string
	if hasContext && len(contextData) > 0 {
		prompt = fmt.Sprintf("分析以下信息并 %s:\n\n%s", task.Description, strings.Join(contextData, "\n\n"))
	} else {
		prompt = task.Description
	}

	// Check for global context
	globalContext, _ := task.Parameters["global_context"].(string)
	systemPrompt := "你是一个分析助手，负责综合和分析信息。请提供清晰、结构化的分析。"
	if globalContext != "" {
		systemPrompt += "\n\n来自用户的重要上下文/指令：\n" + globalContext
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
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
		fmt.Printf("  ✓ 分析完成 (%d 字节)\n", len(analysis))
	}
	if a.interactionHandler != nil {
		a.interactionHandler.Log(fmt.Sprintf("✓ 分析完成 (%d 字节)", len(analysis)))
	}

	return Result{
		TaskType: TaskTypeAnalyze,
		Success:  true,
		Output:   analysis,
	}, nil
}

// ReportSubagent generates formatted reports.
type ReportSubagent struct {
	client             *openai.Client
	model              string
	verbose            bool
	interactionHandler InteractionHandler
}

// NewReportSubagent creates a new ReportSubagent.
func NewReportSubagent(client *openai.Client, model string, verbose bool, interactionHandler InteractionHandler) *ReportSubagent {
	return &ReportSubagent{
		client:             client,
		model:              model,
		verbose:            verbose,
		interactionHandler: interactionHandler,
	}
}

// Type returns the task type this subagent handles.
func (r *ReportSubagent) Type() TaskType {
	return TaskTypeReport
}

// Execute generates a formatted report.
func (r *ReportSubagent) Execute(ctx context.Context, task Task) (Result, error) {
	if r.verbose {
		fmt.Println("📝 报告子Agent")
	}
	if r.interactionHandler != nil {
		r.interactionHandler.Log(fmt.Sprintf("> 报告子Agent: %s", task.Description))
	}

	// Get context from parameters if available
	contextData, hasContext := task.Parameters["context"].([]string)

	var prompt string
	if hasContext && len(contextData) > 0 {
		prompt = fmt.Sprintf("基于以下信息，%s:\n\n%s", task.Description, strings.Join(contextData, "\n\n"))
	} else {
		prompt = task.Description
	}

	// Check for global context
	globalContext, _ := task.Parameters["global_context"].(string)
	systemPrompt := "你是一个报告写作助手，负责创建格式良好、清晰且全面的 Markdown 格式报告。使用适当的标题、列表和格式使报告易于阅读。如果提供的信息包含带有 URL 和描述的图片，请选择最相关的图片，并使用标准 Markdown 图片语法 `![描述](URL)` 将其嵌入报告中。将图片放置在相关文本部分附近。"
	if globalContext != "" {
		systemPrompt += "\n\n来自用户的重要上下文/指令：\n" + globalContext
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
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
		fmt.Printf("  ✓ 报告已生成 (%d 字节)\n", len(report))
	}
	if r.interactionHandler != nil {
		r.interactionHandler.Log(fmt.Sprintf("✓ 报告已生成 (%d 字节)", len(report)))
	}

	return Result{
		TaskType: TaskTypeReport,
		Success:  true,
		Output:   report,
	}, nil
}

// RenderSubagent renders markdown to terminal-friendly format.
type RenderSubagent struct {
	verbose            bool
	renderHTML         bool
	interactionHandler InteractionHandler
}

// NewRenderSubagent creates a new RenderSubagent.
func NewRenderSubagent(verbose bool, renderHTML bool, interactionHandler InteractionHandler) *RenderSubagent {
	return &RenderSubagent{
		verbose:            verbose,
		renderHTML:         renderHTML,
		interactionHandler: interactionHandler,
	}
}

// Type returns the task type this subagent handles.
func (r *RenderSubagent) Type() TaskType {
	return TaskTypeRender
}

// Execute renders markdown content.
func (r *RenderSubagent) Execute(ctx context.Context, task Task) (Result, error) {
	if r.verbose {
		fmt.Println("🎨 渲染子Agent")
	}
	if r.interactionHandler != nil {
		r.interactionHandler.Log(fmt.Sprintf("> 渲染子Agent: %s", task.Description))
	}

	// Get content from parameters or description
	content, ok := task.Parameters["content"].(string)
	if !ok {
		// Try to get from context (passed from previous task)
		if ctxContent, ok := task.Parameters["context"].([]string); ok && len(ctxContent) > 0 {
			// Try to find the output from the REPORT task
			var foundReport bool
			for i := len(ctxContent) - 1; i >= 0; i-- {
				if strings.Contains(ctxContent[i], "Output from REPORT task:") {
					content = ctxContent[i]
					// Extract the content after the header
					if idx := strings.Index(content, "\n"); idx != -1 {
						content = content[idx+1:]
					}
					foundReport = true
					break
				}
			}

			if !foundReport {
				// If no REPORT output found, use the last task's output
				content = ctxContent[len(ctxContent)-1]
				// Extract the content after the header if present
				if idx := strings.Index(content, "Output from "); idx != -1 {
					if newlineIdx := strings.Index(content[idx:], "\n"); newlineIdx != -1 {
						content = content[idx+newlineIdx+1:]
					}
				}
			}
			content = strings.TrimSpace(content)
		} else {
			content = task.Description
		}
	}

	if r.verbose {
		fmt.Printf("  正在渲染 %d 字节的内容\n", len(content))
	}
	if r.interactionHandler != nil {
		r.interactionHandler.Log(fmt.Sprintf("正在渲染 %d 字节的内容", len(content)))
	}

	// Render markdown
	var output string
	if r.renderHTML {
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs
		p := parser.NewWithExtensions(extensions)
		doc := p.Parse([]byte(content))

		htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.CompletePage
		opts := html.RendererOptions{Flags: htmlFlags, Title: "Agent Report"}
		renderer := html.NewRenderer(opts)

		output = string(gomarkdown.Render(doc, renderer))
	} else {
		output = string(markdown.Render(content, 80, 6))
	}

	return Result{
		TaskType: TaskTypeRender,
		Success:  true,
		Output:   output,
	}, nil
}
