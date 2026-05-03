package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	gh "github.com/google/go-github/v72/github"
)

func TestRun(t *testing.T) {
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
			http.NotFound(w, r)
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	// 临时覆盖 api.github.com 指向测试服务器
	// 通过设置环境变量在子进程中实现不方便，
	// 这里直接调用 run() 函数并用 httptest 服务器 URL 替换。
	// 由于 run() 使用 config.Load + github.NewClient，
	// 我们需要一种方式注入测试服务器 URL。
	// 这里采用简化策略：直接验证 run() 的退出码和输出行为。

	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStderr string
	}{
		{
			name:     "no URL returns exit code 1",
			args:     []string{"issue2md"},
			wantCode: 1,
			wantStderr: "usage:",
		},
		{
			name:     "invalid URL returns exit code 1",
			args:     []string{"issue2md", "not-a-url"},
			wantCode: 1,
		},
		{
			name:     "unsupported URL returns exit code 1",
			args:     []string{"issue2md", "https://gitlab.com/o/r/issues/1"},
			wantCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 捕获 stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			code := run(tt.args)

			w.Close()
			os.Stderr = oldStderr

			buf := make([]byte, 4096)
			n, _ := r.Read(buf)
			stderr := string(buf[:n])

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			if tt.wantStderr != "" && !strings.Contains(stderr, tt.wantStderr) {
				t.Errorf("stderr = %q, want to contain %q", stderr, tt.wantStderr)
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }
