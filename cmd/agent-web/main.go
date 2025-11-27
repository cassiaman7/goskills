package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/smallnest/goskills/agent"
	"github.com/spf13/cobra"
)

//go:embed ui/*
var uiAssets embed.FS

var (
	apiKey  string
	apiBase string
	model   string
	addr    string
	verbose bool
)

// WebInteractionHandler implements agent.InteractionHandler for the web interface.
type WebInteractionHandler struct {
	eventChan    chan Event
	responseChan chan string
}

type Event struct {
	Type    string      `json:"type"`
	Content string      `json:"content,omitempty"`
	Plan    *agent.Plan `json:"plan,omitempty"`
}

func NewWebInteractionHandler() *WebInteractionHandler {
	return &WebInteractionHandler{
		eventChan:    make(chan Event, 100),
		responseChan: make(chan string),
	}
}

func (h *WebInteractionHandler) ReviewPlan(plan *agent.Plan) (string, error) {
	h.eventChan <- Event{
		Type: "plan_review",
		Plan: plan,
	}
	// Wait for user response
	response := <-h.responseChan
	return response, nil
}

func (h *WebInteractionHandler) ReviewSearchResults(results string) (bool, error) {
	// Filter out "Content: ..." lines to keep the display clean
	var filteredLines []string
	lines := strings.Split(results, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "Content:") {
			filteredLines = append(filteredLines, line)
		}
	}
	cleanResults := strings.Join(filteredLines, "\n")

	h.Log(fmt.Sprintf("🔎 Search Results:\n%s", cleanResults))
	return false, nil
}

func (h *WebInteractionHandler) ConfirmPodcastGeneration(report string) (bool, error) {
	// Auto-approve or skip for now
	return false, nil
}

func (h *WebInteractionHandler) Log(message string) {
	h.eventChan <- Event{
		Type:    "log",
		Content: message,
	}
}

func (h *WebInteractionHandler) Broadcast(event Event) {
	h.eventChan <- event
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "agent-web",
		Short: "Start the agent web interface",
		Run:   runServer,
	}

	rootCmd.Flags().StringVar(&apiKey, "api-key", os.Getenv("OPENAI_API_KEY"), "OpenAI API Key")
	rootCmd.Flags().StringVar(&apiBase, "api-base", os.Getenv("OPENAI_API_BASE"), "OpenAI API Base URL")
	rootCmd.Flags().StringVar(&model, "model", os.Getenv("OPENAI_MODEL"), "OpenAI Model")
	rootCmd.Flags().StringVar(&addr, "addr", "127.0.0.1:8080", "Address to listen on")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	if apiKey == "" {
		log.Fatal("API key is required")
	}

	handler := NewWebInteractionHandler()

	// Initialize agent config
	config := agent.AgentConfig{
		APIKey:     apiKey,
		APIBase:    apiBase,
		Model:      model,
		Verbose:    verbose,
		RenderHTML: true,
	}

	// Serve static files
	// Use fs.Sub to get the "ui" subdirectory from the embedded filesystem
	uiFS, err := fs.Sub(uiAssets, "ui")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", http.FileServer(http.FS(uiFS)))

	// API endpoints
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		for {
			select {
			case event := <-handler.eventChan:
				data, err := json.Marshal(event)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	})

	// Create a single agent instance for the session
	// Note: In a real multi-user app, this should be managed per session/user
	planningAgent, err := agent.NewPlanningAgent(config, handler)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Run agent in a goroutine
		go func() {
			defer func() {
				if r := recover(); r != nil {
					handler.Broadcast(Event{
						Type:    "error",
						Content: fmt.Sprintf("Panic: %v", r),
					})
				}
			}()

			// Check for direct chat
			if strings.HasPrefix(req.Message, "\\") {
				msg := strings.TrimPrefix(req.Message, "\\")

				planningAgent.AddDeveloperMessage(msg)

				// Log user request
				handler.Broadcast(Event{
					Type:    "log",
					Content: fmt.Sprintf("> User Request: %s", msg),
				})

				handler.Broadcast(Event{
					Type: "done",
				})
				return
			}

			// Add user message to history
			planningAgent.AddUserMessage(req.Message)

			// Plan with review
			plan, err := planningAgent.PlanWithReview(context.Background(), req.Message)
			if err != nil {
				handler.Broadcast(Event{
					Type:    "error",
					Content: err.Error(),
				})
				return
			}

			// Execute
			results, err := planningAgent.Execute(context.Background(), plan)
			if err != nil {
				handler.Broadcast(Event{
					Type:    "error",
					Content: err.Error(),
				})
				return
			}

			// Extract final output
			var finalOutput string
			for i := len(results) - 1; i >= 0; i-- {
				if (results[i].TaskType == agent.TaskTypeRender || results[i].TaskType == agent.TaskTypeReport) && results[i].Success {
					finalOutput = results[i].Output
					break
				}
			}
			if finalOutput == "" {
				for _, result := range results {
					if result.Success {
						finalOutput += result.Output + "\n\n"
					}
				}
			}

			// Add assistant message
			planningAgent.AddAssistantMessage(finalOutput)

			handler.Broadcast(Event{
				Type:    "response",
				Content: finalOutput,
			})

			handler.Broadcast(Event{
				Type: "done",
			})
		}()

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/api/respond", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Response string `json:"response"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Send response to the waiting channel
		// Non-blocking send to avoid hanging if no one is listening
		select {
		case handler.responseChan <- req.Response:
		default:
			// No one waiting, maybe log warning
		}

		w.WriteHeader(http.StatusOK)
	})

	fmt.Printf("Starting server on http://%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
