package main

import "context"

type Retailer struct {
	Name    string
	Scraper Scraper
}

func GetRetailers(ctx context.Context) map[string]*Retailer {
	baseScraper := NewBaseScraper(ctx)

	return map[string]*Retailer{
		"boots":         {Name: "Boots", Scraper: NewBootsScraper(baseScraper)},
		"amazon":        {Name: "Amazon", Scraper: NewAmazonScraper(baseScraper)},
		"lookFantastic": {Name: "Look Fantastic", Scraper: NewLookFantasticScraper(baseScraper)},
	}
}
