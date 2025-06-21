package main

import (
	"fmt"
	"sort"
	"strings"
)

type ProductWithScrapes struct {
	Product *Product
	Scrapes []SuccessScrape
}

func notify(prices map[*Product][]SuccessScrape, client Client) error {
	var message strings.Builder
	message.WriteString("🛍️ **Cheaper prices found** 🤑\n\n")

	groupedByCategory := make(map[string][]*ProductWithScrapes)
	for product, scrapes := range prices {
		sort.Slice(scrapes, func(i, j int) bool {
			return scrapes[i].Price < scrapes[j].Price
		})

		category := product.Category
		if category == "" {
			category = "Other"
		}

		groupedByCategory[category] = append(groupedByCategory[category], &ProductWithScrapes{
			Product: product,
			Scrapes: scrapes,
		})
	}

	categories := make([]string, 0, len(groupedByCategory))
	for category := range groupedByCategory {
		categories = append(categories, category)
	}
	sort.Strings(categories)

	for _, category := range categories {
		fmt.Fprintf(&message, "**%s**\n\n", category)

		products := groupedByCategory[category]
		sort.Slice(products, func(i, j int) bool {
			return products[i].Product.Name < products[j].Product.Name
		})

		for _, product := range products {
			fmt.Fprintf(&message, "**%s**\n", product.Product.Name)
			fmt.Fprintf(&message, "Base price: £%.2f\n", product.Product.BasePrice)

			cheapest := product.Scrapes[0]
			fmt.Fprintln(&message, cheapest.GetCheapestPriceString(product.Product))

			remaining := product.Scrapes[1:]
			if len(remaining) > 0 {
				fmt.Fprintln(&message, "Other prices:")
				for _, scrape := range remaining {
					fmt.Fprintln(&message, scrape.GetOtherPriceString(product.Product))
				}
			}

			fmt.Fprintln(&message)
		}
	}

	return client.SendMessage(message.String())
}

func GetNotifiablePrices(prices map[*Product][]SuccessScrape, minDiscount float64) map[*Product][]SuccessScrape {
	filteredPrices := make(map[*Product][]SuccessScrape)

	for product, scrapes := range prices {
		var notifiableScrapes []SuccessScrape
		baseThreshold := product.BasePrice * (1 - minDiscount)

		for _, scrape := range scrapes {
			shouldNotify := false

			if scrape.CachedPrice != nil {
				cachedPrice := *scrape.CachedPrice
				lowerThreshold := cachedPrice * (1 - minDiscount)
				upperThreshold := cachedPrice * (1 + minDiscount)

				droppedBelowBaseThreshold := scrape.Price <= baseThreshold && cachedPrice > baseThreshold
				outsideCachedThreshold := scrape.Price <= lowerThreshold || scrape.Price >= upperThreshold

				// Notify if the price dropped below the base threshold or if the price is a good discount and has changed significantly from the cache
				shouldNotify = droppedBelowBaseThreshold || (outsideCachedThreshold && scrape.Price <= baseThreshold)
			} else {
				// No cache => notify if the price is a good discount
				shouldNotify = scrape.Price <= baseThreshold
			}

			if shouldNotify {
				notifiableScrapes = append(notifiableScrapes, scrape)
			}
		}

		if len(notifiableScrapes) > 0 {
			filteredPrices[product] = notifiableScrapes
		}
	}

	return filteredPrices
}
