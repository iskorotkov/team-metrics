package confluence

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/iskorotkov/team-metrics/format/progress"
	goconfluence "github.com/virtomize/confluence-go-api"
)

type Content = goconfluence.Content

func New(url, user, token string) (*Client, error) {
	api, err := goconfluence.NewAPI(url, user, token)
	if err != nil {
		return nil, fmt.Errorf("create Confluence client: %w", err)
	}

	api.Client = &http.Client{
		Timeout: 10 * time.Second,
	}

	return &Client{
		c: api,
	}, nil
}

type Client struct {
	c *goconfluence.API
}

func (c *Client) SpacePages(ctx context.Context, space string) ([]Content, error) {
	pages, err := c.c.GetContent(goconfluence.ContentQuery{
		SpaceKey: space,
		Limit:    100,
		OrderBy:  "history.createdDate desc",
	})
	if err != nil {
		return nil, fmt.Errorf("get Confluence pages: %w", err)
	}
	return pages.Results, nil
}

func (c *Client) Pages(ctx context.Context, space string, ids ...string) ([]Content, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	w := progress.ProgressWriter(ctx)
	_, _ = fmt.Fprintf(w, "Fetching pages for %d ids: ", len(ids))

	var pages []Content
	for _, id := range ids {
		page, err := c.c.GetContentByID(id, goconfluence.ContentQuery{
			SpaceKey: space,
		})
		if err != nil {
			return nil, fmt.Errorf("get page %s: %w", id, err)
		}

		pages = append(pages, *page)

		_, _ = fmt.Fprintf(w, ".")
	}

	_, _ = fmt.Fprintf(w, " - done\n\n")

	return pages, nil
}
