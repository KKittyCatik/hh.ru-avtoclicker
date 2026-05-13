package hh

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"sync"

	"golang.org/x/sync/errgroup"
)

type VacancySearchParams struct {
	Text     string
	Area     string
	Salary   int
	Schedule string
}

func (c *Client) SearchVacancies(ctx context.Context, params VacancySearchParams) ([]Vacancy, error) {
	values := url.Values{}
	if params.Text != "" {
		values.Set("text", params.Text)
	}
	if params.Area != "" {
		values.Set("area", params.Area)
	}
	if params.Salary > 0 {
		values.Set("salary", strconv.Itoa(params.Salary))
	}
	if params.Schedule != "" && params.Schedule != "any" {
		values.Set("schedule", params.Schedule)
	}
	values.Set("per_page", "100")

	var all []Vacancy
	page := 0
	pages := 1
	for page < pages {
		values.Set("page", strconv.Itoa(page))
		var resp VacancySearchResponse
		if err := c.DoJSON(ctx, "GET", "/vacancies?"+values.Encode(), nil, &resp); err != nil {
			return nil, fmt.Errorf("search vacancies page %d: %w", page, err)
		}
		all = append(all, resp.Items...)
		pages = resp.Pages
		page++
	}

	return deduplicateVacancies(all), nil
}

func (c *Client) SearchVacanciesByURL(ctx context.Context, rawURL string) ([]Vacancy, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse search URL: %w", err)
	}
	query := parsed.Query()
	query.Set("per_page", "100")
	parsed.RawQuery = query.Encode()

	var all []Vacancy
	page := 0
	pages := 1
	for page < pages {
		query.Set("page", strconv.Itoa(page))
		parsed.RawQuery = query.Encode()
		path := parsed.Path + "?" + parsed.RawQuery
		if parsed.Path == "" {
			path = "/vacancies?" + parsed.RawQuery
		}
		var resp VacancySearchResponse
		if err := c.DoJSON(ctx, "GET", path, nil, &resp); err != nil {
			return nil, fmt.Errorf("search vacancies by URL page %d: %w", page, err)
		}
		all = append(all, resp.Items...)
		pages = resp.Pages
		page++
	}
	return deduplicateVacancies(all), nil
}

func (c *Client) CollectVacanciesByURLs(ctx context.Context, searchURLs []string, concurrency int) ([]Vacancy, error) {
	if concurrency <= 0 {
		concurrency = 1
	}
	g, gctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, concurrency)
	combined := make([]Vacancy, 0)
	var mu sync.Mutex

	for _, u := range searchURLs {
		u := u
		g.Go(func() error {
			select {
			case sem <- struct{}{}:
			case <-gctx.Done():
				return fmt.Errorf("wait semaphore: %w", gctx.Err())
			}
			defer func() { <-sem }()

			items, err := c.SearchVacanciesByURL(gctx, u)
			if err != nil {
				return fmt.Errorf("collect vacancies from %s: %w", u, err)
			}
			mu.Lock()
			combined = append(combined, items...)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("collect vacancies by URLs: %w", err)
	}
	return deduplicateVacancies(combined), nil
}

func deduplicateVacancies(items []Vacancy) []Vacancy {
	seen := make(map[string]struct{}, len(items))
	result := make([]Vacancy, 0, len(items))
	for _, v := range items {
		if _, ok := seen[v.ID]; ok {
			continue
		}
		seen[v.ID] = struct{}{}
		result = append(result, v)
	}
	return result
}
