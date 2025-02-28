package main

import (
	"fmt"
	"sort"
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
		keys := sortRetailersByPrice(prices)

		fmt.Printf("-- %s --\n", product.Name)
		fmt.Printf("Base price: £%.2f\n", product.BasePrice)
		fmt.Printf("Best price: £%.2f at %s %s\n", prices[keys[0]], keys[0].Name, getDiscountString(*product, prices[keys[0]]))

		remaining := keys[1:]
		if len(remaining) > 0 {
			fmt.Println("Other prices:")
			for _, key := range remaining {
				fmt.Printf("- £%.2f at %s %s\n", prices[key], key.Name, getDiscountString(*product, prices[key]))
			}
		}

		fmt.Println()
	}
}

func sortRetailersByPrice(prices map[*Retailer]float64) []*Retailer {
	keys := make([]*Retailer, 0, len(prices))
	for retailer := range prices {
		keys = append(keys, retailer)
	}

	sort.Slice(keys, func(i, j int) bool {
		return prices[keys[i]] < prices[keys[j]]
	})

	return keys
}

func getDiscountString(product Product, price float64) string {
	discount := product.BasePrice - price
	percentage := discount / product.BasePrice * 100
	return fmt.Sprintf("(-£%.2f | %.2f%% off)", discount, percentage)
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
