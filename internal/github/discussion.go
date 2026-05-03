package github

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/shurcooL/graphql"
)

// discussionQuery 是 GraphQL 查询的结构定义
type discussionQuery struct {
	Repository struct {
		Discussion struct {
			Title     string `graphql:"title"`
			Author    *struct{ Login string } `graphql:"author"`
			CreatedAt string `graphql:"createdAt"`
			Body      string `graphql:"body"`
			Answer    *struct {
				ID string `graphql:"id"`
			} `graphql:"answer"`
			Comments struct {
				PageInfo struct {
					HasNextPage bool   `graphql:"hasNextPage"`
					EndCursor   string `graphql:"endCursor"`
				} `graphql:"pageInfo"`
				Nodes []discussionComment `graphql:"nodes"`
			} `graphql:"comments(first: 100, after: $cursor)"`
		} `graphql:"discussion(number: $number)"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

type discussionComment struct {
	ID        string  `graphql:"id"`
	Author    *struct{ Login string } `graphql:"author"`
	CreatedAt string  `graphql:"createdAt"`
	Body      string  `graphql:"body"`
	IsAnswer  bool    `graphql:"isAnswer"`
}

// fetchDiscussion 获取 Discussion 数据（通过 GraphQL API）
func (c *APIClient) fetchDiscussion(ctx context.Context, owner, repo string, number int, enableReactions bool) (*Resource, error) {
	gqlClient := graphql.NewClient(c.graphQLURL, c.httpClient)

	var (
		title      string
		author     string
		createdAt  time.Time
		body       string
		answerID   *int64
		comments   []Comment
		cursor     string
		initialized bool
	)

	for {
		var q discussionQuery
		variables := map[string]interface{}{
			"owner":  graphql.String(owner),
			"repo":   graphql.String(repo),
			"number": graphql.Int(number),
			"cursor": graphql.String(cursor),
		}

		err := gqlClient.Query(ctx, &q, variables)
		if err != nil {
			return nil, fmt.Errorf("fetch discussion %s/%s/%d: %w", owner, repo, number, err)
		}

		d := q.Repository.Discussion

		if !initialized {
			title = d.Title
			if d.Author != nil {
				author = d.Author.Login
			}
			createdAt, err = time.Parse(time.RFC3339, d.CreatedAt)
				if err != nil {
					return nil, fmt.Errorf("parse discussion created_at: %w", err)
				}
			body = d.Body
			if d.Answer != nil {
				if id, err := strconv.ParseInt(d.Answer.ID, 10, 64); err == nil {
					answerID = &id
				}
			}
			initialized = true
		}

		for _, node := range d.Comments.Nodes {
			c, err := discussionNodeToComment(node)
			if err != nil {
				return nil, fmt.Errorf("fetch discussion %s/%s/%d: %w", owner, repo, number, err)
			}
			comments = append(comments, c)
		}

		if !d.Comments.PageInfo.HasNextPage {
			break
		}
		cursor = d.Comments.PageInfo.EndCursor
	}

	return &Resource{
		Type:      TypeDiscussion,
		Title:     title,
		Author:    author,
		URL:       fmt.Sprintf("https://github.com/%s/%s/discussions/%d", owner, repo, number),
		State:     "open",
		CreatedAt: createdAt,
		Body:      body,
		Comments:  comments,
		AnswerID:  answerID,
	}, nil
}

func discussionNodeToComment(node discussionComment) (Comment, error) {
	author := ""
	if node.Author != nil {
		author = node.Author.Login
	}
	createdAt, err := time.Parse(time.RFC3339, node.CreatedAt)
	if err != nil {
		return Comment{}, fmt.Errorf("parse comment created_at: %w", err)
	}
	id, err := strconv.ParseInt(node.ID, 10, 64)
	if err != nil {
		return Comment{}, fmt.Errorf("parse comment id %q: %w", node.ID, err)
	}

	return Comment{
		ID:        id,
		Author:    author,
		CreatedAt: createdAt,
		Body:      node.Body,
	}, nil
}
