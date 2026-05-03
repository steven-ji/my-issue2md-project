package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		envToken        string
		wantURL         string
		wantOutputFile  string
		wantReactions   bool
		wantUserLinks   bool
		wantToken       string
		wantErr         bool
	}{
		{
			name:           "minimal: only URL",
			args:           []string{"issue2md", "https://github.com/o/r/issues/1"},
			wantURL:        "https://github.com/o/r/issues/1",
			wantOutputFile: "",
			wantReactions:  false,
			wantUserLinks:  false,
			wantErr:        false,
		},
		{
			name:           "URL with output file positional arg",
			args:           []string{"issue2md", "https://github.com/o/r/issues/1", "out.md"},
			wantURL:        "https://github.com/o/r/issues/1",
			wantOutputFile: "out.md",
			wantErr:        false,
		},
		{
			name:           "URL with -o flag",
			args:           []string{"issue2md", "-o", "flag.md", "https://github.com/o/r/issues/1"},
			wantURL:        "https://github.com/o/r/issues/1",
			wantOutputFile: "flag.md",
			wantErr:        false,
		},
		{
			name:           "-o flag overrides positional output_file",
			args:           []string{"issue2md", "-o", "flag.md", "https://github.com/o/r/issues/1", "positional.md"},
			wantURL:        "https://github.com/o/r/issues/1",
			wantOutputFile: "flag.md",
			wantErr:        false,
		},
		{
			name:           "-enable-reactions flag",
			args:           []string{"issue2md", "-enable-reactions", "https://github.com/o/r/issues/1"},
			wantURL:        "https://github.com/o/r/issues/1",
			wantReactions:  true,
			wantErr:        false,
		},
		{
			name:           "-enable-user-links flag",
			args:           []string{"issue2md", "-enable-user-links", "https://github.com/o/r/issues/1"},
			wantURL:        "https://github.com/o/r/issues/1",
			wantUserLinks:  true,
			wantErr:        false,
		},
		{
			name:           "GITHUB_TOKEN from env",
			args:           []string{"issue2md", "https://github.com/o/r/issues/1"},
			envToken:       "ghp_test123",
			wantURL:        "https://github.com/o/r/issues/1",
			wantToken:      "ghp_test123",
			wantErr:        false,
		},
		{
			name:    "no URL provided",
			args:    []string{"issue2md"},
			wantErr: true,
		},
		{
			name:    "empty URL",
			args:    []string{"issue2md", ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置/清理环境变量
			if tt.envToken != "" {
				os.Setenv("GITHUB_TOKEN", tt.envToken)
				defer os.Unsetenv("GITHUB_TOKEN")
			} else {
				origToken := os.Getenv("GITHUB_TOKEN")
				os.Unsetenv("GITHUB_TOKEN")
				defer func() {
					if origToken != "" {
						os.Setenv("GITHUB_TOKEN", origToken)
					}
				}()
			}

			cfg, err := Load(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", cfg.URL, tt.wantURL)
			}
			if cfg.OutputFile != tt.wantOutputFile {
				t.Errorf("OutputFile = %q, want %q", cfg.OutputFile, tt.wantOutputFile)
			}
			if cfg.EnableReactions != tt.wantReactions {
				t.Errorf("EnableReactions = %v, want %v", cfg.EnableReactions, tt.wantReactions)
			}
			if cfg.EnableUserLinks != tt.wantUserLinks {
				t.Errorf("EnableUserLinks = %v, want %v", cfg.EnableUserLinks, tt.wantUserLinks)
			}
			if tt.wantToken != "" && cfg.GitHubToken != tt.wantToken {
				t.Errorf("GitHubToken = %q, want %q", cfg.GitHubToken, tt.wantToken)
			}
		})
	}
}
