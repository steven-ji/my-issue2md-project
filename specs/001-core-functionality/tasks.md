# issue2md 任务列表

**Version:** 1.0
**Date:** 2026-05-03
**Scope:** spec.md 001-core-functionality / plan.md

---

## 约定

- **[P]** = 可与同阶段内其他 [P] 任务并行执行
- **依赖** = 必须等待所列任务完成后才能开始
- **TDD** = 测试任务始终位于对应实现任务之前
- 每个任务只涉及一个主要文件的创建或修改

---

## Phase 1: Foundation（数据结构定义）

> 目标：建立编译可通行的类型基础，不包含业务逻辑。

| ID | 任务 | 产出文件 | 依赖 |
|---|---|---|---|
| T1 | 添加外部依赖到 go.mod（`google/go-github` v68+, `shurcooL/graphql`） | `go.mod`, `go.sum` | — |
| T2 [P] | 定义核心类型：`ResourceType` 常量、`Resource` struct、`Comment` struct | `internal/github/types.go` | T1 |
| T3 [P] | 定义 `Config` struct + `Load` 函数签名（返回 error stub） | `internal/config/config.go` | T1 |
| T4 | 定义 `Client` 接口 + `APIClient` struct 声明 + `NewClient` 函数签名（返回 error stub） | `internal/github/client.go` | T2 |
| T5 | 定义 `ParsedURL` struct + `Parse` 函数签名（返回 error stub） | `internal/parser/parser.go` | T2 |
| T6 | 定义 `Options` struct + `Convert` 函数签名（返回 error stub） | `internal/converter/converter.go` | T2 |
| T7 [P] | 创建 Makefile（`make test`, `make build`, `make web` target） | `Makefile` | T1 |

**Phase 1 依赖图：**

```
T1 → T2 ─→ T4
          ─→ T5
          ─→ T6
T1 → T3 [P]
T1 → T7 [P]
```

---

## Phase 2: GitHub Fetcher（API 交互逻辑，TDD）

> 目标：实现 URL 解析 + GitHub API 数据获取。严格遵循 Red → Green → Refactor。

| ID | 任务 | 产出文件 | 依赖 |
|---|---|---|---|
| T8 | 编写 `parser.Parse` 表格驱动测试：合法 Issue/PR/Discussion URL、非法 URL、非 GitHub 域名 | `internal/parser/parser_test.go` | T5 |
| T9 | 实现 `parser.Parse`：URL 解析逻辑，使 T8 全部通过 | `internal/parser/parser.go` | T8 |
| T10 | 编写 `github.fetchIssue` 集成测试（需 `GITHUB_TOKEN`，无 token 时 skip）：验证 Issue 数据获取 + 分页 | `internal/github/issue_test.go` | T4 |
| T11 | 实现 `github.fetchIssue`：REST API 调用 + 自动分页 | `internal/github/issue.go` | T10 |
| T12 | 编写 `github.fetchPullRequest` 集成测试：验证 PR 两种 Comments 合并 + 时间排序 | `internal/github/pull_test.go` | T11 |
| T13 | 实现 `github.fetchPullRequest`：获取 Issue Comments + Review Comments → 合并 → 按 CreatedAt 排序 | `internal/github/pull.go` | T12 |
| T14 | 编写 `github.fetchDiscussion` 集成测试：验证 GraphQL 查询 + Answer 标记 + 分页 | `internal/github/discussion_test.go` | T11 |
| T15 | 实现 `github.fetchDiscussion`：GraphQL 查询 + cursor 分页 + AnswerID 提取 | `internal/github/discussion.go` | T14 |
| T16 | 实现 `github.NewClient` + `APIClient.FetchResource`：组装 REST/GraphQL 客户端，按 resType 分派到 fetch* 方法 | `internal/github/client.go` | T11, T13, T15 |
| T17 | 编写 `github.Client` 端到端集成测试：通过 `FetchResource` 接口验证 Issue/PR/Discussion 完整流程 | `internal/github/client_test.go` | T16 |

**Phase 2 依赖图：**

```
T5 ─→ T8 ─→ T9
T4 ─→ T10 ─→ T11 ─┬→ T12 ─→ T13 ──┐
                   └→ T14 ─→ T15 ──┤
                                    └→ T16 ─→ T17
```

> T8/T9（parser）与 T10/T11（github issue）可并行执行，因为两个包无代码依赖。

---

## Phase 3: Markdown Converter（转换逻辑，TDD）

> 目标：实现 Resource → Markdown 文本转换。纯函数，零 I/O 依赖。

| ID | 任务 | 产出文件 | 依赖 |
|---|---|---|---|
| T18 [P] | 创建 Issue 黄金文件测试数据 | `internal/converter/testdata/issue.md` | T6 |
| T19 [P] | 创建 PR 黄金文件测试数据 | `internal/converter/testdata/pr.md` | T6 |
| T20 [P] | 创建 Discussion 黄金文件测试数据（含 Answer 标记） | `internal/converter/testdata/discussion.md` | T6 |
| T21 | 编写 `converter.frontmatter` 表格驱动测试：YAML Frontmatter 生成（含特殊字符标题、空 Labels 等） | `internal/converter/frontmatter_test.go` | T6 |
| T22 | 实现 `converter.frontmatter`：`text/template` + `yamlQuote` 自定义函数 | `internal/converter/frontmatter.go` | T21 |
| T23 | 编写 `converter.Convert` 表格驱动测试：Issue/PR/Discussion 完整转换 + Reactions 开关 + UserLinks 开关 + Answer 标记 + 黄金文件比对 | `internal/converter/converter_test.go` | T18, T19, T20, T22 |
| T24 | 实现 `converter.Convert`：组装 Frontmatter + 标题行 + 元信息行 + 正文 + Comments 渲染，使 T23 全部通过 | `internal/converter/converter.go` | T23 |

**Phase 3 依赖图：**

```
T6 ─→ T18 [P]
    ─→ T19 [P]
    ─→ T20 [P]
    ─→ T21 ─→ T22 ─┐
                     └→ T23 ─→ T24
T18/19/20 ──────────┘
```

---

## Phase 4: CLI Assembly（命令行入口集成）

> 目标：组装所有包，实现完整的 CLI 工具。

| ID | 任务 | 产出文件 | 依赖 |
|---|---|---|---|
| T25 | 编写 `config.Load` 表格驱动测试：各种参数组合（URL 必填、output_file 位置参数、-o flag、-enable-reactions、-enable-user-links、GITHUB_TOKEN 环境变量、位置参数与 -o 冲突时以 -o 为准） | `internal/config/config_test.go` | T3 |
| T26 | 实现 `config.Load`：`flag` 解析 + 环境变量读取，使 T25 全部通过 | `internal/config/config.go` | T25 |
| T27 | 实现 `cmd/issue2md/main.go`：组装 config → parser → github.Client → converter → 输出（stdout/文件），exit code 规范（0/1/2），错误输出到 stderr | `cmd/issue2md/main.go` | T9, T16, T24, T26 |
| T28 | 编写 CLI 端到端测试：验证完整流程（成功转换 stdout、成功写入文件、无效 URL 报错 exit code 1、API 错误 exit code 2） | `cmd/issue2md/main_test.go` | T27 |

**Phase 4 依赖图：**

```
T3  ─→ T25 ─→ T26 ─┐
T9  ───────────────┤
T16 ───────────────┤→ T27 ─→ T28
T24 ───────────────┘
```

---

## 总结

| Phase | 任务数 | 测试任务 | 实现任务 | 关键验收 |
|---|---|---|---|---|
| 1. Foundation | 7 | 0 | 7 | 编译通过，类型定义完整 |
| 2. GitHub Fetcher | 10 | 4 | 6 | AC-1 (URL识别) 通过 |
| 3. Markdown Converter | 7 | 3 | 4 | AC-2/3/4/6 (转换+Flags) 通过 |
| 4. CLI Assembly | 4 | 2 | 2 | AC-5/7 (输出+错误) 通过 |
| **合计** | **28** | **9** | **19** | **全部 AC 通过** |
