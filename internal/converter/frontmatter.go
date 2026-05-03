package converter

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/steven-ji/issue2md/internal/github"
)

var frontmatterTpl = template.Must(template.New("frontmatter").Funcs(template.FuncMap{
	"yamlQuote":  yamlQuote,
	"yamlLabels": yamlLabels,
}).Parse(`---
title: {{ .Title | yamlQuote }}
url: {{ .URL }}
author: {{ .Author }}
created_at: {{ .CreatedAt.Format "2006-01-02T15:04:05Z" }}
state: {{ .State }}
labels:{{ .Labels | yamlLabels }}
type: {{ .Type }}
---`))

// yamlQuote 为 YAML 字符串值添加引号并转义特殊字符
func yamlQuote(s string) string {
	return fmt.Sprintf("%q", s)
}

// yamlLabels 将 Labels 切片格式化为 YAML 列表
func yamlLabels(labels []string) string {
	if len(labels) == 0 {
		return " []"
	}
	var b strings.Builder
	for _, l := range labels {
		b.WriteString("\n  - " + l)
	}
	return b.String()
}

// generateFrontmatter 生成 YAML Frontmatter 文本
func generateFrontmatter(res *github.Resource) (string, error) {
	var buf strings.Builder
	if err := frontmatterTpl.Execute(&buf, res); err != nil {
		return "", fmt.Errorf("generate frontmatter: %w", err)
	}
	return buf.String(), nil
}
