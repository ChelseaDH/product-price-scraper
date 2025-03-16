package main

import (
	"bytes"
	"fmt"
	"sort"
)

func notify(prices map[*Product][]SuccessScrape, client Client) error {
	message := bytes.Buffer{}
	fmt.Fprintf(&message, "🛍️ **Cheaper prices found** 🤑\n\n")

	for product, scrapes := range prices {
		sort.Slice(scrapes, func(i, j int) bool {
			return scrapes[i].Price < scrapes[j].Price
		})

		fmt.Fprintf(&message, "**%s**\n", product.Name)
		fmt.Fprintf(&message, "Base price: £%.2f\n", product.BasePrice)

		cheapest := scrapes[0]
		fmt.Fprintln(&message, cheapest.GetCheapestPriceString(product))

		remaining := scrapes[1:]
		if len(remaining) > 0 {
			fmt.Fprintln(&message, "Other prices:")
			for _, scrape := range remaining {
				fmt.Fprintln(&message, scrape.GetOtherPriceString(product))
			}
		}

		fmt.Fprintln(&message)
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
