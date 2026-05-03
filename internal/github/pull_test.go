package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	gh "github.com/google/go-github/v72/github"
)

func TestFetchPullRequest(t *testing.T) {
	fixedTime := time.Date(2025, 1, 16, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name               string
		prResp             *gh.PullRequest
		issueCommentsResp  []*gh.IssueComment
		reviewCommentsResp []*gh.PullRequestComment
		enableReactions    bool
		wantTitle          string
		wantAuthor         string
		wantState          string
		wantBody           string
		wantComments       int
		wantReviewCount    int
	}{
		{
			name: "PR with both issue and review comments",
			prResp: &gh.PullRequest{
				Number:    ptr(43),
				Title:     ptr("Fix: Increase buffer size"),
				State:     ptr("closed"),
				Merged:    ptr(true),
				HTMLURL:   ptr("https://github.com/owner/repo/pull/43"),
				Body:      ptr("This PR increases the buffer size."),
				CreatedAt: &gh.Timestamp{Time: fixedTime},
				User:      &gh.User{Login: ptr("bob")},
				Labels:    []*gh.Label{{Name: ptr("fix")}},
			},
			issueCommentsResp: []*gh.IssueComment{
				{
					ID:        ptr(int64(100)),
					Body:      ptr("LGTM"),
					CreatedAt: &gh.Timestamp{Time: fixedTime.Add(30 * time.Minute)},
					User:      &gh.User{Login: ptr("alice")},
				},
				{
					ID:        ptr(int64(102)),
					Body:      ptr("Good point, will follow up."),
					CreatedAt: &gh.Timestamp{Time: fixedTime.Add(75 * time.Minute)},
					User:      &gh.User{Login: ptr("bob")},
				},
			},
			reviewCommentsResp: []*gh.PullRequestComment{
				{
					ID:        ptr(int64(101)),
					Body:      ptr("Should we add a config option?"),
					CreatedAt: &gh.Timestamp{Time: fixedTime.Add(time.Hour)},
					User:      &gh.User{Login: ptr("charlie")},
				},
			},
			enableReactions: false,
			wantTitle:       "Fix: Increase buffer size",
			wantAuthor:      "bob",
			wantState:       "merged",
			wantBody:        "This PR increases the buffer size.",
			wantComments:    3,
			wantReviewCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/repos/owner/repo/pulls/43":
					json.NewEncoder(w).Encode(tt.prResp)
				case "/repos/owner/repo/issues/43/comments":
					json.NewEncoder(w).Encode(tt.issueCommentsResp)
				case "/repos/owner/repo/pulls/43/comments":
					json.NewEncoder(w).Encode(tt.reviewCommentsResp)
				default:
					http.NotFound(w, r)
				}
			})

			server := httptest.NewServer(mux)
			defer server.Close()

			client := newAPIClient("", server.URL, server.URL+"/graphql")
			resource, err := client.fetchPullRequest(context.Background(), "owner", "repo", 43, tt.enableReactions)
			if err != nil {
				t.Fatalf("fetchPullRequest() error = %v", err)
			}

			if resource.Type != TypePullRequest {
				t.Errorf("Type = %q, want %q", resource.Type, TypePullRequest)
			}
			if resource.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", resource.Title, tt.wantTitle)
			}
			if resource.Author != tt.wantAuthor {
				t.Errorf("Author = %q, want %q", resource.Author, tt.wantAuthor)
			}
			if resource.State != tt.wantState {
				t.Errorf("State = %q, want %q", resource.State, tt.wantState)
			}
			if resource.Body != tt.wantBody {
				t.Errorf("Body = %q, want %q", resource.Body, tt.wantBody)
			}
			if len(resource.Comments) != tt.wantComments {
				t.Fatalf("len(Comments) = %d, want %d", len(resource.Comments), tt.wantComments)
			}

			// 验证 Review Comments 标记
			reviewCount := 0
			for _, c := range resource.Comments {
				if c.IsReview {
					reviewCount++
				}
			}
			if reviewCount != tt.wantReviewCount {
				t.Errorf("review comments count = %d, want %d", reviewCount, tt.wantReviewCount)
			}

			// 验证评论按时间正序排列
			if !sort.IsSorted(byCreatedAt(resource.Comments)) {
				t.Error("comments are not sorted by CreatedAt ascending")
			}
		})
	}
}

// byCreatedAt 实现排序接口用于验证评论时间顺序
type byCreatedAt []Comment

func (a byCreatedAt) Len() int           { return len(a) }
func (a byCreatedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byCreatedAt) Less(i, j int) bool { return a[i].CreatedAt.Before(a[j].CreatedAt) }
