package slack

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/slack-go/slack"
)

type SearchMessage = slack.SearchMessage

func New(token string) *Client {
	return &Client{
		c: slack.New(token,
			slack.OptionConfigToken(token),
			slack.OptionHTTPClient(&http.Client{
				Timeout: 10 * time.Second,
			}),
		),
	}
}

type Client struct {
	c *slack.Client
}

func (c *Client) Messages(ctx context.Context, query string) ([]SearchMessage, error) {
	messages, err := c.c.SearchMessagesContext(ctx, query, slack.SearchParameters{
		Sort:          "timestamp",
		SortDirection: "desc",
		Count:         100,
	})
	if err != nil {
		return nil, fmt.Errorf("search Slack messages: %w", err)
	}
	return messages.Matches, nil
}
