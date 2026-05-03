package main

import (
	"context"
	"fmt"
	"os"

	"github.com/steven-ji/issue2md/internal/config"
	"github.com/steven-ji/issue2md/internal/converter"
	"github.com/steven-ji/issue2md/internal/github"
	"github.com/steven-ji/issue2md/internal/parser"
)

const (
	exitSuccess     = 0
	exitGeneralErr  = 1
	exitAPIErr      = 2
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	cfg, err := config.Load(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return exitGeneralErr
	}

	parsed, err := parser.Parse(cfg.URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return exitGeneralErr
	}

	client := github.NewClient(cfg.GitHubToken)
	resource, err := client.FetchResource(context.Background(), parsed.Owner, parsed.Repo, parsed.Type, parsed.Number, cfg.EnableReactions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return exitAPIErr
	}

	markdown, err := converter.Convert(resource, converter.Options{
		EnableReactions: cfg.EnableReactions,
		EnableUserLinks: cfg.EnableUserLinks,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return exitGeneralErr
	}

	if cfg.OutputFile != "" {
		if err := os.WriteFile(cfg.OutputFile, []byte(markdown), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error: write file: %v\n", err)
			return exitGeneralErr
		}
	} else {
		fmt.Print(markdown)
	}

	return exitSuccess
}
