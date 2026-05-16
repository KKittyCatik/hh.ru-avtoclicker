package browser

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"

	"hh-autoresponder/internal/hh"
)

var vacancyIDPattern = regexp.MustCompile(`/vacancy/(\d+)`)

func (ac *AccountContext) SearchVacancies(ctx context.Context, searchURL string) ([]hh.Vacancy, error) {
	tctx, cancel := withTimeout(ctx, 2*time.Minute)
	defer cancel()

	if _, err := ac.Page.Goto(searchURL); err != nil {
		return nil, fmt.Errorf("open search url: %w", err)
	}
	if _, err := waitForSelectorWithDebug(ac.Page, `[data-qa="vacancy-serp__results"]`, 20*time.Second); err != nil {
		return nil, err
	}

	all := make([]hh.Vacancy, 0)
	seen := map[string]struct{}{}
	for pageNum := 0; pageNum < 5; pageNum++ {
		cards, err := ac.Page.QuerySelectorAll(`[data-qa="vacancy-serp__vacancy"]`)
		if err != nil {
			return nil, fmt.Errorf("query vacancy cards: %w", err)
		}
		for _, card := range cards {
			vac, err := parseVacancyCard(card)
			if err != nil || vac.ID == "" {
				continue
			}
			if _, ok := seen[vac.ID]; ok {
				continue
			}
			seen[vac.ID] = struct{}{}
			all = append(all, *vac)
		}

		next, err := ac.Page.QuerySelector(`[data-qa="pager-next"]`)
		if err != nil {
			return nil, fmt.Errorf("query next pager: %w", err)
		}
		if next == nil {
			break
		}
		_, _ = ac.Page.Evaluate("(el) => el.scrollIntoView({block: 'center'})", next)
		if err := randomPause(tctx, 2*time.Second, 4*time.Second); err != nil {
			return nil, err
		}
		if err := ac.Page.Click(`[data-qa="pager-next"]`); err != nil {
			return nil, fmt.Errorf("click next page: %w", err)
		}
		if _, err := waitForSelectorWithDebug(ac.Page, `[data-qa="vacancy-serp__results"]`, 20*time.Second); err != nil {
			return nil, err
		}
	}
	return all, nil
}

func (ac *AccountContext) GetVacancyDetails(ctx context.Context, vacancyID string) (*hh.Vacancy, error) {
	tctx, cancel := withTimeout(ctx, defaultOpTimeout)
	defer cancel()

	url := fmt.Sprintf("https://hh.ru/vacancy/%s", vacancyID)
	if _, err := ac.Page.Goto(url); err != nil {
		return nil, fmt.Errorf("open vacancy page: %w", err)
	}
	if _, err := waitForSelectorWithDebug(ac.Page, `[data-qa="vacancy-title"]`, 20*time.Second); err != nil {
		return nil, err
	}

	vac := &hh.Vacancy{ID: vacancyID, URL: url}
	vac.Name = textBySelector(ac.Page, `[data-qa="vacancy-title"]`)
	vac.Description = textBySelector(ac.Page, `[data-qa="vacancy-description"]`)
	vac.Employer.Name = textBySelector(ac.Page, `[data-qa="vacancy-company-name"]`)
	_ = randomPause(tctx, 500*time.Millisecond, 2*time.Second)
	return vac, nil
}

func parseVacancyCard(card playwright.ElementHandle) (*hh.Vacancy, error) {
	titleNode, err := card.QuerySelector(`[data-qa="serp-item__title"]`)
	if err != nil || titleNode == nil {
		return nil, fmt.Errorf("title not found")
	}
	href, _ := titleNode.GetAttribute("href")
	id := ""
	if match := vacancyIDPattern.FindStringSubmatch(href); len(match) == 2 {
		id = match[1]
	}
	title, _ := titleNode.TextContent()

	employerNode, _ := card.QuerySelector(`[data-qa="vacancy-serp__vacancy-employer"]`)
	employer := textOrEmpty(employerNode)
	compensationNode, _ := card.QuerySelector(`[data-qa="vacancy-serp__vacancy-compensation"]`)
	compensation := textOrEmpty(compensationNode)
	formatNode, _ := card.QuerySelector(`[data-qa="vacancy-serp__vacancy-work-format"]`)
	format := textOrEmpty(formatNode)

	v := &hh.Vacancy{
		ID:          id,
		Name:        strings.TrimSpace(title),
		Description: compensation,
		Employer:    hh.Employer{Name: employer},
		Schedule:    format,
		URL:         href,
	}
	return v, nil
}

func textBySelector(page playwright.Page, selector string) string {
	node, err := page.QuerySelector(selector)
	if err != nil || node == nil {
		return ""
	}
	text, err := node.TextContent()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(text)
}

func textOrEmpty(node playwright.ElementHandle) string {
	if node == nil {
		return ""
	}
	text, _ := node.TextContent()
	return strings.TrimSpace(text)
}
