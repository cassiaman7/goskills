# Go Claude Skills Parser

A Go package to parse Claude Skill packages from a directory structure. This parser is designed according to the specifications found in the [official Claude documentation](https://docs.claude.com/en/docs/agents-and-tools/agent-skills/).

## Features

- Parses `SKILL.md` for skill metadata and instructions.
- Extracts YAML frontmatter into a Go struct (`SkillMeta`).
- Captures the Markdown body of the skill.
- Discovers resource files in `scripts/`, `references/`, and `assets/` directories.
- Packaged as a reusable Go module.

## Installation

To use this package in your project, you can use `go get` to add it to your dependencies.

```shell
go get github.com/smallnest/goskills
```

## Usage

Here is an example of how to use the `ParseSkillPackage` function to parse a skill directory.

```go
package main

import (
	"fmt"
	"log"

	"github.com/your-username/goskills" // Replace with the actual import path
)

func main() {
	// Path to the skill directory you want to parse
	skillDirectory := "./examples/skills/artifacts-builder"

	skillPackage, err := goskills.ParseSkillPackage(skillDirectory)
	if err != nil {
		log.Fatalf("Failed to parse skill package: %v", err)
	}

	// Print the parsed information
	fmt.Printf("Successfully Parsed Skill: %s\n", skillPackage.Meta.Name)
	fmt.Println("---------------------------------")
	
	fmt.Printf("Description: %s\n", skillPackage.Meta.Description)
	
	if skillPackage.Meta.Model != "" {
		fmt.Printf("Model: %s\n", skillPackage.Meta.Model)
	}

	if len(skillPackage.Meta.AllowedTools) > 0 {
		fmt.Printf("Allowed Tools: %v\n", skillPackage.Meta.AllowedTools)
	}

	// Print discovered resources
	if len(skillPackage.Resources.Scripts) > 0 {
		fmt.Printf("Scripts: %v\n", skillPackage.Resources.Scripts)
	}
    if len(skillPackage.Resources.References) > 0 {
		fmt.Printf("References: %v\n", skillPackage.Resources.References)
	}
    if len(skillPackage.Resources.Assets) > 0 {
		fmt.Printf("Assets: %v\n", skillPackage.Resources.Assets)
	}

	// Print a short excerpt of the body
	bodyExcerpt := skillPackage.Body
	if len(bodyExcerpt) > 150 {
		bodyExcerpt = bodyExcerpt[:150] + "..."
	}
	fmt.Printf("Body Excerpt: %s\n", bodyExcerpt)
}
```

## Running Tests

To run the tests for this package, navigate to the project root directory and run:

```shell
go test
```
