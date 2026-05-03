package converter

import (
	"strings"
	"testing"
	"time"

	"github.com/steven-ji/issue2md/internal/github"
)

func TestGenerateFrontmatter(t *testing.T) {
	fixedTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		resource *github.Resource
		wantContains []string
		wantNotContains []string
	}{
		{
			name: "basic issue frontmatter",
			resource: &github.Resource{
				Type:      github.TypeIssue,
				Title:     "Bug: Connection timeout",
				Author:    "alice",
				URL:       "https://github.com/owner/repo/issues/42",
				State:     "open",
				Labels:    []string{"bug", "priority:high"},
				CreatedAt: fixedTime,
			},
			wantContains: []string{
				"title: \"Bug: Connection timeout\"",
				"url: https://github.com/owner/repo/issues/42",
				"author: alice",
				"created_at: 2025-01-15T10:30:00Z",
				"state: open",
				"- bug",
				"- priority:high",
				"type: issue",
			},
		},
		{
			name: "title with quotes",
			resource: &github.Resource{
				Title:     `He said "hello"`,
				Type:      github.TypeIssue,
				Author:    "bob",
				URL:       "https://github.com/owner/repo/issues/1",
				State:     "closed",
				Labels:    nil,
				CreatedAt: fixedTime,
			},
			wantContains: []string{
				"title:",
				"state: closed",
				"labels: []",
			},
		},
		{
			name: "empty labels",
			resource: &github.Resource{
				Type:      github.TypeDiscussion,
				Title:     "How to?",
				Author:    "dave",
				URL:       "https://github.com/owner/repo/discussions/7",
				State:     "open",
				Labels:    nil,
				CreatedAt: fixedTime,
			},
			wantContains: []string{
				"labels: []",
				"type: discussion",
			},
			wantNotContains: []string{"- "},
		},
		{
			name: "pull request type",
			resource: &github.Resource{
				Type:      github.TypePullRequest,
				Title:     "Fix: something",
				Author:    "bob",
				URL:       "https://github.com/owner/repo/pull/43",
				State:     "merged",
				Labels:    []string{"fix"},
				CreatedAt: fixedTime,
			},
			wantContains: []string{
				"type: pull_request",
				"state: merged",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateFrontmatter(tt.resource)
			if err != nil {
				t.Fatalf("generateFrontmatter() error = %v", err)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("frontmatter missing %q\nGot:\n%s", want, got)
				}
			}
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(got, notWant) {
					t.Errorf("frontmatter should not contain %q\nGot:\n%s", notWant, got)
				}
			}
		})
	}
}
