package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// Config holds the application configuration
type Config struct {
	SkillsDir        string
	Model            string
	APIBase          string
	APIKey           string
	AutoApproveTools bool
	AllowedScripts   []string
	Verbose          bool
	Loop             bool
	McpConfig        string
}

// LoadConfig loads configuration from flags and environment variables
func LoadConfig(cmd *cobra.Command) (*Config, error) {
	cfg := &Config{}

	// 1. Load from flags (if set)
	var err error
	cfg.SkillsDir, err = cmd.Flags().GetString("skills-dir")
	if err != nil {
		return nil, err
	}
	cfg.Model, err = cmd.Flags().GetString("model")
	if err != nil {
		return nil, err
	}
	cfg.APIBase, err = cmd.Flags().GetString("api-base")
	if err != nil {
		return nil, err
	}
	cfg.APIKey, err = cmd.Flags().GetString("api-key")
	if err != nil {
		return nil, err
	}
	cfg.AutoApproveTools, err = cmd.Flags().GetBool("auto-approve")
	if err != nil {
		return nil, err
	}
	cfg.Verbose, err = cmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, err
	}
	cfg.Loop, err = cmd.Flags().GetBool("loop")
	if err != nil {
		return nil, err
	}
	cfg.AllowedScripts, err = cmd.Flags().GetStringSlice("allow-scripts")
	if err != nil {
		return nil, err
	}
	cfg.McpConfig, err = cmd.Flags().GetString("mcp-config")
	if err != nil {
		return nil, err
	}

	// 2. Load from environment variables (fallback if flag not set or empty, except bools)
	// Note: Cobra flags usually handle defaults, but we check env vars here for precedence if needed
	// or simply rely on Cobra's binding if we bound them.
	// Here we manually check env vars for critical items if flags are default/empty.

	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("OPENAI_API_KEY")
	}
	if cfg.APIBase == "" {
		cfg.APIBase = os.Getenv("OPENAI_API_BASE")
	}
	if cfg.Model == "" {
		cfg.Model = os.Getenv("OPENAI_MODEL")
	}
	cfg.APIBase = strings.TrimSuffix(cfg.APIBase, "/")

	// Resolve SkillsDir to absolute path
	if cfg.SkillsDir == "" {
		cfg.SkillsDir = "~/.claude/skills" // Default
	}
	absSkillsDir, err := filepath.Abs(cfg.SkillsDir)
	if err != nil {
		return nil, err
	}
	cfg.SkillsDir = absSkillsDir

	return cfg, nil
}

// SetupFlags registers the flags with the command
func SetupFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("skills-dir", "d", "./examples/skills", "Path to the skills directory")
	cmd.Flags().StringP("model", "m", "", "OpenAI-compatible model name (falls back to OPENAI_MODEL env var)")
	cmd.Flags().StringP("api-base", "b", "", "OpenAI-compatible API base URL (falls back to OPENAI_API_BASE env var)")
	cmd.Flags().StringP("api-key", "k", "", "OpenAI-compatible API key (falls back to OPENAI_API_KEY env var)")
	cmd.Flags().Bool("auto-approve", false, "Auto-approve all tool calls (WARNING: potentially unsafe)")
	cmd.Flags().StringSlice("allow-scripts", nil, "Comma-separated list of allowed script names (e.g. 'run_myscript_py')")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolP("loop", "l", false, "Enable interactive loop mode")
	cmd.Flags().String("mcp-config", "", "Path to MCP configuration file")
}
