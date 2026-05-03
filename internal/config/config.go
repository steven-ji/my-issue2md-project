package config

import (
	"flag"
	"fmt"
	"os"
)

// Config 是应用运行配置
type Config struct {
	URL             string
	OutputFile      string // 位置参数或 -o 指定的输出路径，空字符串表示 stdout
	EnableReactions bool
	EnableUserLinks bool
	GitHubToken     string // 从环境变量 GITHUB_TOKEN 读取
}

// Load 从命令行参数和环境变量构建 Config
// args[0] 为程序名，与 os.Args 格式一致
func Load(args []string) (*Config, error) {
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)

	outputFile := fs.String("o", "", "output file path")
	enableReactions := fs.Bool("enable-reactions", false, "include reactions in output")
	enableUserLinks := fs.Bool("enable-user-links", false, "render usernames as GitHub profile links")

	if err := fs.Parse(args[1:]); err != nil {
		return nil, fmt.Errorf("parse flags: %w", err)
	}

	// 位置参数：URL [output_file]
	posArgs := fs.Args()
	if len(posArgs) == 0 || posArgs[0] == "" {
		return nil, fmt.Errorf("usage: issue2md [flags] <url> [output_file]")
	}

	url := posArgs[0]
	var positionalOutput string
	if len(posArgs) > 1 {
		positionalOutput = posArgs[1]
	}

	// -o flag 优先于位置参数
	output := positionalOutput
	if *outputFile != "" {
		output = *outputFile
	}

	return &Config{
		URL:             url,
		OutputFile:      output,
		EnableReactions: *enableReactions,
		EnableUserLinks: *enableUserLinks,
		GitHubToken:     os.Getenv("GITHUB_TOKEN"),
	}, nil
}
