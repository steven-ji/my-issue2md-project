# API Sketch — internal/github & internal/converter

本文档描述 `internal/github` 和 `internal/converter` 包对外暴露的主要接口签名，作为后续开发的参考。

---

## internal/github

负责与 GitHub API 交互，获取 Issue / PR / Discussion 的原始数据。

### 核心类型

```go
// Resource 统一表示 GitHub 资源的元信息与内容
type Resource struct {
    Type     ResourceType
    Title    string
    Author   string
    URL      string
    State    string
    Labels   []string
    CreatedAt time.Time
    Body     string
    Comments []Comment
    Answer   *int // Discussion 中被标记为 Answer 的评论索引（0-based），nil 表示无
}

type ResourceType string

const (
    TypeIssue       ResourceType = "issue"
    TypePullRequest ResourceType = "pull_request"
    TypeDiscussion  ResourceType = "discussion"
)

type Comment struct {
    ID        int
    Author    string
    CreatedAt time.Time
    Body      string
    IsReview  bool   // 是否为 PR Review Comment
    Reactions map[string]int // emoji -> count，仅当启用时填充
}
```

### 核心接口

```go
// Client 是 GitHub API 的抽象，便于测试时替换为 mock
type Client interface {
    // FetchResource 根据 URL 解析出的类型和标识，获取完整资源数据
    FetchResource(ctx context.Context, owner, repo string, resType ResourceType, number int) (*Resource, error)
}

// NewClient 创建基于 GitHub REST API 的 Client 实例
// token 可为空字符串，表示匿名访问（仅公有仓库）
func NewClient(token string) Client
```

### 设计说明

- `Client` 接口将 API 交互与业务逻辑解耦，测试中可注入 stub 实现。
- `FetchResource` 内部处理分页，调用方无需关心翻页逻辑。
- Discussion 使用 GraphQL API 获取（REST API 不支持 Discussion），由 `NewClient` 内部封装，对外统一接口。

---

## internal/converter

负责将 `github.Resource` 转换为 Markdown 文本。

### 核心类型

```go
// Options 控制 Markdown 输出的可选内容
type Options struct {
    EnableReactions  bool
    EnableUserLinks  bool
}
```

### 核心函数

```go
// Convert 将 Resource 转换为完整的 Markdown 文本（含 YAML Frontmatter）
func Convert(res *github.Resource, opts Options) (string, error)
```

### 设计说明

- `Convert` 是纯函数，无副作用，不依赖 I/O，易于单元测试。
- Frontmatter 中的字段名与 `github.Resource` 字段对应，由 `Convert` 内部格式化。
- Reactions 和用户链接的渲染由 `Options` 控制，`Convert` 内部根据选项决定是否输出对应区块。
