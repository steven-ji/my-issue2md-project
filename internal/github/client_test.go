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

func TestAPIClient_FetchResource(t *testing.T) {
	fixedTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	issueResp := &gh.Issue{
		Number:    ptr(42),
		Title:     ptr("Test Issue"),
		State:     ptr("open"),
		HTMLURL:   ptr("https://github.com/owner/repo/issues/42"),
		Body:      ptr("Issue body"),
		CreatedAt: &gh.Timestamp{Time: fixedTime},
		User:      &gh.User{Login: ptr("alice")},
	}
	commentsResp := []*gh.IssueComment{}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/owner/repo/issues/42":
			json.NewEncoder(w).Encode(issueResp)
		case "/repos/owner/repo/issues/42/comments":
			json.NewEncoder(w).Encode(commentsResp)
		default:
			// GraphQL 端点
			resp := `{
				"data": {
					"repository": {
						"discussion": {
							"title": "A Discussion",
							"author": {"login": "frank"},
							"createdAt": "2025-01-17T14:00:00Z",
							"body": "Discussion body",
							"answer": null,
							"comments": {
								"pageInfo": {"hasNextPage": false, "endCursor": ""},
								"nodes": []
							}
						}
					}
				}
			}`
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(resp))
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := newAPIClient("", server.URL, server.URL)

	tests := []struct {
		name     string
		resType  ResourceType
		number   int
		wantType ResourceType
		wantErr  bool
	}{
		{
			name:     "fetch issue via FetchResource",
			resType:  TypeIssue,
			number:   42,
			wantType: TypeIssue,
			wantErr:  false,
		},
		{
			name:     "fetch discussion via FetchResource",
			resType:  TypeDiscussion,
			number:   7,
			wantType: TypeDiscussion,
			wantErr:  false,
		},
		{
			name:    "unsupported resource type",
			resType: ResourceType("unknown"),
			number:  1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource, err := client.FetchResource(context.Background(), "owner", "repo", tt.resType, tt.number, false)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resource.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", resource.Type, tt.wantType)
			}
		})
	}
}
