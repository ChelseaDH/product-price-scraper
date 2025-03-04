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
	ExtractPrice(url string) (float64, error)
}

type BaseScraper struct {
	collector *colly.Collector
}

func NewBaseScraper(ctx context.Context) *BaseScraper {
	c := colly.NewCollector()
	c.Context = ctx

	extensions.RandomUserAgent(c)

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Request URL: %s failed with response: %v and error: %v\n", r.Request.URL, r, err)
	})

	return &BaseScraper{collector: c}
}

type BootsScraper struct {
	baseScraper *BaseScraper
}

func NewBootsScraper(base *BaseScraper) *BootsScraper {
	return &BootsScraper{baseScraper: base}
}

func (b *BootsScraper) ExtractPrice(url string) (float64, error) {
	return b.baseScraper.ExtractPrice(url, "div#PDP_productPrice", func(e *colly.HTMLElement) string {
		return e.Text
	})
}

type AmazonScraper struct {
	baseScraper *BaseScraper
}

func NewAmazonScraper(base *BaseScraper) *AmazonScraper {
	return &AmazonScraper{baseScraper: base}
}

func (a *AmazonScraper) ExtractPrice(url string) (float64, error) {
	return a.baseScraper.ExtractPrice(url, "span#tp_price_block_total_price_ww", func(e *colly.HTMLElement) string {
		return e.Text
	})
}

type LookFantasticScraper struct {
	baseScraper *BaseScraper
}

func NewLookFantasticScraper(base *BaseScraper) *LookFantasticScraper {
	return &LookFantasticScraper{baseScraper: base}
}

func (l LookFantasticScraper) ExtractPrice(url string) (float64, error) {
	return l.baseScraper.ExtractPrice(url, "div#product-price", func(e *colly.HTMLElement) string {
		return e.ChildText("span.text-gray-900")
	})
}

func (b *BaseScraper) ExtractPrice(url string, selector string, getText func(e *colly.HTMLElement) string) (float64, error) {
	var price float64

	b.collector.OnHTML(selector, func(e *colly.HTMLElement) {
		price = parsePrice(getText(e))
	})

	err := b.collector.Visit(url)
	if err != nil {
		return 0, err
	}

	return price, nil
}

func parsePrice(price string) float64 {
	re := regexp.MustCompile(`£(\d+\.\d{1,2})`)
	matches := re.FindStringSubmatch(strings.TrimSpace(price))
	if len(matches) > 1 {
		parsed, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			return parsed
		}
	}

	return 0
}
