package tool

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

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
