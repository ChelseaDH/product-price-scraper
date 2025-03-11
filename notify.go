package main

import (
	"bytes"
	"fmt"
	"sort"
)

func notify(prices map[*Product][]SuccessScrape, client Client, minDiscount float64) error {
	message := bytes.Buffer{}
	fmt.Fprintf(&message, "🛍️ **New prices found** 🤑\n\n")

	for product, scrapes := range prices {
		var cheaperPrices []SuccessScrape
		for _, scrape := range scrapes {
			if scrape.Price <= (product.BasePrice * (1 - minDiscount)) {
				cheaperPrices = append(cheaperPrices, scrape)
			}
		}

		if len(cheaperPrices) == 0 {
			continue
		}

		sort.Slice(cheaperPrices, func(i, j int) bool {
			return cheaperPrices[i].Price < cheaperPrices[j].Price
		})

		fmt.Fprintf(&message, "**%s**\n", product.Name)
		fmt.Fprintf(&message, "Base price: £%.2f\n", product.BasePrice)

		cheapest := cheaperPrices[0]
		fmt.Fprintf(&message, "Best price: **£%.2f** at [%s](%s) %s\n", cheapest.Price, cheapest.Retailer.Name,
			cheapest.Url, product.getDiscountString(cheapest.Price))

		remaining := cheaperPrices[1:]
		if len(remaining) > 0 {
			fmt.Fprintln(&message, "Other prices:")
			for _, scrape := range remaining {
				fmt.Fprintf(&message, "- £%.2f at [%s](%s) %s\n", scrape.Price, scrape.Retailer.Name, scrape.Url,
					product.getDiscountString(scrape.Price))
			}
		}

		fmt.Fprintln(&message)
	}

	return client.SendMessage(message.String())
}

func shouldNotify(prices map[*Product][]SuccessScrape, cachedPrices map[CacheKey]float64, minDiscount float64) bool {
	for product, scrapes := range prices {
		for _, scrape := range scrapes {
			key := CacheKey{
				Retailer: scrape.Retailer.Name,
				Product:  product.Name,
			}
			cachedPrice, ok := cachedPrices[key]

			if (ok && scrape.Price != cachedPrice) || (!ok && scrape.Price <= (product.BasePrice*(1-minDiscount))) {
				return true
			}
		}
	}
	return false
}
