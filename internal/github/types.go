package github

import "time"

// ResourceType 表示 GitHub 资源类型
type ResourceType string

const (
	TypeIssue       ResourceType = "issue"
	TypePullRequest ResourceType = "pull_request"
	TypeDiscussion  ResourceType = "discussion"
)

// Resource 统一表示 GitHub 资源的元信息与内容
type Resource struct {
	Type      ResourceType
	Title     string
	Author    string
	URL       string
	State     string // "open", "closed", "merged"
	Labels    []string
	CreatedAt time.Time
	Body      string
	Comments  []Comment
	Reactions map[string]int // 主楼 Reactions，emoji → count
	AnswerID  *int64         // Discussion 中被标记为 Answer 的评论 ID，nil 表示无
}

// Comment 统一表示所有类型的评论
type Comment struct {
	ID        int64
	Author    string
	CreatedAt time.Time
	Body      string
	IsReview  bool           // 是否为 PR Review Comment
	Reactions map[string]int // emoji → count，仅当启用 reactions 时填充
}
