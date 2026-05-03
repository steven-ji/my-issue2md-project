package converter

import (
	"fmt"
	"strings"

	"github.com/steven-ji/issue2md/internal/github"
)

// Options 控制 Markdown 输出的可选内容
type Options struct {
	EnableReactions bool
	EnableUserLinks bool
}

// Convert 将 Resource 转换为完整 Markdown 文本（含 YAML Frontmatter）
// 纯函数，无副作用，不依赖 I/O
func Convert(res *github.Resource, opts Options) (string, error) {
	var buf strings.Builder

	// 1. YAML Frontmatter
	fm, err := generateFrontmatter(res)
	if err != nil {
		return "", fmt.Errorf("convert: %w", err)
	}
	buf.WriteString(fm)
	buf.WriteString("\n\n")

	// 2. 标题
	buf.WriteString("# " + res.Title + "\n\n")

	// 3. 元信息行
	buf.WriteString(formatMetaLine(res, opts))
	buf.WriteString("\n\n---\n\n")

	// 4. 正文
	buf.WriteString(res.Body + "\n")

	// 5. 主楼 Reactions
	if opts.EnableReactions && res.Reactions != nil {
		buf.WriteString(formatReactions(res.Reactions) + "\n")
	}

	// 6. 评论
	if len(res.Comments) > 0 {
		buf.WriteString("\n## Comments\n\n")
		for _, c := range res.Comments {
			buf.WriteString(formatComment(c, res, opts))
		}
	}

	return strings.TrimRight(buf.String(), "\n") + "\n", nil
}

// formatMetaLine 生成元信息行（Author | Created | State | Labels）
func formatMetaLine(res *github.Resource, opts Options) string {
	author := res.Author
	if opts.EnableUserLinks && author != "" {
		author = fmt.Sprintf("[%s](https://github.com/%s)", author, author)
	}

	state := capitalize(res.State)

	parts := []string{
		fmt.Sprintf("**Author:** %s", author),
		fmt.Sprintf("**Created:** %s", res.CreatedAt.Format("2006-01-02")),
		fmt.Sprintf("**State:** %s", state),
	}
	if len(res.Labels) > 0 {
		parts = append(parts, fmt.Sprintf("**Labels:** %s", strings.Join(res.Labels, ", ")))
	}

	return strings.Join(parts, " | ")
}

// formatComment 生成单条评论的 Markdown
func formatComment(c github.Comment, res *github.Resource, opts Options) string {
	var buf strings.Builder

	author := c.Author
	if opts.EnableUserLinks && author != "" {
		author = fmt.Sprintf("[%s](https://github.com/%s)", author, author)
	}

	prefix := "Comment by"
	if c.IsReview {
		prefix = "Review Comment by"
	}

	buf.WriteString(fmt.Sprintf("### %s %s — %s\n\n", prefix, author, c.CreatedAt.Format("2006-01-02T15:04:05Z")))
	buf.WriteString(c.Body + "\n")

	// Answer 标记
	if res.AnswerID != nil && c.ID == *res.AnswerID {
		buf.WriteString("\n> **✅ Answer**\n")
	}

	// 评论 Reactions
	if opts.EnableReactions && c.Reactions != nil {
		buf.WriteString(formatReactions(c.Reactions) + "\n")
	}

	buf.WriteString("\n")
	return buf.String()
}

// emojiOrder 定义 Reactions 输出的固定顺序
var emojiOrder = []string{"👍", "👎", "😄", "🎉", "😕", "❤️", "🚀", "👀"}

// formatReactions 将 Reactions map 格式化为 HTML 注释
func formatReactions(reactions map[string]int) string {
	var parts []string
	for _, emoji := range emojiOrder {
		if count, ok := reactions[emoji]; ok {
			parts = append(parts, fmt.Sprintf("%s %d", emoji, count))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf("<!-- Reactions: %s -->", strings.Join(parts, "  "))
}

// capitalize 将字符串首字母大写
func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
