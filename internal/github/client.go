package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	gh "github.com/google/go-github/v72/github"
)

// Client 是 GitHub API 的抽象接口
type Client interface {
	FetchResource(ctx context.Context, owner, repo string, resType ResourceType, number int, enableReactions bool) (*Resource, error)
}

// APIClient 是基于 GitHub REST/GraphQL API 的 Client 实现
type APIClient struct {
	restClient *gh.Client
	httpClient *http.Client
	graphQLURL string
}

// FetchResource 根据 resType 分派到对应的 fetch 方法
func (c *APIClient) FetchResource(ctx context.Context, owner, repo string, resType ResourceType, number int, enableReactions bool) (*Resource, error) {
	switch resType {
	case TypeIssue:
		return c.fetchIssue(ctx, owner, repo, number, enableReactions)
	case TypePullRequest:
		return c.fetchPullRequest(ctx, owner, repo, number, enableReactions)
	case TypeDiscussion:
		return c.fetchDiscussion(ctx, owner, repo, number, enableReactions)
	default:
		return nil, fmt.Errorf("fetch resource: unsupported type %q", resType)
	}
}

// NewClient 创建 APIClient 实例，token 为空时匿名访问（仅公有仓库）
func NewClient(token string) Client {
	return newAPIClient(token, "https://api.github.com", "https://api.github.com/graphql")
}

// newAPIClient 创建指向指定 base URL 的 APIClient（用于测试）
func newAPIClient(token, restBaseURL, graphQLURL string) *APIClient {
	transport := http.DefaultTransport
	if token != "" {
		transport = &tokenTransport{token: token, fallback: http.DefaultTransport}
	}
	httpClient := &http.Client{Transport: transport}

	restClient := gh.NewClient(httpClient)
	// go-github 要求 BaseURL 必须以 "/" 结尾
	if !strings.HasSuffix(restBaseURL, "/") {
		restBaseURL += "/"
	}
	baseURL, err := url.Parse(restBaseURL)
	if err != nil {
		panic(fmt.Sprintf("internal error: invalid base URL %q: %v", restBaseURL, err))
	}
	restClient.BaseURL = baseURL

	return &APIClient{
		restClient: restClient,
		httpClient: httpClient,
		graphQLURL: graphQLURL,
	}
}

// tokenTransport 为请求添加 Bearer token 认证头
type tokenTransport struct {
	token    string
	fallback http.RoundTripper
}

func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.fallback.RoundTrip(req)
}
