package converter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/steven-ji/issue2md/internal/github"
)

func TestConvert(t *testing.T) {
	fixedTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		resource       *github.Resource
		opts           Options
		goldenFile     string
		wantContains   []string
		wantNotContains []string
	}{
		{
			name: "issue matches golden file",
			resource: &github.Resource{
				Type:      github.TypeIssue,
				Title:     "Bug: Connection timeout on large payloads",
				Author:    "alice",
				URL:       "https://github.com/owner/repo/issues/42",
				State:     "open",
				Labels:    []string{"bug", "priority:high"},
				CreatedAt: fixedTime,
				Body:      "When sending payloads larger than 10MB, the connection consistently times out after 30 seconds.",
				Comments: []github.Comment{
					{
						ID:        1,
						Author:    "bob",
						CreatedAt: fixedTime.Add(30 * time.Minute),
						Body:      "I can reproduce this. It seems related to the buffer size configuration.",
					},
					{
						ID:        2,
						Author:    "alice",
						CreatedAt: fixedTime.Add(45 * time.Minute),
						Body:      "> I can reproduce this. It seems related to the buffer size configuration.\n\nYes, increasing the buffer size to 20MB resolves it temporarily but is not a proper fix.",
					},
				},
			},
			opts:       Options{},
			goldenFile: "issue.md",
		},
		{
			name: "pull request matches golden file",
			resource: &github.Resource{
				Type:      github.TypePullRequest,
				Title:     "Fix: Increase buffer size for large payloads",
				Author:    "bob",
				URL:       "https://github.com/owner/repo/pull/43",
				State:     "merged",
				Labels:    []string{"fix"},
				CreatedAt: time.Date(2025, 1, 16, 9, 0, 0, 0, time.UTC),
				Body:      "This PR increases the default buffer size from 10MB to 50MB to handle large payloads without timeout.",
				Comments: []github.Comment{
					{
						ID:        100,
						Author:    "alice",
						CreatedAt: time.Date(2025, 1, 16, 9, 30, 0, 0, time.UTC),
						Body:      "LGTM, tested with 40MB payload.",
						IsReview:  false,
					},
					{
						ID:        101,
						Author:    "charlie",
						CreatedAt: time.Date(2025, 1, 16, 10, 0, 0, 0, time.UTC),
						Body:      "Should we also add a configurable max payload size? This feels like a hardcoded value.",
						IsReview:  true,
					},
				},
			},
			opts:       Options{},
			goldenFile: "pr.md",
		},
		{
			name: "discussion matches golden file",
			resource: &github.Resource{
				Type:      github.TypeDiscussion,
				Title:     "How to configure buffer size?",
				Author:    "dave",
				URL:       "https://github.com/owner/repo/discussions/7",
				State:     "open",
				Labels:    nil,
				CreatedAt: time.Date(2025, 1, 17, 14, 0, 0, 0, time.UTC),
				Body:      "I can't find documentation on how to configure the buffer size. Can someone point me in the right direction?",
				Comments: []github.Comment{
					{
						ID:        1,
						Author:    "eve",
						CreatedAt: time.Date(2025, 1, 17, 14, 30, 0, 0, time.UTC),
						Body:      "You can set it via the `BUFFER_SIZE` environment variable.",
					},
				},
				AnswerID: ptr(int64(1)),
			},
			opts:       Options{},
			goldenFile: "discussion.md",
		},
		{
			name: "reactions enabled",
			resource: &github.Resource{
				Type:      github.TypeIssue,
				Title:     "Bug",
				Author:    "alice",
				URL:       "https://github.com/owner/repo/issues/1",
				State:     "open",
				Labels:    nil,
				CreatedAt: fixedTime,
				Body:      "Bug body",
				Reactions: map[string]int{"👍": 5, "🎉": 2},
				Comments: []github.Comment{
					{
						ID:        10,
						Author:    "bob",
						CreatedAt: fixedTime.Add(time.Hour),
						Body:      "Comment",
						Reactions: map[string]int{"❤️": 1},
					},
				},
			},
			opts: Options{EnableReactions: true},
			wantContains: []string{
				"<!-- Reactions: 👍 5  🎉 2 -->",
				"<!-- Reactions: ❤️ 1 -->",
			},
		},
		{
			name: "reactions disabled",
			resource: &github.Resource{
				Type:      github.TypeIssue,
				Title:     "Bug",
				Author:    "alice",
				URL:       "https://github.com/owner/repo/issues/1",
				State:     "open",
				Labels:    nil,
				CreatedAt: fixedTime,
				Body:      "Bug body",
				Reactions: map[string]int{"👍": 5},
				Comments: []github.Comment{
					{
						ID:        10,
						Author:    "bob",
						CreatedAt: fixedTime.Add(time.Hour),
						Body:      "Comment",
						Reactions: map[string]int{"❤️": 1},
					},
				},
			},
			opts: Options{EnableReactions: false},
			wantNotContains: []string{
				"<!-- Reactions:",
			},
		},
		{
			name: "user links enabled",
			resource: &github.Resource{
				Type:      github.TypeIssue,
				Title:     "Bug",
				Author:    "alice",
				URL:       "https://github.com/owner/repo/issues/1",
				State:     "open",
				Labels:    nil,
				CreatedAt: fixedTime,
				Body:      "Body",
				Comments: []github.Comment{
					{
						ID:        10,
						Author:    "bob",
						CreatedAt: fixedTime.Add(time.Hour),
						Body:      "Comment",
					},
				},
			},
			opts: Options{EnableUserLinks: true},
			wantContains: []string{
				"[alice](https://github.com/alice)",
				"[bob](https://github.com/bob)",
			},
		},
		{
			name: "review comment prefix",
			resource: &github.Resource{
				Type:      github.TypePullRequest,
				Title:     "PR",
				Author:    "alice",
				URL:       "https://github.com/owner/repo/pull/1",
				State:     "open",
				Labels:    nil,
				CreatedAt: fixedTime,
				Body:      "PR body",
				Comments: []github.Comment{
					{
						ID:        10,
						Author:    "bob",
						CreatedAt: fixedTime.Add(time.Hour),
						Body:      "Review comment",
						IsReview:  true,
					},
				},
			},
			opts: Options{},
			wantContains: []string{
				"Review Comment by bob",
			},
			wantNotContains: []string{
				"### Comment by bob",
			},
		},
		{
			name: "state capitalization",
			resource: &github.Resource{
				Type:      github.TypeIssue,
				Title:     "Test",
				Author:    "alice",
				URL:       "https://github.com/owner/repo/issues/1",
				State:     "closed",
				Labels:    nil,
				CreatedAt: fixedTime,
				Body:      "Body",
			},
			opts: Options{},
			wantContains: []string{
				"**State:** Closed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert(tt.resource, tt.opts)
			if err != nil {
				t.Fatalf("Convert() error = %v", err)
			}

			// 黄金文件比对
			if tt.goldenFile != "" {
				goldenPath := filepath.Join("testdata", tt.goldenFile)
				want, err := os.ReadFile(goldenPath)
				if err != nil {
					t.Fatalf("read golden file %s: %v", goldenPath, err)
				}
				if got != string(want) {
					t.Errorf("Convert() output doesn't match golden file %s\nGot:\n%s\nWant:\n%s", tt.goldenFile, got, string(want))
				}
				return
			}

			// 包含/不包含断言
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, got)
				}
			}
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(got, notWant) {
					t.Errorf("output should not contain %q\nGot:\n%s", notWant, got)
				}
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }
