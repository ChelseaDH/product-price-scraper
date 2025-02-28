package main

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
	"regexp"
	"strconv"
	"strings"
)

type Scraper interface {
	ExtractPrice(url string) (float64, error)
}

type BootsScraper struct{}

func (b *BootsScraper) ExtractPrice(url string) (float64, error) {
	return scrapePriceViaColly(url, "div#PDP_productPrice", func(e *colly.HTMLElement) string {
		return e.Text
	})
}

type AmazonScraper struct{}

func (a *AmazonScraper) ExtractPrice(url string) (float64, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var text string

	// Run Chrome headless and scrape price
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Text("span.a-offscreen", &text),
	)
	if err != nil {
		return 0, err
	}

	return parsePrice(text), nil
}

type LookFantasticScraper struct{}

func (l LookFantasticScraper) ExtractPrice(url string) (float64, error) {
	return scrapePriceViaColly(url, "div#product-price", func(e *colly.HTMLElement) string {
		return e.ChildText("span.text-gray-900")
	})
}

func scrapePriceViaColly(url string, selector string, getText func(e *colly.HTMLElement) string) (float64, error) {
	c := colly.NewCollector()

	var price float64

	c.OnHTML(selector, func(e *colly.HTMLElement) {
		price = parsePrice(getText(e))
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Request URL: %s failed with response: %v and error: %v\n", r.Request.URL, r, err)
	})

	err := c.Visit(url)
	if err != nil {
		return 0, err
	}

	return price, nil
}

func parsePrice(price string) float64 {
	re := regexp.MustCompile(`[\d.,]+`)
	matches := re.FindString(strings.TrimSpace(price))
	if matches != "" {
		parsed, err := strconv.ParseFloat(matches, 64)
		if err == nil {
			return parsed
		}
	}

	return 0
}
