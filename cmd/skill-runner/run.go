package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/smallnest/goskills"
	"github.com/smallnest/goskills/config"
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
		}

		agent, err := goskills.NewAgent(runnerCfg)
		if err != nil {
			return fmt.Errorf("failed to create agent: %w", err)
		}

		ctx := context.Background()
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