package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchDiscussion(t *testing.T) {
	fixedTime := time.Date(2025, 1, 17, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		graphQLHandler   http.HandlerFunc
		wantTitle        string
		wantAuthor       string
		wantBody         string
		wantComments     int
		wantAnswerMarked bool
	}{
		{
			name: "discussion with answer",
			graphQLHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := `{
					"data": {
						"repository": {
							"discussion": {
								"title": "How to configure buffer size?",
								"author": {"login": "dave"},
								"createdAt": "2025-01-17T14:00:00Z",
								"body": "I can't find the documentation.",
								"answer": {"id": "1"},
								"comments": {
									"pageInfo": {"hasNextPage": false, "endCursor": ""},
									"nodes": [
										{
											"id": "1",
											"author": {"login": "eve"},
											"createdAt": "2025-01-17T14:30:00Z",
											"body": "Set the BUFFER_SIZE env var.",
											"isAnswer": true
										},
										{
											"id": "2",
											"author": {"login": "dave"},
											"createdAt": "2025-01-17T14:45:00Z",
											"body": "Thanks!",
											"isAnswer": false
										}
									]
								}
							}
						}
					}
				}`
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(resp))
			},
			wantTitle:        "How to configure buffer size?",
			wantAuthor:       "dave",
			wantBody:         "I can't find the documentation.",
			wantComments:     2,
			wantAnswerMarked: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/", tt.graphQLHandler)

			server := httptest.NewServer(mux)
			defer server.Close()

			// 使用 newAPIClient 创建客户端，GraphQL URL 指向测试服务器
			client := newAPIClient("", server.URL, server.URL)

			resource, err := client.fetchDiscussion(context.Background(), "owner", "repo", 7, false)
			if err != nil {
				t.Fatalf("fetchDiscussion() error = %v", err)
			}

			if resource.Type != TypeDiscussion {
				t.Errorf("Type = %q, want %q", resource.Type, TypeDiscussion)
			}
			if resource.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", resource.Title, tt.wantTitle)
			}
			if resource.Author != tt.wantAuthor {
				t.Errorf("Author = %q, want %q", resource.Author, tt.wantAuthor)
			}
			if resource.Body != tt.wantBody {
				t.Errorf("Body = %q, want %q", resource.Body, tt.wantBody)
			}
			if len(resource.Comments) != tt.wantComments {
				t.Fatalf("len(Comments) = %d, want %d", len(resource.Comments), tt.wantComments)
			}
			if resource.CreatedAt != fixedTime {
				t.Errorf("CreatedAt = %v, want %v", resource.CreatedAt, fixedTime)
			}
			if tt.wantAnswerMarked && resource.AnswerID == nil {
				t.Error("expected AnswerID to be set, got nil")
			}
			if tt.wantAnswerMarked {
				answerFound := false
				for _, c := range resource.Comments {
					if c.ID == *resource.AnswerID {
						answerFound = true
						break
					}
				}
				if !answerFound {
					t.Error("AnswerID does not match any comment ID")
				}
			}
		})
	}
}
