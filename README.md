# Team Metrics

A command-line tool that aggregates team activity metrics from multiple sources (GitHub, Jira, Confluence, and Slack) and displays them as horizontal bar charts.

## Features

- **Multi-source data collection**: Fetch data from GitHub, Jira, Confluence, and Slack APIs
- **Parallel processing**: Run multiple providers concurrently for faster data collection
- **Real-time progress**: Stream progress updates as data is being collected
- **Visual output**: Display metrics as horizontal bar charts grouped by user
- **Flexible configuration**: Choose which providers to run via environment variables

## Installation

```bash
git clone https://github.com/iskorotkov/team-metrics
cd team-metrics
go build
```

## Configuration

Create a `.env` file in the project root with the following variables:

```bash
# Required: Comma-separated list of providers to run
MODE=github,jira,confluence,slack

# GitHub configuration
GITHUB_TOKEN=your_github_token
GITHUB_OWNER=your_github_org
GITHUB_REPO=your_repository

# Jira configuration
JIRA_URL=https://your-org.atlassian.net
JIRA_USER=your_email@company.com
JIRA_TOKEN=your_jira_api_token
JIRA_QUERY=project = "YOUR_PROJECT" AND assignee is not EMPTY

# Confluence configuration
CONFLUENCE_URL=https://your-org.atlassian.net
CONFLUENCE_USER=your_email@company.com
CONFLUENCE_TOKEN=your_confluence_api_token
CONFLUENCE_SPACE=YOUR_SPACE_KEY

# Slack configuration
SLACK_TOKEN=your_slack_bot_token
SLACK_QUERY=in:#your-channel after:2024-01-01
```

### API Token Setup

- **GitHub**: Create a personal access token with `repo` and `user` scopes
- **Jira/Confluence**: Generate API tokens from your Atlassian account settings
- **Slack**: Create a bot token with appropriate permissions for message searching

## Usage

```bash
# Run all configured providers
go run .

# Or use the built binary
./team-metrics
```

## Sample Output

```
Fetching reviews for 15 PRs: ............... - done

Open PRs:
Alice Johnson    3    ...
Bob Smith        2    ..
Charlie Brown    1    .

JIRA Issues:
Alice Johnson    5    .....
Charlie Brown    3    ...
Bob Smith        1    .

Slack Messages:
Bob Smith        12   ............
Alice Johnson    8    ........
Charlie Brown    4    ....
```

## Architecture

- **Concurrent Processing**: Each provider runs in its own goroutine using error groups
- **Streaming Output**: Real-time progress updates via `io.Pipe`
- **Modular Design**: Clean separation between data providers, formatters, and transformers
- **Generic Utilities**: Reusable components for data transformation and visualization

## Requirements

- Go 1.24.4 or later
- API access to your organization's GitHub, Jira, Confluence, and/or Slack

## License

MIT License - see [LICENSE](LICENSE) file for details.