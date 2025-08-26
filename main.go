package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"slices"
	"strings"

	"github.com/iskorotkov/team-metrics/format/bars"
	"github.com/iskorotkov/team-metrics/format/progress"
	"github.com/iskorotkov/team-metrics/providers/confluence"
	"github.com/iskorotkov/team-metrics/providers/github"
	"github.com/iskorotkov/team-metrics/providers/jira"
	"github.com/iskorotkov/team-metrics/providers/slack"
	"github.com/iskorotkov/team-metrics/transform/maps"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

var providers = map[string]func(context.Context, io.Writer) error{
	"github":     runGithub,
	"jira":       runJira,
	"confluence": runConfluence,
	"slack":      runSlack,
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading .env file: %v\n", err)
	}

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	var readers []*io.PipeReader
	for s := range strings.SplitSeq(os.Getenv("MODE"), ",") {
		r, w := io.Pipe()
		readers = append(readers, r)

		ctx := progress.WithProgressWriter(ctx, w)

		eg.Go(func() error {
			defer func() {
				_ = w.Close()
			}()

			fn, ok := providers[s]
			if !ok {
				return fmt.Errorf("unknown provider: %s", s)
			}
			if err := fn(ctx, w); err != nil {
				return fmt.Errorf("run %s provider: %w", s, err)
			}
			return nil
		})
	}

	var chs []chan []byte
	for _, r := range readers {
		ch := make(chan []byte, 2048)
		chs = append(chs, ch)

		eg.Go(func() error {
			defer close(ch)

			var chunk [64]byte
			for {
				n, err := r.Read(chunk[:])
				if n > 0 {
					ch <- slices.Clone(chunk[:n])
				}
				if errors.Is(err, io.EOF) {
					return nil
				}
				if err != nil {
					return fmt.Errorf("read from pipe: %w", err)
				}
			}
		})
	}

	for _, ch := range chs {
		for chunk := range ch {
			if _, err := os.Stdout.Write(chunk); err != nil {
				return fmt.Errorf("write to stdout: %w", err)
			}
		}
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("wait for goroutines: %w", err)
	}

	return nil
}

func runSlack(ctx context.Context, w io.Writer) error {
	c := slack.New(os.Getenv("SLACK_TOKEN"))

	messages, err := c.Messages(ctx, os.Getenv("SLACK_QUERY"))
	if err != nil {
		return fmt.Errorf("get Slack messages: %w", err)
	}

	messagesByUser := make(map[string][]slack.SearchMessage, len(messages))
	for _, message := range messages {
		name := message.User
		messagesByUser[name] = append(messagesByUser[name], message)
	}

	_, _ = fmt.Fprintf(w, "Slack Messages:\n%s\n", bars.Bars(maps.Count(messagesByUser)))

	return nil
}

func runConfluence(ctx context.Context, w io.Writer) error {
	c, err := confluence.New(os.Getenv("CONFLUENCE_URL"), os.Getenv("CONFLUENCE_USER"), os.Getenv("CONFLUENCE_TOKEN"))
	if err != nil {
		return fmt.Errorf("create Confluence client: %w", err)
	}

	spacePages, err := c.SpacePages(ctx, os.Getenv("CONFLUENCE_SPACE"))
	if err != nil {
		return fmt.Errorf("get Confluence pages: %w", err)
	}

	ids := make([]string, 0, len(spacePages))
	for _, page := range spacePages {
		ids = append(ids, page.ID)
	}

	pages, err := c.Pages(ctx, os.Getenv("CONFLUENCE_SPACE"), ids...)
	if err != nil {
		return fmt.Errorf("get Confluence pages by ids: %w", err)
	}

	pagesByUser := make(map[string][]confluence.Content, len(pages))
	for _, page := range pages {
		name := page.History.CreatedBy.DisplayName
		pagesByUser[name] = append(pagesByUser[name], page)
	}

	_, _ = fmt.Fprintf(w, "Confluence Pages:\n%s\n", bars.Bars(maps.Count(pagesByUser)))

	return nil
}

func runJira(ctx context.Context, w io.Writer) error {
	j, err := jira.New(os.Getenv("JIRA_URL"), os.Getenv("JIRA_USER"), os.Getenv("JIRA_TOKEN"))
	if err != nil {
		return fmt.Errorf("create JIRA client: %w", err)
	}

	issues, err := j.Issues(ctx, os.Getenv("JIRA_QUERY"))
	if err != nil {
		return fmt.Errorf("get JIRA issues: %w", err)
	}

	issuesByUser := make(map[string][]jira.Issue, len(issues))
	for _, issue := range issues {
		if issue.Fields.Assignee == nil {
			continue
		}
		name := issue.Fields.Assignee.DisplayName
		issuesByUser[name] = append(issuesByUser[name], issue)
	}

	_, _ = fmt.Fprintf(w, "JIRA Issues:\n%s\n", bars.Bars(maps.Count(issuesByUser)))

	issueKeys := make([]string, 0, len(issues))
	for _, issue := range issues {
		issueKeys = append(issueKeys, issue.Key)
	}

	comments, err := j.IssueComments(ctx, issueKeys...)
	if err != nil {
		return fmt.Errorf("get JIRA issue comments: %w", err)
	}

	commentsByUser := make(map[string][]jira.Comment, len(issues))
	for _, comment := range comments {
		name := comment.Author.DisplayName
		commentsByUser[name] = append(commentsByUser[name], comment)
	}

	_, _ = fmt.Fprintf(w, "JIRA Comments:\n%s\n", bars.Bars(maps.Count(commentsByUser)))

	return nil
}

func runGithub(ctx context.Context, w io.Writer) error {
	gh := github.New(os.Getenv("GITHUB_TOKEN"))

	prs, err := gh.OpenPRs(ctx, os.Getenv("GITHUB_OWNER"), os.Getenv("GITHUB_REPO"))
	if err != nil {
		return fmt.Errorf("get open PRs: %w", err)
	}

	prsByUser := make(map[string][]*github.PullRequest, len(prs))
	for _, pr := range prs {
		login := pr.GetUser().GetLogin()
		prsByUser[login] = append(prsByUser[login], pr)
	}

	_, _ = fmt.Fprintf(w, "Open PRs:\n%s\n", bars.Bars(maps.Count(prsByUser)))

	prs, err = gh.ClosedPRs(ctx, os.Getenv("GITHUB_OWNER"), os.Getenv("GITHUB_REPO"))
	if err != nil {
		return fmt.Errorf("get closed PRs: %w", err)
	}

	closedPRsByUser := make(map[string][]*github.PullRequest, len(prs))
	for _, pr := range prs {
		login := pr.GetUser().GetLogin()
		closedPRsByUser[login] = append(closedPRsByUser[login], pr)
	}

	_, _ = fmt.Fprintf(w, "Closed PRs:\n%s\n", bars.Bars(maps.Count(closedPRsByUser)))

	prNumbers := make([]int, 0, len(prs))
	for _, pr := range prs {
		prNumbers = append(prNumbers, pr.GetNumber())
	}

	reviews, err := gh.PRReviews(ctx, os.Getenv("GITHUB_OWNER"), os.Getenv("GITHUB_REPO"), prNumbers...)
	if err != nil {
		return fmt.Errorf("get PR reviews: %w", err)
	}

	reviewsByUser := make(map[string][]*github.PullRequestReview, len(reviews))
	for _, review := range reviews {
		login := review.GetUser().GetLogin()
		reviewsByUser[login] = append(reviewsByUser[login], review)
	}

	_, _ = fmt.Fprintf(w, "PR Reviews:\n%s\n", bars.Bars(maps.Count(reviewsByUser)))

	comments, err := gh.PRComments(ctx, os.Getenv("GITHUB_OWNER"), os.Getenv("GITHUB_REPO"), prNumbers...)
	if err != nil {
		return fmt.Errorf("get PR comments: %w", err)
	}

	commentsByUser := make(map[string][]*github.PullRequestComment, len(comments))
	for _, comment := range comments {
		login := comment.GetUser().GetLogin()
		commentsByUser[login] = append(commentsByUser[login], comment)
	}

	_, _ = fmt.Fprintf(w, "PR Comments:\n%s\n", bars.Bars(maps.Count(commentsByUser)))

	return nil
}
