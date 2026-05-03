package parser

import (
	"testing"

	"github.com/steven-ji/issue2md/internal/github"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		rawURL     string
		wantOwner  string
		wantRepo   string
		wantType   github.ResourceType
		wantNumber int
		wantErr    bool
	}{
		{
			name:       "valid issue URL",
			rawURL:     "https://github.com/owner/repo/issues/42",
			wantOwner:  "owner",
			wantRepo:   "repo",
			wantType:   github.TypeIssue,
			wantNumber: 42,
			wantErr:    false,
		},
		{
			name:       "valid pull request URL",
			rawURL:     "https://github.com/owner/repo/pull/7",
			wantOwner:  "owner",
			wantRepo:   "repo",
			wantType:   github.TypePullRequest,
			wantNumber: 7,
			wantErr:    false,
		},
		{
			name:       "valid discussion URL",
			rawURL:     "https://github.com/owner/repo/discussions/3",
			wantOwner:  "owner",
			wantRepo:   "repo",
			wantType:   github.TypeDiscussion,
			wantNumber: 3,
			wantErr:    false,
		},
		{
			name:       "issue URL with trailing slash",
			rawURL:     "https://github.com/owner/repo/issues/42/",
			wantOwner:  "owner",
			wantRepo:   "repo",
			wantType:   github.TypeIssue,
			wantNumber: 42,
			wantErr:    false,
		},
		{
			name:       "issue URL with query parameters",
			rawURL:     "https://github.com/owner/repo/issues/42?foo=bar",
			wantOwner:  "owner",
			wantRepo:   "repo",
			wantType:   github.TypeIssue,
			wantNumber: 42,
			wantErr:    false,
		},
		{
			name:      "empty URL",
			rawURL:    "",
			wantErr:   true,
		},
		{
			name:      "malformed URL",
			rawURL:    "://not-a-url",
			wantErr:   true,
		},
		{
			name:      "non-GitHub domain",
			rawURL:    "https://gitlab.com/owner/repo/issues/42",
			wantErr:   true,
		},
		{
			name:      "repo homepage (unsupported)",
			rawURL:    "https://github.com/owner/repo",
			wantErr:   true,
		},
		{
			name:      "repo issues list (unsupported)",
			rawURL:    "https://github.com/owner/repo/issues",
			wantErr:   true,
		},
		{
			name:      "non-numeric issue number",
			rawURL:    "https://github.com/owner/repo/issues/abc",
			wantErr:   true,
		},
		{
			name:      "unsupported path segment (releases)",
			rawURL:    "https://github.com/owner/repo/releases/1",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.rawURL)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got nil", tt.rawURL)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tt.rawURL, err)
			}
			if got.Owner != tt.wantOwner {
				t.Errorf("Owner = %q, want %q", got.Owner, tt.wantOwner)
			}
			if got.Repo != tt.wantRepo {
				t.Errorf("Repo = %q, want %q", got.Repo, tt.wantRepo)
			}
			if got.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantType)
			}
			if got.Number != tt.wantNumber {
				t.Errorf("Number = %d, want %d", got.Number, tt.wantNumber)
			}
		})
	}
}
