package main

import (
	"context"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"regexp"
	"strconv"
	"strings"
)

type Scraper interface {
	ExtractPrice(ctx context.Context, url string) (float64, error)
}

type baseScraper struct {
	selector string
	getText  func(e *colly.HTMLElement) string
}

func newBaseScraper(selector string, getText func(e *colly.HTMLElement) string) *baseScraper {
	return &baseScraper{
		selector: selector,
		getText:  getText,
	}
}

func (b *baseScraper) extractPrice(ctx context.Context, url string) (float64, error) {
	c := colly.NewCollector()
	c.Context = ctx

	extensions.RandomUserAgent(c)

	var price *float64
	var scrapeError error
	var foundElement bool

	// Set up handlers
	c.OnHTML(b.selector, func(e *colly.HTMLElement) {
		foundElement = true
		scrapedPrice, err := parsePrice(b.getText(e))
		if err != nil {
			scrapeError = fmt.Errorf("failed to parse price: %w", err)
			return
		}
		price = scrapedPrice
	})

	c.OnScraped(func(r *colly.Response) {
		if !foundElement && scrapeError == nil {
			scrapeError = fmt.Errorf("no matching elements found for selector %s at %s", b.selector, r.Request.URL)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		scrapeError = fmt.Errorf("request URL: %s failed with response: %v and error: %v\n", r.Request.URL, r, err)
	})

	// Visit URL and wait for completion
	err := c.Visit(url)
	if err != nil {
		return 0, fmt.Errorf("failed to visit %s: %w", url, err)
	}

	c.Wait()

	// Determine result
	if scrapeError != nil {
		return 0, scrapeError
	}

	if price == nil {
		return 0, fmt.Errorf("no price found at %s", url)
	}

	return *price, nil
}

type BootsScraper struct {
	baseScraper *baseScraper
}

func (b *BootsScraper) ExtractPrice(ctx context.Context, url string) (float64, error) {
	return b.baseScraper.extractPrice(ctx, url)
}

func NewBootsScraper() *BootsScraper {
	return &BootsScraper{
		baseScraper: newBaseScraper("div#PDP_productPrice", func(e *colly.HTMLElement) string {
			return e.Text
		}),
	}
}

type AmazonScraper struct {
	baseScraper *baseScraper
}

func (a *AmazonScraper) ExtractPrice(ctx context.Context, url string) (float64, error) {
	return a.baseScraper.extractPrice(ctx, url)
}

func NewAmazonScraper() *AmazonScraper {
	return &AmazonScraper{
		baseScraper: newBaseScraper("span#tp_price_block_total_price_ww", func(e *colly.HTMLElement) string {
			return e.Text
		}),
	}
}

type LookFantasticScraper struct {
	baseScraper *baseScraper
}

func (l *LookFantasticScraper) ExtractPrice(ctx context.Context, url string) (float64, error) {
	return l.baseScraper.extractPrice(ctx, url)
}

func NewLookFantasticScraper() *LookFantasticScraper {
	return &LookFantasticScraper{
		baseScraper: newBaseScraper("div#product-price", func(e *colly.HTMLElement) string {
			return e.ChildText("span")
		}),
	}
}

func parsePrice(price string) (*float64, error) {
	re := regexp.MustCompile(`£(\d+\.\d{1,2})`)
	matches := re.FindStringSubmatch(strings.TrimSpace(price))
	if len(matches) > 1 {
		parsed, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse price %s", matches[1])
		}
		return &parsed, nil
	}

	return nil, fmt.Errorf("no price found")
}
