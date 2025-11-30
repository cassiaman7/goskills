# Go Claude Skills Parser

[English](README.md) | 简体中文

这是一个 Go 语言包，用于从目录结构中解析 Claude Skill 包。该解析器是根据 [官方 Claude 文档](https://docs.claude.com/en/docs/agents-and-tools/agent-skills/) 中的规范设计的。

[![License](https://img.shields.io/:license-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![GoDoc](https://godoc.org/github.com/smallnest/goskills?status.png)](http://godoc.org/github.com/smallnest/goskills)  [![github actions](https://github.com/smallnest/goskills/actions/workflows/go.yml/badge.svg)](https://github.com/smallnest/goskills/actions) [![Go Report Card](https://goreportcard.com/badge/github.com/smallnest/goskills)](https://goreportcard.com/report/github.com/smallnest/goskills) 

[![YouTube Video](https://img.youtube.com/vi/Lod9DnAfd9c/maxresdefault.jpg)](https://www.youtube.com/watch?v=Lod9DnAfd9c)

## 特性

- 解析 `SKILL.md` 获取技能元数据和指令。
- 提取 YAML frontmatter 到 Go 结构体 (`SkillMeta`)。
- 捕获技能的 Markdown 正文。
- 发现 `scripts/`、`references/` 和 `assets/` 目录中的资源文件。
- 打包为可重用的 Go 模块。
- 包含用于管理和检查技能的命令行接口。
- 包含一个深度研究 Agent。

## 安装

要在你的项目中使用此包，可以使用 `go get`：

```shell
go get github.com/smallnest/goskills
```

## 深度研究 Agent

本项目包含一个独立的深度研究 Agent (`agent-web`)，展示了可组合 AI 技能的强大功能。

- **规划器-执行器-子代理架构**：用于解决复杂任务的稳健设计。
- **Web 界面**：现代化的 Web 界面，提供出色的用户体验。
- **无外部框架**：纯 Go 实现，易于理解和扩展。

![Agent Workflow](docs/images/agent_worflow.png)
![Agent Web Interface](docs/images/agent_web.png)

**演示环境**：[https://agent.rpcx.io](https://agent.rpcx.io)

快速开始：

```shell
export OPENAI_API_KEY="YOUR_KEY"
export TAVILY_API_KEY="YOUR_TAVILY_KEY"
make
./agent-web -v
```

更多详情请参阅 [agent.md](agent.md)。

## 命令行接口

本项目提供了两个独立的命令行工具：

### 1. 技能管理 CLI (`goskills-cli`)

位于 `cmd/skill-cli`，此工具帮助你检查和管理本地 Claude 技能。

#### 构建 `goskills-cli`
你可以从项目根目录构建可执行文件：
```shell
go build -o goskills-cli ./cmd/skill-cli
```

#### 命令
以下是 `goskills-cli` 的可用命令：

#### list
列出给定目录中的所有有效技能。
```shell
./goskills-cli list ./examples/skills
```

#### parse
解析单个技能并显示其结构摘要。
```shell
./goskills-cli parse ./examples/skills/artifacts-builder
```

#### detail
显示单个技能的完整详细信息，包括完整的正文内容。
```shell
./goskills-cli detail ./examples/skills/artifacts-builder
```

#### files
列出组成技能包的所有文件。
```shell
./goskills-cli files ./examples/skills/artifacts-builder
```

#### search
在目录中按名称或描述搜索技能。搜索不区分大小写。
```shell
./goskills-cli search ./examples/skills "web app"
```

### 2. 技能运行器 CLI (`goskills-runner`)

位于 `cmd/skill-runner`，此工具通过集成像 OpenAI 模型这样的大型语言模型 (LLM) 来模拟 Claude 技能使用工作流。

#### 构建 `goskills` 运行器
你可以从项目根目录构建可执行文件：
```shell
go build -o goskills ./cmd/skill-runner
```

#### 命令
以下是 `goskills` 的可用命令：

#### run
处理用户请求，首先发现可用技能，然后要求 LLM 选择最合适的技能，最后通过将所选技能的内容作为系统提示提供给 LLM 来执行该技能。

**需要设置 `OPENAI_API_KEY` 环境变量。**

```shell
# 使用默认 OpenAI 模型 (gpt-4o) 的示例
export OPENAI_API_KEY="YOUR_OPENAI_API_KEY"
./goskills run "create an algorithm that generates abstract art"

# 使用自定义 OpenAI 兼容模型和 API 基础 URL（使用环境变量）的示例
export OPENAI_API_KEY="YOUR_OPENAI_API_KEY"
export OPENAI_API_BASE="https://qianfan.baidubce.com/v2"
export OPENAI_MODEL="deepseek-v3"
./goskills run "create an algorithm that generates abstract art"

# 使用自定义 OpenAI 兼容模型和 API 基础 URL（使用命令行标志）的示例
export OPENAI_API_KEY="YOUR_OPENAI_API_KEY"
./goskills run --model deepseek-v3 --api-base https://qianfan.baidubce.com/v2 "create an algorithm that generates abstract art"

# 使用自定义 OpenAI 兼容模型和 API 基础 URL（使用命令行标志），无人工介入自动批准的示例
./goskills run --auto-approve --model deepseek-v3 --api-base https://qianfan.baidubce.com/v2 "使用markitdown 工具解析网页 https://baike.baidu.com/item/%E5%AD%94%E5%AD%90/1584"

# 使用自定义 OpenAI 兼容模型和 API 基础 URL（使用命令行标志），在循环模式下且不自动退出的示例
./goskills run --auto-approve --model deepseek-v3 --api-base https://qianfan.baidubce.com/v2 --skills-dir=./examples/skills "使用markitdown 工具解析网 页 https://baike.baidu.com/item/%E5%AD%94%E5%AD%90/1584" -l
```

## 库的使用

下面是一个如何使用 `goskills` 库中的 `ParseSkillPackage` 函数来解析技能目录的示例。

```go
package main

import (
	"fmt"
	"log"

	"github.com/smallnest/goskills"
)

func main() {
	// 你想要解析的技能目录的路径
	skillDirectory := "./examples/skills/artifacts-builder"

	skillPackage, err := goskills.ParseSkillPackage(skillDirectory)
	if err != nil {
		log.Fatalf("解析技能包失败: %v", err)
	}

	// 打印解析后的信息
	fmt.Printf("成功解析技能: %s\n", skillPackage.Meta.Name)
	// ... 等等
}
```

### ParseSkillPackages

要查找并解析目录及其子目录中的所有有效技能包，可以使用 `ParseSkillPackages` 函数。它递归扫描给定的路径，识别所有包含 `SKILL.md` 文件的目录，并返回成功解析的 `*SkillPackage` 对象切片。

```go
package main

import (
	"fmt"
	"log"

	"github.com/smallnest/goskills"
)

func main() {
	// 包含所有技能的目录
	skillsRootDirectory := "./examples/skills"

	packages, err := goskills.ParseSkillPackages(skillsRootDirectory)
	if err != nil {
		log.Fatalf("解析技能包失败: %v", err)
	}

	fmt.Printf("找到 %d 个技能:\n", len(packages))
	for _, pkg := range packages {
		fmt.Printf("- 路径: %s, 名称: %s\n", pkg.Path, pkg.Meta.Name)
	}
}
```

## 安装

```
brew install smallnest/goskills/goskills
```

或者

```
# 添加 tap
brew tap smallnest/goskills

# 安装 goskills
brew install goskills

```
