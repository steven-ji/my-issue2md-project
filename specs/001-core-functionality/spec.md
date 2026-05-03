# issue2md Core Functionality Specification

## Overview

`issue2md` 是一个命令行工具，输入一个 GitHub Issue/PR/Discussion 的 URL，自动将其转换为 GitHub Flavored Markdown 文件。

---

## User Stories

### US-1: CLI 转换单个 Issue

作为开发者，我希望在终端输入 `issue2md https://github.com/owner/repo/issues/42`，工具自动识别这是一个 Issue，将其标题、元信息、正文和所有评论转换为 Markdown 并输出到 stdout，以便我可以用管道重定向或存档。

### US-2: CLI 转换 PR

作为开发者，我希望输入一个 PR 的 URL，工具自动识别类型，将 PR 描述、Review Comments 和普通 Comments 按时间线平铺输出，以便我归档 PR 中的讨论过程。

### US-3: CLI 转换 Discussion

作为开发者，我希望输入一个 Discussion 的 URL，工具自动识别类型，将 Discussion 及其所有回复（含嵌套）转换为 Markdown，被标记为 Answer 的评论有显著标记。

### US-4: 输出到文件

作为开发者，我希望通过 `issue2md <url> output.md` 或 `issue2md <url> -o output.md` 将结果直接写入文件，而不必手动重定向 stdout。

### US-5: 包含 Reactions

作为开发者，我希望通过 `-enable-reactions` 标志在输出中包含每条内容的 Reactions 统计，以便了解社区反馈。

### US-6: 用户链接

作为开发者，我希望通过 `-enable-user-links` 标志将用户名渲染为指向其 GitHub 主页的链接，以便快速定位参与者。

### US-7: [Future] Web 界面

作为用户，我希望通过 Web 界面粘贴 URL 并下载 Markdown 文件。（本迭代不实现，仅作为未来规划记录。）

---

## Functional Requirements

### FR-1: URL 自动识别

- 工具必须自动解析 URL 结构来判断资源类型（Issue / PR / Discussion），无需用户手动指定。
- 识别规则：

| URL Pattern | 类型 |
|---|---|
| `github.com/{owner}/{repo}/issues/{number}` | Issue |
| `github.com/{owner}/{repo}/pull/{number}` | Pull Request |
| `github.com/{owner}/{repo}/discussions/{number}` | Discussion |

- 不支持的 URL 格式必须报错退出。

### FR-2: 认证

- 仅通过环境变量 `GITHUB_TOKEN` 传入 Personal Access Token。
- **不提供** `--token` 命令行参数，防止密钥泄露到 Shell 历史。
- 未提供 Token 时，仅支持公有仓库访问。
- Token 无效或权限不足时，透传 GitHub API 错误信息。

### FR-3: 数据获取

- **Issue:** 获取标题、作者、创建时间、状态（Open/Closed）、标签、正文、所有评论。
- **PR:** 获取标题、作者、创建时间、状态（Open/Closed/Merged）、标签、正文、Review Comments 和普通 Comments，按时间正序平铺展示。不获取 diff 或 commits 信息。
- **Discussion:** 获取标题、作者、创建时间、正文、所有回复（含嵌套）。被标记为 Answer 的评论需显著标记。
- 分页：必须自动获取所有评论，不设上限。GitHub API 默认每页 30 条，工具需自动翻页直到获取全部。

### FR-4: 输出格式

- 输出标准 GitHub Flavored Markdown。
- 文件头部包含 YAML Frontmatter，便于静态站点生成器（Hugo/Jekyll 等）解析。
- 图片和附件保留原始远程链接，不下载到本地。

### FR-5: 命令行接口

```
issue2md [flags] <url> [output_file]
```

**位置参数：**

| 参数 | 必填 | 说明 |
|---|---|---|
| `<url>` | 是 | GitHub Issue/PR/Discussion 的 URL |
| `[output_file]` | 否 | 输出文件路径。未提供时输出到 stdout |

**Flags：**

| Flag | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `-enable-reactions` | bool | false | 在主楼和评论下方显示 Reactions 统计 |
| `-enable-user-links` | bool | false | 将用户名渲染为 GitHub 主页链接 |
| `-o` | string | "" | 输出文件路径（与位置参数 `output_file` 功能等价） |

**冲突处理：** 若同时提供位置参数 `output_file` 和 `-o` flag，以 `-o` 为准。

---

## Non-Functional Requirements

### NFR-1: 架构与解耦

- GitHub API 交互逻辑与 Markdown 渲染逻辑必须分离，便于未来替换数据源或输出格式。
- 使用 Go 标准库 `net/http` 作为 HTTP 客户端，不引入第三方 HTTP 库。
- 不使用全局变量传递状态，所有依赖通过函数参数或结构体成员显式注入。

### NFR-2: 错误处理

- 所有错误必须显式处理，使用 `fmt.Errorf("...: %w", err)` 包装。
- 错误信息输出到 stderr，正常内容输出到 stdout。
- URL 无效或资源不存在：以非零 exit code 退出，stderr 输出清晰错误信息。
- GitHub API 限流：透传 API 错误信息，不做重试。
- 网络不可达：输出友好错误信息，以非零 exit code 退出。

### NFR-3: 可测试性

- 核心逻辑（URL 解析、Markdown 渲染）必须可单元测试，不依赖真实 GitHub API 调用。
- 优先编写表格驱动测试（Table-Driven Tests）。

---

## Acceptance Criteria

### AC-1: URL 识别

- [ ] 输入有效的 Issue URL，正确识别为 Issue 类型
- [ ] 输入有效的 PR URL，正确识别为 PR 类型
- [ ] 输入有效的 Discussion URL，正确识别为 Discussion 类型
- [ ] 输入无效的 URL，输出错误信息到 stderr 并以非零 exit code 退出
- [ ] 输入非 GitHub 域名的 URL，输出错误信息到 stderr 并以非零 exit code 退出

### AC-2: Issue 转换

- [ ] 输出包含 YAML Frontmatter（title, url, author, created_at, state, labels）
- [ ] 输出包含标题（一级标题）、作者、创建时间、状态、标签
- [ ] 输出包含正文（保留原始 Markdown 格式）
- [ ] 输出包含所有评论（按时间正序）
- [ ] 每条评论包含作者、时间、内容

### AC-3: PR 转换

- [ ] 输出包含 PR 描述和所有 Comments
- [ ] Review Comments 与普通 Comments 按时间线平铺展示
- [ ] 不包含 diff 或 commits 信息

### AC-4: Discussion 转换

- [ ] 输出包含 Discussion 正文和所有回复
- [ ] 被标记为 Answer 的评论有显著标记

### AC-5: 输出目标

- [ ] 不提供 output_file 时，内容输出到 stdout
- [ ] 提供 output_file 位置参数时，内容写入文件
- [ ] 提供 `-o` flag 时，内容写入指定文件
- [ ] 同时提供位置参数和 `-o` 时，以 `-o` 为准

### AC-6: Flags

- [ ] 默认不包含 Reactions
- [ ] `-enable-reactions` 开启后，主楼和每条评论下方显示 Reactions 统计
- [ ] 默认用户名为纯文本
- [ ] `-enable-user-links` 开启后，用户名渲染为 `[username](https://github.com/username)`

### AC-7: 认证与错误

- [ ] 未设置 `GITHUB_TOKEN` 时，可访问公有仓库
- [ ] `GITHUB_TOKEN` 无效时，透传 GitHub API 错误信息
- [ ] 资源不存在时，输出清晰错误信息到 stderr 并以非零 exit code 退出

---

## Output Format Example

### Issue

```markdown
---
title: "Bug: Connection timeout on large payloads"
url: https://github.com/owner/repo/issues/42
author: alice
created_at: 2025-01-15T10:30:00Z
state: open
labels:
  - bug
  - priority:high
type: issue
---

# Bug: Connection timeout on large payloads

**Author:** alice | **Created:** 2025-01-15 | **State:** Open | **Labels:** bug, priority:high

---

When sending payloads larger than 10MB, the connection consistently times out after 30 seconds.

## Comments

### Comment by bob — 2025-01-15T11:00:00Z

I can reproduce this. It seems related to the buffer size configuration.

### Comment by alice — 2025-01-15T11:15:00Z

> I can reproduce this. It seems related to the buffer size configuration.

Yes, increasing the buffer size to 20MB resolves it temporarily but is not a proper fix.

<!-- Reactions (when -enable-reactions is set):
👍 5  🎉 2
-->
```

### Pull Request

```markdown
---
title: "Fix: Increase buffer size for large payloads"
url: https://github.com/owner/repo/pull/43
author: bob
created_at: 2025-01-16T09:00:00Z
state: merged
labels:
  - fix
type: pull_request
---

# Fix: Increase buffer size for large payloads

**Author:** bob | **Created:** 2025-01-16 | **State:** Merged | **Labels:** fix

---

This PR increases the default buffer size from 10MB to 50MB to handle large payloads without timeout.

## Comments

### Comment by alice — 2025-01-16T09:30:00Z

LGTM, tested with 40MB payload.

### Review Comment by charlie — 2025-01-16T10:00:00Z

Should we also add a configurable max payload size? This feels like a hardcoded value.

### Comment by bob — 2025-01-16T10:15:00Z

Good point, I'll add a config option in a follow-up PR.
```

### Discussion (with Answer)

```markdown
---
title: "How to configure buffer size?"
url: https://github.com/owner/repo/discussions/7
author: dave
created_at: 2025-01-17T14:00:00Z
state: open
labels: []
type: discussion
---

# How to configure buffer size?

**Author:** dave | **Created:** 2025-01-17 | **State:** Open

---

I can't find documentation on how to configure the buffer size. Can someone point me in the right direction?

## Comments

### Comment by eve — 2025-01-17T14:30:00Z

You can set it via the `BUFFER_SIZE` environment variable.

> **✅ Answer**
```
