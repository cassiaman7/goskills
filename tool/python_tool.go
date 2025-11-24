package tool

import (
	"bytes"
	"fmt"

	"os"
	"os/exec"
	"text/template"
)

type PythonTool struct {
}

func (t *PythonTool) Run(args map[string]any, code string) (string, error) {
	tmpl, err := template.New("python").Parse(code)
	if err != nil {
		return "", fmt.Errorf("failed to parse python template: %w", err)
	}

	var script bytes.Buffer
	err = tmpl.Execute(&script, args)
	if err != nil {
		return "", fmt.Errorf("failed to execute python template: %w", err)
	}

	tmpfile, err := os.CreateTemp("", "python-*.py")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(script.Bytes()); err != nil {
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}
	if err := tmpfile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	return RunPythonScript(tmpfile.Name(), nil)
}

// RunPythonScript executes a Python script and returns its combined stdout and stderr.
// It tries to use 'python3' first, then falls back to 'python'.
func RunPythonScript(scriptPath string, args []string) (string, error) {
	pythonExe, err := exec.LookPath("python3")
	if err != nil {
		pythonExe, err = exec.LookPath("python")
		if err != nil {
			return "", fmt.Errorf("failed to find python3 or python in PATH: %w", err)
		}
	}

	cmd := exec.Command(pythonExe, append([]string{scriptPath}, args...)...)
	cmd.Env = os.Environ()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run python script '%s' with '%s': %w\nStdout: %s\nStderr: %s", scriptPath, pythonExe, err, stdout.String(), stderr.String())
	}

	return stdout.String() + stderr.String(), nil
}
