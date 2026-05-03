package github

import (
	"context"
	"fmt"
	"sort"

	gh "github.com/google/go-github/v72/github"
)

// fetchPullRequest 获取 PR 数据（Issue Comments + Review Comments 合并，按时间排序）
func (c *APIClient) fetchPullRequest(ctx context.Context, owner, repo string, number int, enableReactions bool) (*Resource, error) {
	pr, _, err := c.restClient.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("fetch pull request %s/%s/%d: %w", owner, repo, number, err)
	}

	issueComments, err := c.fetchPRIssueComments(ctx, owner, repo, number, enableReactions)
	if err != nil {
		return nil, fmt.Errorf("fetch PR issue comments %s/%s/%d: %w", owner, repo, number, err)
	}

	reviewComments, err := c.fetchReviewComments(ctx, owner, repo, number, enableReactions)
	if err != nil {
		return nil, fmt.Errorf("fetch PR review comments %s/%s/%d: %w", owner, repo, number, err)
	}

	// 合并两种 Comments，Review Comments 标记 IsReview=true
	comments := make([]Comment, 0, len(issueComments)+len(reviewComments))
	comments = append(comments, issueComments...)
	comments = append(comments, reviewComments...)
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})

	var labels []string
	for _, l := range pr.Labels {
		labels = append(labels, l.GetName())
	}

	state := pr.GetState()
	if pr.GetMerged() {
		state = "merged"
	}

	resource := &Resource{
		Type:      TypePullRequest,
		Title:     pr.GetTitle(),
		Author:    pr.GetUser().GetLogin(),
		URL:       pr.GetHTMLURL(),
		State:     state,
		Labels:    labels,
		CreatedAt: pr.GetCreatedAt().Time,
		Body:      pr.GetBody(),
		Comments:  comments,
	}

	if enableReactions {
		// PR 主楼 Reactions 需通过 Issue API 获取（PR 本质也是 Issue）
		issue, _, err := c.restClient.Issues.Get(ctx, owner, repo, number)
		if err != nil {
			return nil, fmt.Errorf("fetch PR reactions %s/%s/%d: %w", owner, repo, number, err)
		}
		if issue.Reactions != nil {
			resource.Reactions = reactionsMap(issue.Reactions)
		}
	}

	return resource, nil
}

// fetchPRIssueComments 获取 PR 的 Issue Comments（PR 级别评论）
func (c *APIClient) fetchPRIssueComments(ctx context.Context, owner, repo string, number int, enableReactions bool) ([]Comment, error) {
	var allComments []Comment
	opts := &gh.IssueListCommentsOptions{
		ListOptions: gh.ListOptions{PerPage: 100},
	}

	for {
		comments, resp, err := c.restClient.Issues.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, fmt.Errorf("list PR issue comments: %w", err)
		}

		for _, c := range comments {
			comment := Comment{
				ID:        c.GetID(),
				Author:    c.GetUser().GetLogin(),
				CreatedAt: c.GetCreatedAt().Time,
				Body:      c.GetBody(),
				IsReview:  false,
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

// fetchReviewComments 获取 PR 的 Review Comments（代码行级别评论）
func (c *APIClient) fetchReviewComments(ctx context.Context, owner, repo string, number int, enableReactions bool) ([]Comment, error) {
	var allComments []Comment
	opts := &gh.PullRequestListCommentsOptions{
		ListOptions: gh.ListOptions{PerPage: 100},
	}

	for {
		comments, resp, err := c.restClient.PullRequests.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, fmt.Errorf("list PR review comments: %w", err)
		}

		for _, c := range comments {
			comment := Comment{
				ID:        c.GetID(),
				Author:    c.GetUser().GetLogin(),
				CreatedAt: c.GetCreatedAt().Time,
				Body:      c.GetBody(),
				IsReview:  true,
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
