package main

import (
	"bytes"
	"context"
	"fmt"
	"sort"
)

type Retailer struct {
	Name    string
	Scraper Scraper
}

type FailedScrape struct {
	Product  *Product
	Retailer *Retailer
	Error    error
}

type SuccessScrape struct {
	Retailer *Retailer
	Price    float64
	Url      string
}

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

	client, err := getClient(config, context.Background())
	products := getProducts(config, retailers)

	err = findPricesAndNotify(products, client)
	if err != nil {
		fmt.Println("Error finding prices and notifying:", err)
	}

	err = client.Stop()
	if err != nil {
		fmt.Println("Error stopping client:", err)
	}
}

func findPricesAndNotify(products []Product, client Client) error {
	cheaperPrices := make(map[*Product][]SuccessScrape)
	var failures []FailedScrape

	for i, product := range products {
		var successScrapes []SuccessScrape

		for retailer, link := range product.RetailerLinks {
			price, err := retailer.Scraper.ExtractPrice(link)
			if err != nil {
				failures = append(failures, FailedScrape{Product: &products[i], Retailer: retailer, Error: err})
				continue
			}

			if price < product.BasePrice {
				successScrapes = append(successScrapes, SuccessScrape{Retailer: retailer, Price: price, Url: link})
			}
		}

		if len(successScrapes) > 0 {
			cheaperPrices[&products[i]] = successScrapes
		}
	}

	if len(cheaperPrices) == 0 {
		return nil
	}

	message := bytes.Buffer{}
	fmt.Fprintf(&message, "🛍️ **Cheaper prices found** 🤑\n\n")

	for product, scrapes := range cheaperPrices {
		sort.Slice(scrapes, func(i, j int) bool {
			return scrapes[i].Price < scrapes[j].Price
		})

		fmt.Fprintf(&message, "**%s**\n", product.Name)
		fmt.Fprintf(&message, "Base price: £%.2f\n", product.BasePrice)
		fmt.Fprintf(&message, "Best price: **£%.2f** at [%s](%s) %s\n", scrapes[0].Price, scrapes[0].Retailer.Name, scrapes[0].Url, getDiscountString(*product, scrapes[0].Price))

		remaining := scrapes[1:]
		if len(remaining) > 0 {
			fmt.Fprintln(&message, "Other prices:")
			for _, scrape := range remaining {
				fmt.Fprintf(&message, "- £%.2f at [%s](%s) %s\n", scrape.Price, scrape.Retailer.Name, scrape.Url, getDiscountString(*product, scrape.Price))
			}
		}

		fmt.Fprintln(&message)
	}

	err := client.SendMessage(message.String())
	return err
}

func getDiscountString(product Product, price float64) string {
	discount := product.BasePrice - price
	percentage := discount / product.BasePrice * 100
	return fmt.Sprintf("(-£%.2f | %.2f%% off)", discount, percentage)
}
