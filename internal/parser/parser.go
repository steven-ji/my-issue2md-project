package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/steven-ji/issue2md/internal/github"
)

// ParsedURL 是 URL 解析的结果
type ParsedURL struct {
	Owner  string
	Repo   string
	Type   github.ResourceType
	Number int
}

// splitAndValidatePath 分割 URL 路径并验证基本结构
// 期望格式: owner/repo/type/number
func splitAndValidatePath(rawPath string) (owner, repo, typeStr, numberStr string, err error) {
	path := strings.Trim(rawPath, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("split path: unsupported path %q", rawPath)
	}

	owner, repo, typeStr, numberStr = parts[0], parts[1], parts[2], parts[3]
	if owner == "" || repo == "" {
		return "", "", "", "", fmt.Errorf("split path: missing owner or repo in path %q", rawPath)
	}

	return owner, repo, typeStr, numberStr, nil
}

// Parse 解析 GitHub URL，返回结构化的 ParsedURL
// 不支持的 URL 格式返回错误
func Parse(rawURL string) (*ParsedURL, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("parse url: empty input")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	if u.Host != "github.com" {
		return nil, fmt.Errorf("parse url: unsupported domain %q", u.Host)
	}

	owner, repo, typeStr, numberStr, err := splitAndValidatePath(u.Path)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	var resType github.ResourceType
	switch typeStr {
	case "issues":
		resType = github.TypeIssue
	case "pull":
		resType = github.TypePullRequest
	case "discussions":
		resType = github.TypeDiscussion
	default:
		return nil, fmt.Errorf("parse url: unsupported resource type %q", typeStr)
	}

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return nil, fmt.Errorf("parse url: invalid number %q: %w", numberStr, err)
	}

	return &ParsedURL{
		Owner:  owner,
		Repo:   repo,
		Type:   resType,
		Number: number,
	}, nil
}
