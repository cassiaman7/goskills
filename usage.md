Before running any `goskills` command, you must set your OpenAI API key:

```bash
export OPENAI_API_KEY="YOUR_OPENAI_API_KEY"
```

---

# goskills Usage Examples

This document provides examples of how to execute the `goskills` command-line tool from various programming languages.

The base command for these examples is:

```shell
goskills run --auto-approve --model deepseek-v3 --api-base https://qianfan.baidubce.com/v2 --skills-dir=~/.claude/skills "使用markitdown 工具解析网页 https://baike.baidu.com/item/%E5%AD%94%E5%AD%90/1584"
```

---

## Shell (Bash)

This is the most direct way to run the command.

```bash
#!/bin/bash

# Define the prompt
PROMPT="使用markitdown 工具解析网页 https://baike.baidu.com/item/%E5%AD%94%E5%AD%90/1584"

# Execute the command and capture the output
RESULT=$(goskills run --auto-approve --model deepseek-v3 --api-base https://qianfan.baidubce.com/v2 --skills-dir=~/.claude/skills "$PROMPT")

# Or execute and print directly
# goskills run --auto-approve --model deepseek-v3 --api-base https://qianfan.baidubce.com/v2 "$PROMPT" --skills-dir=~/.claude/skills


echo "Output:"
echo "$RESULT"
```

---

## Python

Using the `subprocess` module is the standard way to run external commands in Python.

```python
import subprocess
import shlex

# Define the command as a list of arguments for safety
command = [
    "goskills", "run",
    "--auto-approve",
    "--model", "deepseek-v3",
    "--api-base", "https://qianfan.baidubce.com/v2",
    "--skills-dir=~/.claude/skills",
    "使用markitdown 工具解析网页 https://baike.baidu.com/item/%E5%AD%94%E5%AD%90/1584"
]

# Or, build the command from a string using shlex for proper quoting
# cmd_str = 'goskills run --auto-approve --model deepseek-v3 --api-base https://qianfan.baidubce.com/v2 "使用markitdown 工具解析网页 https://baike.baidu.com/item/%E5%AD%94%E5%AD%90/1584" --skills-dir=~/.claude/skills'
# command = shlex.split(cmd_str)


try:
    # Execute the command, capture stdout and stderr
    result = subprocess.run(
        command,
        check=True,        # Raise an exception if the command returns a non-zero exit code
        capture_output=True, # Capture stdout and stderr
        text=True          # Decode stdout/stderr as text
    )
    
    print("Command executed successfully:")
    print("Output:\n", result.stdout)

except FileNotFoundError:
    print("Error: 'goskills' command not found. Make sure it's in your PATH.")
except subprocess.CalledProcessError as e:
    print(f"Command failed with exit code {e.returncode}:")
    print("Stderr:\n", e.stderr)

```

---

## JavaScript (Node.js)

In Node.js, you can use the `child_process` module.

```javascript
const { exec } = require('child_process');

// Use single quotes for the outer string to easily handle inner double quotes
const command = 'goskills run --auto-approve --model deepseek-v3 --api-base https://qianfan.baidubce.com/v2 --skills-dir=~/.claude/skills "使用markitdown 工具解析网页 https://baike.baidu.com/item/%E5%AD%94%E5%AD%90/1584"';

exec(command, (error, stdout, stderr) => {
    if (error) {
        console.error(`Execution error: ${error.message}`);
        if (stderr) {
            console.error(`Stderr: ${stderr}`);
        }
        return;
    }

    console.log(`Command output:\n${stdout}`);
});
```

---

## Go

In Go, the `os/exec` package is used to run external commands.

```go
package main

import (
	"fmt"
	"os/exec"
)

func main() {
	prompt := "使用markitdown 工具解析网页 https://baike.baidu.com/item/%E5%AD%94%E5%AD%90/1584"
	
	cmd := exec.Command("goskills", "run",
		"--auto-approve",
		"--model", "deepseek-v3",
		"--api-base", "https://qianfan.baidubce.com/v2",
		"--skills-dir=~/.claude/skills",
		prompt)

	// CombinedOutput runs the command and returns its combined standard output and standard error.
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		fmt.Printf("Output:\n%s\n", string(output))
		return
	}

	fmt.Printf("Command output:\n%s\n", string(output))
}
```

---

## Java

Using `ProcessBuilder` is the modern and recommended way to execute commands in Java.

```java
import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.io.IOException;

public class GoSkillsRunner {
    public static void main(String[] args) {
        try {
            ProcessBuilder pb = new ProcessBuilder(
                "goskills", "run",
                "--auto-approve",
                "--model", "deepseek-v3",
                "--api-base", "https://qianfan.baidubce.com/v2",
                "--skills-dir=~/.claude/skills",
                "使用markitdown 工具解析网页 https://baike.baidu.com/item/%E5%AD%94%E5%AD%90/1584"
            );

            // Redirect error stream to the same as the standard output stream
            pb.redirectErrorStream(true);

            Process process = pb.start();

            // Read the output from the command
            StringBuilder output = new StringBuilder();
            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    output.append(line).append("\n");
                }
            }

            int exitCode = process.waitFor();
            System.out.println("Exit Code: " + exitCode);
            System.out.println("Output:\n" + output.toString());

        } catch (IOException | InterruptedException e) {
            e.printStackTrace();
        }
    }
}

```

---

## Rust

In Rust, the `std::process::Command` struct is used to execute external commands.

```rust
use std::process::Command;

fn main() {
    let prompt = "使用markitdown 工具解析网页 https://baike.baidu.com/item/%E5%AD%94%E5%AD%90/1584";

    let output = Command::new("goskills")
        .arg("run")
        .arg("--auto-approve")
        .arg("--model")
        .arg("deepseek-v3")
        .arg("--api-base")
        .arg("https://qianfan.baidubce.com/v2")
        .arg("--skills-dir=~/.claude/skills")
        .arg(prompt)
        .output()
        .expect("Failed to execute command");

    if output.status.success() {
        println!("Command executed successfully:");
        println!("Output:\n{}", String::from_utf8_lossy(&output.stdout));
    } else {
        eprintln!("Command failed with exit code: {:?}", output.status.code());
        eprintln!("Stderr:\n{}", String::from_utf8_lossy(&output.stderr));
    }
}
```

---

## C++

In C++, `std::system` provides a simple way to execute shell commands, though `popen` or `fork`/`exec` offer more control for capturing output. For this example, `std::system` is used for simplicity.

```cpp
#include <iostream>
#include <string>
#include <cstdlib> // For std::system

int main() {
    std::string prompt = "使用markitdown 工具解析网页 https://baike.baidu.com/item/%E5%AD%94%E5%AD%90/1584";
    std::string command = "goskills run --auto-approve --model deepseek-v3 --api-base https://qianfan.baidubce.com/v2 --skills-dir=~/.claude/skills \"" + prompt + "\"";

    // system() executes the command and returns its exit status
    // For capturing output, popen or platform-specific APIs would be needed.
    int result = std::system(command.c_str());

    if (result == 0) {
        std::cout << "Command executed successfully." << std::endl;
    } else {
        std::cerr << "Command failed with exit code: " << result << std::endl;
    }

    return 0;
}
```

---

## C

In C, the `system()` function is part of `stdlib.h` and allows executing shell commands. Like C++, it's simple but doesn't easily capture output.

```c
#include <stdio.h>
#include <stdlib.h> // For system()
#include <string.h> // For strlen, strcat

int main() {
    char prompt[] = "使用markitdown 工具解析网页 https://baike.baidu.com/item/%E5%AD%94%E5%AD%90/1584";
    // Allocate enough space for the command string
    // "goskills run --auto-approve --model deepseek-v3 --api-base https://qianfan.baidubce.com/v2 "" + prompt + """ --skills-dir=~/.claude/skills"
    char command[512]; // Adjust size as necessary

    snprintf(command, sizeof(command), "goskills run --auto-approve --model deepseek-v3 --api-base https://qianfan.baidubce.com/v2 --skills-dir=~/.claude/skills \"%s\"", prompt);

    // system() executes the command and returns its exit status
    // For capturing output, popen or platform-specific APIs would be needed.
    int result = system(command);

    if (result == 0) {
        printf("Command executed successfully.\n");
    } else {
        fprintf(stderr, "Command failed with exit code: %d\n", result);
    }

    return 0;
}

```