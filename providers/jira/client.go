package jira

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/iskorotkov/team-metrics/format/progress"
)

type (
	Issue   = jira.Issue
	Comment = jira.Comment
)

func New(url string, user, token string) (*Client, error) {
	c, err := jira.NewClient(&http.Client{
		Timeout: 10 * time.Second,
		Transport: &jira.BasicAuthTransport{
			Username: user,
			Password: token,
		},
	}, url)
	if err != nil {
		return nil, fmt.Errorf("create JIRA client: %w", err)
	}

	return &Client{
		c: c,
	}, nil
}

type Client struct {
	c *jira.Client
}

func (c *Client) Issues(ctx context.Context, query string) ([]Issue, error) {
	issues, _, err := c.c.Issue.SearchWithContext(ctx, query, &jira.SearchOptions{
		MaxResults: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("search issues: %w", err)
	}
	return issues, nil
}

func (c *Client) IssueComments(ctx context.Context, keys ...string) ([]Comment, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	w := progress.ProgressWriter(ctx)
	_, _ = fmt.Fprintf(w, "Fetching comments for %d issues: ", len(keys))

	var comments []Comment
	for _, key := range keys {
		issue, _, err := c.c.Issue.GetWithContext(ctx, key, &jira.GetQueryOptions{})
		if err != nil {
			return nil, fmt.Errorf("get comments for issue %s: %w", key, err)
		}

		for _, comment := range issue.Fields.Comments.Comments {
			comments = append(comments, *comment)
		}

		_, _ = fmt.Fprintf(w, ".")
	}

	_, _ = fmt.Fprintf(w, " - done\n\n")

	return comments, nil
}
