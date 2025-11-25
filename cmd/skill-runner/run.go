package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/smallnest/goskills"
	"github.com/smallnest/goskills/config"
	goskills_mcp "github.com/smallnest/goskills/mcp"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [prompt]",
	Short: "Processes a user request by selecting and executing a skill.",
	Long: `Processes a user request by simulating the Claude skill-use workflow with an OpenAI-compatible model.
	
This command first discovers all available skills, then asks the LLM to select the most appropriate one.
Finally, it executes the selected skill by feeding its content to the LLM as a system prompt.
If the LLM decides to call a tool, the tool will be executed and its output fed back to the LLM.

Requires the OPENAI_API_KEY environment variable to be set.
You can specify a custom model and API base URL using flags.`,
	Args: cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		userPrompt := strings.Join(args, " ")
		if len(args) == 0 {
			userPromptBytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			userPrompt = strings.TrimSpace(string(userPromptBytes))
		}

		if userPrompt == "" {
			return cmd.Help()
		}

		cfg, err := config.LoadConfig(cmd)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		runnerCfg := goskills.RunnerConfig{
			APIKey:           cfg.APIKey,
			APIBase:          cfg.APIBase,
			Model:            cfg.Model,
			SkillsDir:        cfg.SkillsDir,
			Verbose:          cfg.Verbose,
			AutoApproveTools: cfg.AutoApproveTools,
			AllowedScripts:   cfg.AllowedScripts,
			Loop:             cfg.Loop,
		}

		ctx := context.Background()

		// Initialize MCP Client
		var mcpClient *goskills_mcp.Client
		var mcpConfigPath string

		if cfg.McpConfig != "" {
			mcpConfigPath = cfg.McpConfig
		} else {
			// Check local mcp.json
			if _, err := os.Stat("mcp.json"); err == nil {
				mcpConfigPath = "mcp.json"
			} else {
				// // Check ~/.claude.json
				// homeDir, err := os.UserHomeDir()
				// if err == nil {
				// 	path := fmt.Sprintf("%s/.claude.json", homeDir)
				// 	if _, err := os.Stat(path); err == nil {
				// 		mcpConfigPath = path
				// 	}
				// }
			}
		}

		if mcpConfigPath != "" {
			if cfg.Verbose {
				fmt.Printf("📂 Loading MCP config from: %s\n", mcpConfigPath)
			}
			mcpConfig, err := goskills_mcp.LoadConfig(mcpConfigPath)
			if err != nil {
				fmt.Printf("⚠️ Failed to load MCP config: %v\n", err)
			} else {
				mcpClient, err = goskills_mcp.NewClient(ctx, mcpConfig)
				if err != nil {
					fmt.Printf("⚠️ Failed to create MCP client: %v\n", err)
				} else {
					defer mcpClient.Close()
					if cfg.Verbose {
						fmt.Println("✅ MCP Client initialized.")
					}
				}
			}
		}

		agent, err := goskills.NewAgent(runnerCfg, mcpClient)
		if err != nil {
			return fmt.Errorf("failed to create agent: %w", err)
		}

		if runnerCfg.Loop {
			return agent.RunLoop(ctx, userPrompt)
		}

		result, err := agent.Run(ctx, userPrompt)
		if err != nil {
			return err
		}

		fmt.Println(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	config.SetupFlags(runCmd)
}
