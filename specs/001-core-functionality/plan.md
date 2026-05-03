# issue2md 技术实现方案

**Version:** 1.0
**Date:** 2026-05-03
**Scope:** spec.md 001-core-functionality

---

## 1. 技术上下文

| 维度 | 选型 | 理由 |
|---|---|---|
| 语言 | Go >= 1.24 | spec 要求 >= 1.21，项目 go.mod 需从 1.18 升级 |
| CLI 参数解析 | `flag` 标准库 | 宪法 1.2：标准库优先。`flag` 足以满足当前参数需求 |
| GitHub REST API | `google/go-github` (v68+) | 成熟稳定，类型安全，社区标准库 |
| GitHub GraphQL API | `shurcooL/graphql` | Discussion 仅 GraphQL 支持，该库轻量、零反射 |
| Markdown 生成 | 标准库 `fmt` + `text/template` | 宪法 1.2：不引入第三方 Markdown 库 |
| HTTP 客户端 | `net/http` | 宪法 1.2：标准库优先 |
| 数据存储 | 无 | spec 明确：初期不需要数据库，实时 API 获取 |

**唯一外部依赖：** `google/go-github` + `shurcooL/graphql`（均因 GitHub API 交互必要性引入）。

---

## 2. "合宪性"审查

### 第一条：简单性原则

| 条款 | 合规性 | 说明 |
|---|---|---|
| 1.1 YAGNI | ✅ | 仅实现 spec 中明确要求的功能，Web 入口本次不实现 |
| 1.2 标准库优先 | ✅ | Web 用 `net/http`，CLI 用 `flag`，Markdown 用 `text/template` |
| 1.3 反过度工程 | ✅ | `converter.Convert` 为纯函数，`github.Client` 接口仅一个方法，不引入不必要的抽象层 |

### 第二条：测试先行铁律

| 条款 | 合规性 | 说明 |
|---|---|---|
| 2.1 TDD 循环 | ✅ | 每个功能先写失败测试，再实现。开发顺序见第 7 节 |
| 2.2 表格驱动 | ✅ | `parser`、`converter`、`config` 包均使用 `[]struct{ name string; ... }` 风格 |
| 2.3 拒绝 Mocks | ✅ | `github.Client` 接口用于注入 stub（硬编码数据），非 mock 框架。集成测试使用真实 API 调用（需 `GITHUB_TOKEN`） |

### 第三条：明确性原则

| 条款 | 合规性 | 说明 |
|---|---|---|
| 3.1 错误处理 | ✅ | 所有错误 `fmt.Errorf("...: %w", err)` 包装，无一遗漏 |
| 3.2 无全局变量 | ✅ | 配置通过 `config.Config` 结构体传递，`github.Client` 通过依赖注入，无包级可变状态 |

---

## 3. 项目结构与依赖关系

```
issue2md/
├── cmd/
│   └── issue2md/           # CLI 入口：main()，仅做组装和调用
│       └── main.go
├── internal/
│   ├── config/             # 配置管理：读取环境变量 + 解析 CLI flags
│   │   ├── config.go       #   Config struct + Load() 函数
│   │   └── config_test.go
│   ├── parser/             # URL 解析：识别资源类型，提取 owner/repo/number
│   │   ├── parser.go       #   Parse() 函数
│   │   └── parser_test.go
│   ├── github/             # GitHub API 交互：获取 Resource 数据
│   │   ├── client.go       #   Client 接口 + APIClient 实现
│   │   ├── issue.go        #   Issue 获取逻辑（REST + 分页）
│   │   ├── pull.go         #   PR 获取逻辑（REST + 分页，合并两种 Comments）
│   │   ├── discussion.go   #   Discussion 获取逻辑（GraphQL）
│   │   ├── types.go        #   Resource, Comment, ResourceType 等类型定义
│   │   └── client_test.go
│   └── converter/          # Markdown 渲染：Resource → Markdown string
│       ├── converter.go    #   Convert() 函数
│       ├── frontmatter.go  #   YAML Frontmatter 生成
│       ├── converter_test.go
│       └── testdata/       #   黄金文件测试数据
├── web/
│   ├── templates/          # [Future] HTML 模板
│   └── static/             # [Future] 静态资源
├── specs/                  # 规格文档
├── Makefile
├── go.mod
└── go.sum
```

### 包依赖方向（单向，无循环）

```
cmd/issue2md
    ├── internal/cli       (组装：解析参数 → 调用 parser → 调用 github → 调用 converter → 输出)
    │       ├── internal/config
    │       ├── internal/parser
    │       ├── internal/github
    │       └── internal/converter
    ├── internal/config
    ├── internal/parser
    ├── internal/github
    └── internal/converter
```

**依赖规则：**
- `cmd/` 可依赖所有 `internal/` 包。
- `internal/converter` 可依赖 `internal/github` 的类型定义（`Resource`, `Comment`），但不依赖其接口或实现。
- `internal/parser` 不依赖任何其他 `internal/` 包。
- `internal/github` 不依赖 `internal/converter` 或 `internal/parser`。
- `internal/config` 不依赖任何其他 `internal/` 包。

---

## 4. 核心数据结构

### internal/parser

```go
// ParsedURL 是 URL 解析的结果
type ParsedURL struct {
    Owner  string
    Repo   string
    Type   ResourceType // 复用 github.ResourceType
    Number int
}
```

### internal/github

```go
type ResourceType string

const (
    TypeIssue       ResourceType = "issue"
    TypePullRequest ResourceType = "pull_request"
    TypeDiscussion  ResourceType = "discussion"
)

// Resource 统一表示 GitHub 资源
type Resource struct {
    Type      ResourceType
    Title     string
    Author    string
    URL       string
    State     string            // "open", "closed", "merged"
    Labels    []string
    CreatedAt time.Time
    Body      string
    Comments  []Comment
    AnswerID  *int64            // Discussion 中被标记为 Answer 的评论 ID，nil 表示无
}

// Comment 统一表示所有类型的评论
type Comment struct {
    ID        int64
    Author    string
    CreatedAt time.Time
    Body      string
    IsReview  bool             // 是否为 PR Review Comment
    Reactions map[string]int   // emoji → count，仅当启用 reactions 时填充
}
```

### internal/config

```go
// Config 是应用运行配置
type Config struct {
    URL              string
    OutputFile       string         // 位置参数或 -o 指定的输出路径，空字符串表示 stdout
    EnableReactions  bool
    EnableUserLinks  bool
    GitHubToken      string         // 从环境变量 GITHUB_TOKEN 读取
}
```

### internal/converter

```go
// Options 控制 Markdown 输出
type Options struct {
    EnableReactions bool
    EnableUserLinks bool
}
```

---

## 5. 接口设计

### internal/github — Client 接口

```go
// Client 是 GitHub API 的抽象接口
type Client interface {
    // FetchResource 获取完整的 GitHub 资源数据
    // 内部处理分页，调用方无需关心翻页。
    // enableReactions 控制是否填充 Comment.Reactions 字段。
    FetchResource(ctx context.Context, owner, repo string, resType ResourceType, number int, enableReactions bool) (*Resource, error)
}

// NewClient 创建 APIClient 实例
// token 为空时匿名访问（仅公有仓库）
func NewClient(token string) Client
```

**实现策略：**
- `APIClient` 结构体持有 `*github.Client`（go-github REST）和 `*http.Client`（用于 GraphQL）。
- `FetchResource` 根据 `resType` 分派到 `fetchIssue`、`fetchPullRequest`、`fetchDiscussion` 私有方法。
- 分页逻辑封装在各私有方法内部。

### internal/parser — Parse 函数

```go
// Parse 解析 GitHub URL，返回结构化的 ParsedURL
// 不支持的 URL 格式返回错误
func Parse(rawURL string) (*ParsedURL, error)
```

### internal/converter — Convert 函数

```go
// Convert 将 Resource 转换为完整 Markdown 文本（含 YAML Frontmatter）
// 纯函数，无副作用，不依赖 I/O
func Convert(res *github.Resource, opts Options) (string, error)
```

### internal/config — Load 函数

```go
// Load 从命令行参数和环境变量构建 Config
func Load(args []string) (*Config, error)
```

---

## 6. 数据流

```
用户输入 URL
     │
     ▼
┌──────────┐
│  config   │ ── Load(args) ──→ Config{URL, OutputFile, EnableReactions, EnableUserLinks, GitHubToken}
└──────────┘
     │
     ▼
┌──────────┐
│  parser   │ ── Parse(url) ──→ ParsedURL{Owner, Repo, Type, Number}
└──────────┘
     │
     ▼
┌──────────┐
│  github   │ ── Client.FetchResource(ctx, owner, repo, type, number, enableReactions) ──→ *Resource
└──────────┘
     │
     ▼
┌───────────┐
│ converter  │ ── Convert(resource, Options{EnableReactions, EnableUserLinks}) ──→ markdown string
└───────────┘
     │
     ▼
  stdout 或写入文件
```

---

## 7. 开发顺序（TDD）

每个步骤遵循 Red → Green → Refactor 循环。

| 阶段 | 包 | 测试先行 | 实现 | 验收 |
|---|---|---|---|---|
| 1 | `internal/parser` | 表格驱动：合法/非法 URL 识别 | `Parse()` | AC-1 全部通过 |
| 2 | `internal/converter` | 表格驱动 + 黄金文件：Issue/PR/Discussion 各一个 | `Convert()` + Frontmatter | AC-2, AC-3, AC-4, AC-6 通过 |
| 3 | `internal/github` | 集成测试（需 `GITHUB_TOKEN`）| `Client` + `fetchIssue/fetchPullRequest/fetchDiscussion` | 数据获取正确 |
| 4 | `internal/config` | 表格驱动：各种参数组合 | `Load()` | Flag 和参数解析正确 |
| 5 | `cmd/issue2md` | 端到端测试 | `main()` 组装全部依赖 | AC-5, AC-7 通过 |

**阶段依赖：** 1 → 2 → 3 → (4 与 3 可并行) → 5

---

## 8. 关键实现细节

### 8.1 PR Comments 合并策略

PR 存在两种 Comments：
- **Issue Comments**（PR 级别）：`GET /repos/{owner}/{repo}/issues/{number}/comments`
- **Review Comments**（代码行级别）：`GET /repos/{owner}/{repo}/pulls/{number}/comments`

合并策略：
1. 分别获取两种 Comments，Review Comments 的 `IsReview = true`
2. 合并为单一 `[]Comment` 切片
3. 按 `CreatedAt` 正序排序

### 8.2 Discussion GraphQL 查询

Discussion 无 REST API，必须使用 GraphQL。查询结构：

```graphql
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    discussion(number: $number) {
      title
      author { login }
      createdAt
      body
      answer { id }
      comments(first: 100, after: $cursor) {
        pageInfo { hasNextPage endCursor }
        nodes {
          id
          author { login }
          createdAt
          body
          isAnswer
        }
      }
    }
  }
}
```

分页通过 `cursor` 实现，循环直到 `hasNextPage == false`。

### 8.3 YAML Frontmatter 生成

使用 `text/template` 生成，模板：

```yaml
---
title: {{ .Title | yamlQuote }}
url: {{ .URL }}
author: {{ .Author }}
created_at: {{ .CreatedAt.Format "2006-01-02T15:04:05Z" }}
state: {{ .State }}
labels:
{{- range .Labels }}
  - {{ . }}
{{- end }}
type: {{ .Type }}
---
```

`yamlQuote` 为自定义模板函数，处理标题中的特殊字符（冒号、引号等）。

### 8.4 Answer 标记

Discussion 中被标记为 Answer 的评论，在 Markdown 中渲染为：

```markdown
> **✅ Answer**
```

位于评论正文之后。

### 8.5 Reactions 渲染

当 `-enable-reactions` 启用时，在评论正文下方以 HTML 注释输出：

```markdown
<!-- Reactions: 👍 5  🎉 2  ❤️ 1 -->
```

使用 HTML 注释而非可见文本，避免干扰 Markdown 渲染，同时保持信息可检索。

### 8.6 Exit Code 规范

| Exit Code | 含义 |
|---|---|
| 0 | 成功 |
| 1 | 一般错误（URL 无效、参数错误） |
| 2 | GitHub API 错误（不存在、权限不足、限流） |

---

## 9. 不在范围内

以下内容明确 **不在** 本次实现范围：

- Web 界面（US-7）
- 批量转换（如某 Label 下所有 Issue）
- Gist 支持
- 图片/附件下载
- 自定义 Markdown 模板
- API 重试机制
- 缓存层
