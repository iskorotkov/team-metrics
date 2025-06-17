package github

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-github/v72/github"
)

type (
	PullRequest        = github.PullRequest
	PullRequestReview  = github.PullRequestReview
	PullRequestComment = github.PullRequestComment
)

func New(token string) *Client {
	return &Client{
		c: github.NewClient(&http.Client{
			Timeout: 10 * time.Second,
			Transport: &github.BasicAuthTransport{
				Username: "",
				Password: token,
			},
		}),
	}
}

type Client struct {
	c *github.Client
}

func (c *Client) OpenPRs(ctx context.Context, owner, repo string) ([]*PullRequest, error) {
	prs, _, err := c.c.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get open PRs: %w", err)
	}
	return prs, nil
}

func (c *Client) ClosedPRs(ctx context.Context, owner, repo string) ([]*PullRequest, error) {
	prs, _, err := c.c.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
		State: "closed",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get closed PRs: %w", err)
	}
	return prs, nil
}

func (c *Client) PRReviews(ctx context.Context, owner, repo string, numbers ...int) ([]*PullRequestReview, error) {
	if len(numbers) == 0 {
		return nil, nil
	}

	fmt.Printf("Fetching reviews for %d PRs: ", len(numbers))

	var allReviews []*PullRequestReview
	for _, number := range numbers {
		reviews, _, err := c.c.PullRequests.ListReviews(ctx, owner, repo, number, &github.ListOptions{
			PerPage: 100,
		})
		if err != nil {
			return nil, fmt.Errorf("get reviews for PR #%d: %w", number, err)
		}

		allReviews = append(allReviews, reviews...)
		fmt.Print(".")
	}

	fmt.Print(" - done\n\n")

	return allReviews, nil
}

func (c *Client) PRComments(ctx context.Context, owner, repo string, numbers ...int) ([]*PullRequestComment, error) {
	if len(numbers) == 0 {
		return nil, nil
	}

	fmt.Printf("Fetching comments for %d PRs: ", len(numbers))

	var allComments []*PullRequestComment
	for _, number := range numbers {
		comments, _, err := c.c.PullRequests.ListComments(ctx, owner, repo, number, &github.PullRequestListCommentsOptions{
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("get comments for PR #%d: %w", number, err)
		}

		allComments = append(allComments, comments...)
		fmt.Print(".")
	}

	fmt.Print(" - done\n\n")

	return allComments, nil
}
