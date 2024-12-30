# github-actions-usage-checker

This project can be used to check if a configured github action is used in a repository or organization. The output is a list of repositories that are using the specified action.

The tool is written in Go and uses the github API to first fetch the workflows of the repositories and then check if the specified actions are used in said workflows. If no branch is specified, the tool will check the default branch of the repository. A github token is required to use the tool in order to avoid rate limiting.

## Installation

```bash
go install github.com/parithosh/github-actions-usage-checker@latest
```

Or build from source:
```bash
git clone https://github.com/parithosh/github-actions-usage-checker.git
cd github-actions-usage-checker
go build
```

## Configuration

1. Copy the example config file:
```bash
cp config.yaml.example config.yaml
```

2. Edit `config.yaml` with your settings:
```yaml
# GitHub token (required)
github_token: "your-token-here"

# Actions to search for
actions:
  - "actions/checkout@v4"
  - "actions/setup-node@v4"

# Organizations to scan (all repositories)
organizations:
  - "your-org"

# Specific repositories and branches
repositories:
  owner/repo:
    - "main"
    - "develop"
  another/repo: []  # empty array = default branch
```

The GitHub token can also be provided via the `GITHUB_TOKEN` environment variable.

## Usage

```bash
./github-actions-usage-checker --config ./config.yaml
```

## Output Example

```
📂 Scanning organization: your-org
Found 25 repositories in your-org

📊 Scan Results:
================
your-org/repo1: actions/checkout@v4 (ci.yml, release.yml)
your-org/repo1: actions/setup-node@v4 (ci.yml)
your-org/repo2: actions/checkout@v4 (test.yml)

✨ Found 3 matching action(s) across all repositories
```

If no matches are found:
```
❌ No matching GitHub Actions found in any repository.

Actions searched for:
  • actions/checkout@v4
  • actions/setup-node@v4
```

