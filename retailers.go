package main

type Retailer struct {
	Name    string
	Scraper Scraper
}

func GetRetailers() map[string]*Retailer {
	return map[string]*Retailer{
		"boots":         {Name: "Boots", Scraper: NewBootsScraper()},
		"amazon":        {Name: "Amazon", Scraper: NewAmazonScraper()},
		"lookFantastic": {Name: "Look Fantastic", Scraper: NewLookFantasticScraper()},
	}
}
