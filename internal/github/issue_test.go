package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gh "github.com/google/go-github/v72/github"
)

func TestFetchIssue(t *testing.T) {
	fixedTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name            string
		number          int
		issueResp       *gh.Issue
		commentsResp    []*gh.IssueComment
		enableReactions bool
		wantTitle       string
		wantAuthor      string
		wantState       string
		wantLabels      []string
		wantBody        string
		wantComments    int
		wantReactions   bool
	}{
		{
			name:   "basic issue without reactions",
			number: 42,
			issueResp: &gh.Issue{
				Number:    ptr(42),
				Title:     ptr("Test Issue"),
				State:     ptr("open"),
				HTMLURL:   ptr("https://github.com/owner/repo/issues/42"),
				Body:      ptr("Issue body"),
				CreatedAt: &gh.Timestamp{Time: fixedTime},
				User:      &gh.User{Login: ptr("alice")},
				Labels:    []*gh.Label{{Name: ptr("bug")}, {Name: ptr("priority:high")}},
			},
			commentsResp: []*gh.IssueComment{
				{
					ID:        ptr(int64(1)),
					Body:      ptr("First comment"),
					CreatedAt: &gh.Timestamp{Time: fixedTime.Add(time.Hour)},
					User:      &gh.User{Login: ptr("bob")},
				},
				{
					ID:        ptr(int64(2)),
					Body:      ptr("Second comment"),
					CreatedAt: &gh.Timestamp{Time: fixedTime.Add(2 * time.Hour)},
					User:      &gh.User{Login: ptr("alice")},
				},
			},
			enableReactions: false,
			wantTitle:       "Test Issue",
			wantAuthor:      "alice",
			wantState:       "open",
			wantLabels:      []string{"bug", "priority:high"},
			wantBody:        "Issue body",
			wantComments:    2,
			wantReactions:   false,
		},
		{
			name:   "issue with reactions enabled",
			number: 7,
			issueResp: &gh.Issue{
				Number:    ptr(7),
				Title:     ptr("Bug Report"),
				State:     ptr("closed"),
				HTMLURL:   ptr("https://github.com/owner/repo/issues/7"),
				Body:      ptr("Closed issue"),
				CreatedAt: &gh.Timestamp{Time: fixedTime},
				User:      &gh.User{Login: ptr("charlie")},
				Reactions: &gh.Reactions{
					TotalCount: ptr(3),
					PlusOne:    ptr(2),
					Heart:      ptr(1),
				},
			},
			commentsResp: []*gh.IssueComment{
				{
					ID:        ptr(int64(10)),
					Body:      ptr("Nice fix"),
					CreatedAt: &gh.Timestamp{Time: fixedTime.Add(time.Hour)},
					User:      &gh.User{Login: ptr("dave")},
					Reactions: &gh.Reactions{
						TotalCount: ptr(1),
						Hooray:     ptr(1),
					},
				},
			},
			enableReactions: true,
			wantTitle:       "Bug Report",
			wantAuthor:      "charlie",
			wantState:       "closed",
			wantLabels:      nil,
			wantBody:        "Closed issue",
			wantComments:    1,
			wantReactions:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/repos/owner/repo/issues/" + itoa(tt.number):
					json.NewEncoder(w).Encode(tt.issueResp)
				case "/repos/owner/repo/issues/" + itoa(tt.number) + "/comments":
					json.NewEncoder(w).Encode(tt.commentsResp)
				default:
					http.NotFound(w, r)
				}
			})

			server := httptest.NewServer(mux)
			defer server.Close()

			client := newAPIClient("", server.URL, server.URL+"/graphql")
			resource, err := client.fetchIssue(context.Background(), "owner", "repo", tt.number, tt.enableReactions)
			if err != nil {
				t.Fatalf("fetchIssue() error = %v", err)
			}

			if resource.Type != TypeIssue {
				t.Errorf("Type = %q, want %q", resource.Type, TypeIssue)
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
			if len(resource.Labels) != len(tt.wantLabels) {
				t.Errorf("len(Labels) = %d, want %d", len(resource.Labels), len(tt.wantLabels))
			}
			for i, label := range resource.Labels {
				if i < len(tt.wantLabels) && label != tt.wantLabels[i] {
					t.Errorf("Labels[%d] = %q, want %q", i, label, tt.wantLabels[i])
				}
			}
			if tt.wantReactions {
				if len(resource.Comments) > 0 && resource.Comments[0].Reactions == nil {
					t.Error("expected Reactions to be populated, got nil")
				}
			} else {
				if len(resource.Comments) > 0 && resource.Comments[0].Reactions != nil {
					t.Error("expected Reactions to be nil, got non-nil")
				}
			}
		})
	}
}
