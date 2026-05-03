package github

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v72/github"
)

// fetchIssue 获取 Issue 数据（含所有评论，自动分页）
func (c *APIClient) fetchIssue(ctx context.Context, owner, repo string, number int, enableReactions bool) (*Resource, error) {
	issue, _, err := c.restClient.Issues.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("fetch issue %s/%s/%d: %w", owner, repo, number, err)
	}

	comments, err := c.fetchIssueComments(ctx, owner, repo, number, enableReactions)
	if err != nil {
		return nil, fmt.Errorf("fetch issue comments %s/%s/%d: %w", owner, repo, number, err)
	}

	var labels []string
	for _, l := range issue.Labels {
		labels = append(labels, l.GetName())
	}

	resource := &Resource{
		Type:      TypeIssue,
		Title:     issue.GetTitle(),
		Author:    issue.GetUser().GetLogin(),
		URL:       issue.GetHTMLURL(),
		State:     issue.GetState(),
		Labels:    labels,
		CreatedAt: issue.GetCreatedAt().Time,
		Body:      issue.GetBody(),
		Comments:  comments,
	}

	if enableReactions {
		resource.Reactions = reactionsMap(issue.Reactions)
	}

	return resource, nil
}

// fetchIssueComments 获取 Issue 的所有评论（自动分页）
func (c *APIClient) fetchIssueComments(ctx context.Context, owner, repo string, number int, enableReactions bool) ([]Comment, error) {
	var allComments []Comment
	opts := &gh.IssueListCommentsOptions{
		ListOptions: gh.ListOptions{PerPage: 100},
	}

	for {
		comments, resp, err := c.restClient.Issues.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, fmt.Errorf("list issue comments: %w", err)
		}

		for _, c := range comments {
			comment := Comment{
				ID:        c.GetID(),
				Author:    c.GetUser().GetLogin(),
				CreatedAt: c.GetCreatedAt().Time,
				Body:      c.GetBody(),
			}
			if enableReactions {
				comment.Reactions = reactionsMap(c.Reactions)
			}
			allComments = append(allComments, comment)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allComments, nil
}

// reactionsMap 将 go-github Reactions 转换为 emoji → count 映射
func reactionsMap(r *gh.Reactions) map[string]int {
	if r == nil {
		return nil
	}
	m := map[string]int{
		"👍": r.GetPlusOne(),
		"👎": r.GetMinusOne(),
		"😄": r.GetLaugh(),
		"🎉": r.GetHooray(),
		"😕": r.GetConfused(),
		"❤️": r.GetHeart(),
		"🚀": r.GetRocket(),
		"👀": r.GetEyes(),
	}
	// 移除 count 为 0 的条目
	for k, v := range m {
		if v == 0 {
			delete(m, k)
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}
