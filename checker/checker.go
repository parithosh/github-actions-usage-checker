package checker

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github-actions-usage-checker/config"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type Result map[string]map[string][]string

func Run(ctx context.Context, cfg *config.Config) (Result, error) {
	if cfg.GithubToken == "" {
		return nil, fmt.Errorf("github token is required")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GithubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	results := make(Result)

	// Check organization repositories
	for _, org := range cfg.Organizations {
		fmt.Printf("📂 Scanning organization: %s\n", org)

		opt := &github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		}

		var allRepos []*github.Repository
		for {
			repos, resp, err := client.Repositories.ListByOrg(ctx, org, opt)
			if err != nil {
				return nil, fmt.Errorf("failed to list repositories for org %s: %v", org, err)
			}
			allRepos = append(allRepos, repos...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}

		fmt.Printf("Found %d repositories in %s\n", len(allRepos), org)

		for i, repo := range allRepos {
			fmt.Printf("\rScanning repository %d/%d: %s", i+1, len(allRepos), repo.GetName())
			if err := checkRepository(ctx, client, repo.GetOwner().GetLogin(), repo.GetName(), "", cfg.Actions, results); err != nil {
				return nil, err
			}
		}
		fmt.Println() // New line after progress bar
	}

	// Check specific repositories
	for repo, branches := range cfg.Repositories {
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid repository format: %s", repo)
		}
		owner, repoName := parts[0], parts[1]

		if len(branches) == 0 {
			fmt.Printf("Scanning repository: %s/%s\n", owner, repoName)
			if err := checkRepository(ctx, client, owner, repoName, "", cfg.Actions, results); err != nil {
				return nil, err
			}
		} else {
			for _, branch := range branches {
				fmt.Printf("Scanning repository: %s/%s (branch: %s)\n", owner, repoName, branch)
				if err := checkRepository(ctx, client, owner, repoName, branch, cfg.Actions, results); err != nil {
					return nil, err
				}
			}
		}
	}

	// Print summary
	fmt.Println("\n📊 Scan Results:")
	fmt.Println("================")

	totalHits := 0
	for repo, actions := range results {
		if len(actions) > 0 {
			for action, workflows := range actions {
				fmt.Printf("%s: %s (%s)\n", repo, action, strings.Join(workflows, ", "))
				totalHits++
			}
		}
	}

	if totalHits == 0 {
		fmt.Println("\n❌ No matching GitHub Actions found in any repository.")
		fmt.Println("\nActions searched for:")
		for _, action := range cfg.Actions {
			fmt.Printf("  • %s\n", action)
		}
	} else {
		fmt.Printf("\n✨ Found %d matching action(s) across all repositories\n", totalHits)
	}

	return results, nil
}

func checkRepository(ctx context.Context, client *github.Client, owner, repo, branch string, actions []string, results Result) error {
	opts := &github.RepositoryContentGetOptions{}
	if branch != "" {
		opts.Ref = branch
	}

	_, contents, _, err := client.Repositories.GetContents(ctx, owner, repo, ".github/workflows", opts)
	if err != nil {
		return nil // Skip if no workflows directory
	}

	repoKey := fmt.Sprintf("%s/%s", owner, repo)
	results[repoKey] = make(map[string][]string)

	for _, workflowFile := range contents {
		if !strings.HasSuffix(workflowFile.GetName(), ".yml") && !strings.HasSuffix(workflowFile.GetName(), ".yaml") {
			continue
		}

		file, _, _, err := client.Repositories.GetContents(ctx, owner, repo, workflowFile.GetPath(), opts)
		if err != nil {
			continue
		}

		content, contentErr := file.GetContent()
		if contentErr != nil {
			continue
		}

		var decodedContent []byte
		// GitHub API returns raw content for small files, no need to decode
		if !strings.HasPrefix(content, "data:") && !isBase64(content) {
			decodedContent = []byte(content)
		} else {
			// Remove any whitespace and newlines that might interfere with base64 decoding
			content = strings.TrimSpace(content)
			var err error
			decodedContent, err = base64.StdEncoding.DecodeString(content)
			if err != nil {
				// Try URL-safe base64 if standard fails
				decodedContent, err = base64.URLEncoding.DecodeString(content)
				if err != nil {
					continue
				}
			}
		}

		for _, action := range actions {
			if strings.Contains(string(decodedContent), action) {
				results[repoKey][action] = append(results[repoKey][action], workflowFile.GetName())
			}
		}
	}

	return nil
}

func isBase64(s string) bool {
	if len(s)%4 != 0 {
		return false
	}
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=') {
			return false
		}
	}
	return true
}
