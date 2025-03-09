package main

import (
	"context"
	"fmt"
)

type Product struct {
	Name          string
	BasePrice     float64
	RetailerLinks map[*Retailer]string
}
type Products []Product

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

func GetProducts(config Config, retailers map[string]*Retailer) Products {
	var products []Product

	for _, p := range config.Products {
		product := Product{
			Name:          p.Name,
			BasePrice:     p.BasePrice,
			RetailerLinks: make(map[*Retailer]string),
		}

		for retailerName, link := range p.Links {
			retailer := retailers[retailerName]
			product.RetailerLinks[retailer] = link
		}

		products = append(products, product)
	}

	return products
}

func (p Products) GetPrices(ctx context.Context) (map[*Product][]SuccessScrape, []FailedScrape) {
	prices := make(map[*Product][]SuccessScrape)
	var failures []FailedScrape

	for i, product := range p {
		var successScrapes []SuccessScrape

		for retailer, link := range product.RetailerLinks {
			price, err := retailer.Scraper.ExtractPrice(ctx, link)
			if err != nil {
				failures = append(failures, FailedScrape{Product: &p[i], Retailer: retailer, Error: err})
				continue
			}

			successScrapes = append(successScrapes, SuccessScrape{Retailer: retailer, Price: price, Url: link})
		}

		prices[&p[i]] = successScrapes
	}

	return prices, failures
}

func (p Products) FindPricesAndNotify(ctx context.Context, client Client, cache *Cache, minDiscount float64) error {
	prices, _ := p.GetPrices(ctx)
	cachedPrices, err := cache.GetScrapes()
	if err != nil {
		return fmt.Errorf("error getting cached prices: %v", err)
	}

	if shouldNotify(prices, cachedPrices, minDiscount) {
		fmt.Println("New prices found, notifying")
		err = notify(prices, client)
		if err != nil {
			return fmt.Errorf("error notifying products: %v", err)
		}
	} else {
		fmt.Println("No new prices found")
	}

	return cache.SetScrapes(prices)
}

func (p Product) getDiscountString(price float64) string {
	discount := p.BasePrice - price
	percentage := discount / p.BasePrice * 100
	return fmt.Sprintf("(-£%.2f | %.2f%% off)", discount, percentage)
}
