package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github-actions-usage-checker/checker"
	"github-actions-usage-checker/config"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	results, err := checker.Run(context.Background(), cfg)
	if err != nil {
		log.Fatalf("Failed to run checker: %v", err)
	}

	for repo, actions := range results {
		if len(actions) > 0 {
			fmt.Printf("\nRepository %s uses the following actions:\n", repo)
			for action, workflows := range actions {
				fmt.Printf("  - %s in workflows: %v\n", action, workflows)
			}
		}
	}
}
