package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/iskorotkov/team-metrics/format/bars"
	"github.com/iskorotkov/team-metrics/providers/github"
	"github.com/iskorotkov/team-metrics/transform/maps"
	"github.com/joho/godotenv"
)

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

	fmt.Printf("Open PRs:\n%s\n", bars.Bars(maps.Count(prsByUser)))

	prs, err = gh.ClosedPRs(ctx, os.Getenv("GITHUB_OWNER"), os.Getenv("GITHUB_REPO"))
	if err != nil {
		return fmt.Errorf("get closed PRs: %w", err)
	}

	closedPRsByUser := make(map[string][]*github.PullRequest, len(prs))
	for _, pr := range prs {
		login := pr.GetUser().GetLogin()
		closedPRsByUser[login] = append(closedPRsByUser[login], pr)
	}

	fmt.Printf("Closed PRs:\n%s\n", bars.Bars(maps.Count(closedPRsByUser)))

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

	fmt.Printf("PR Reviews:\n%s\n", bars.Bars(maps.Count(reviewsByUser)))

	comments, err := gh.PRComments(ctx, os.Getenv("GITHUB_OWNER"), os.Getenv("GITHUB_REPO"), prNumbers...)
	if err != nil {
		return fmt.Errorf("get PR comments: %w", err)
	}

	commentsByUser := make(map[string][]*github.PullRequestComment, len(comments))
	for _, comment := range comments {
		login := comment.GetUser().GetLogin()
		commentsByUser[login] = append(commentsByUser[login], comment)
	}

	fmt.Printf("PR Comments:\n%s\n", bars.Bars(maps.Count(commentsByUser)))

	return nil
}
