# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

This is a Go project using Go modules. Use standard Go toolchain commands:

- `go run .` - Run the application
- `go build` - Build the binary
- `go mod tidy` - Clean up dependencies
- `go mod download` - Download dependencies
- `go test ./...` - Run all tests
- `go fmt ./...` - Format code

## Environment Variables

The application requires several environment variables to be configured:

- `MODE` - Comma-separated list of providers to run (github,jira,confluence,slack)
- **GitHub**: `GITHUB_TOKEN`, `GITHUB_OWNER`, `GITHUB_REPO`
- **Jira**: `JIRA_URL`, `JIRA_USER`, `JIRA_TOKEN`, `JIRA_QUERY`
- **Confluence**: `CONFLUENCE_URL`, `CONFLUENCE_USER`, `CONFLUENCE_TOKEN`, `CONFLUENCE_SPACE`
- **Slack**: `SLACK_TOKEN`, `SLACK_QUERY`

Environment variables are loaded from a `.env` file using godotenv.

## Architecture

This is a team metrics aggregation tool that collects data from multiple sources (GitHub, Jira, Confluence, Slack) and displays activity metrics in a bar chart format.

### Core Architecture

**Parallel Provider Pattern**: The application runs multiple providers concurrently using errgroups. Each provider is specified in the `MODE` environment variable and runs in its own goroutine.

**Streaming Output**: Uses `io.Pipe` to stream progress updates from each provider to stdout in real-time. Each provider writes to its own pipe, and the main goroutine reads from all pipes concurrently.

**Provider Interface**: Each provider implements a function signature `func(context.Context, io.Writer) error` and is registered in the `providers` map in `main.go:24-29`.

### Package Structure

- **`providers/`** - Data collection clients for external services
  - `github/` - GitHub API client (PRs, reviews, comments)
  - `jira/` - Jira API client (issues, comments)
  - `confluence/` - Confluence API client (pages)
  - `slack/` - Slack API client (messages)

- **`format/`** - Output formatting utilities
  - `bars/` - Creates horizontal bar charts from count data
  - `progress/` - Context-based progress writer for streaming updates

- **`transform/`** - Data transformation utilities
  - `maps/` - Generic map transformation functions (count aggregation)

### Provider Implementation Pattern

Each provider follows the same pattern:
1. Create authenticated client using environment variables
2. Fetch data from external API
3. Group data by user (DisplayName, Login, etc.)
4. Transform to counts using `maps.Count()`
5. Format as bar chart using `bars.Bars()`
6. Write progress updates to the provided writer

### Key Dependencies

- `github.com/google/go-github/v72` - GitHub API client
- `github.com/andygrunwald/go-jira` - Jira API client  
- `github.com/virtomize/confluence-go-api` - Confluence API client
- `github.com/slack-go/slack` - Slack API client
- `golang.org/x/sync/errgroup` - Concurrent goroutine management
- `github.com/joho/godotenv` - Environment variable loading

## Data Flow

1. **Initialization**: Load environment variables, parse MODE string
2. **Parallel Execution**: Start one goroutine per provider with dedicated pipe
3. **Data Collection**: Each provider fetches and processes data independently
4. **Real-time Output**: Progress updates stream to stdout as they're generated
5. **Aggregation**: Results are displayed as bar charts showing user activity counts

The application focuses on measuring individual contributor activity across different platforms, aggregating by user identity to show team productivity metrics.