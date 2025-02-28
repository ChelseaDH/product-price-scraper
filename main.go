package main

import (
	"fmt"
)

func main() {
	retailers := map[string]*Retailer{
		"boots":         {Name: "Boots", Scraper: &BootsScraper{}},
		"amazon":        {Name: "Amazon", Scraper: &AmazonScraper{}},
		"lookFantastic": {Name: "Look Fantastic", Scraper: &LookFantasticScraper{}},
	}

	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	cheaperPrices := make(map[*Product]map[*Retailer]float64)

	for _, product := range getProductsFromConfig(config, retailers) {
		for retailer, link := range product.RetailerLinks {
			price, err := retailer.Scraper.ExtractPrice(link)
			if err != nil {
				fmt.Printf("Error scraping price: %v\n", err)
			}

			if price < product.BasePrice {
				if cheaperPrices[&product] == nil {
					cheaperPrices[&product] = make(map[*Retailer]float64)
				}
				cheaperPrices[&product][retailer] = price
			}
		}
	}

	fmt.Println("Cheaper prices:")
	fmt.Println("---------------")
	for product, prices := range cheaperPrices {
		for retailer, price := range prices {
			fmt.Printf("%s is currently £%.2f at %s (-£%.2f)\n", product.Name, price, retailer.Name, product.BasePrice-price)
		}
	}
}

type Retailer struct {
	Name    string
	Scraper Scraper
}

type Product struct {
	Name          string
	BasePrice     float64
	RetailerLinks map[*Retailer]string
}
