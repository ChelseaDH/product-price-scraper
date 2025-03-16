package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

type Product struct {
	Name          string
	BasePrice     float64
	Category      string
	RetailerLinks map[*Retailer]string
}
type Products []Product

type FailedScrape struct {
	Product  *Product
	Retailer *Retailer
	Error    error
}

func (f FailedScrape) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("product", f.Product.Name),
		slog.String("retailer", f.Retailer.Name),
		slog.String("err", f.Error.Error()),
	)
}

type FailedScrapes []FailedScrape

func (fs FailedScrapes) LogValue() slog.Value {
	attrs := make([]slog.Value, len(fs))
	for i, f := range fs {
		attrs[i] = f.LogValue()
	}
	return slog.AnyValue(attrs)
}

type SuccessScrape struct {
	Retailer    *Retailer
	Price       float64
	Url         string
	CachedPrice *float64
}

func GetProducts(config Config, retailers map[string]*Retailer) Products {
	var products []Product

	for _, p := range config.Products {
		product := Product{
			Name:          p.Name,
			BasePrice:     p.BasePrice,
			Category:      p.Category,
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

func (p Products) GetPrices(ctx context.Context, cachedPrices map[CacheKey]float64) (map[*Product][]SuccessScrape, []FailedScrape) {
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

			key := CacheKey{
				Retailer: retailer.Name,
				Product:  product.Name,
			}
			cachedPrice, _ := cachedPrices[key]

			successScrapes = append(successScrapes, SuccessScrape{Retailer: retailer, Price: price, Url: link, CachedPrice: &cachedPrice})
		}

		prices[&p[i]] = successScrapes
	}

	return prices, failures
}

func (p Products) FindPricesAndNotify(ctx context.Context, logger *slog.Logger, client Client, cache *Cache, minDiscount float64) error {
	logger.Info("Starting scrape")

	cachedPrices, err := cache.GetScrapes()
	if err != nil {
		return fmt.Errorf("error getting cached prices: %v", err)
	}

	prices, failures := p.GetPrices(ctx, cachedPrices)
	if failures != nil {
		logger.Warn("Failures returned from getting prices", slog.Any("failures", failures))
	}

	notifiablePrices := GetNotifiablePrices(prices, minDiscount)
	if len(notifiablePrices) == 0 {
		logger.Info("No prices found to notify")
	} else {
		logger.Info("Prices found to notify", slog.Any("prices", notifiablePrices))
		err = notify(notifiablePrices, client)
		if err != nil {
			return fmt.Errorf("error notifying products: %v", err)
		}
	}

	return cache.SetScrapes(prices)
}

func (p Product) getDiscountString(price float64) string {
	discount := p.BasePrice - price
	percentage := discount / p.BasePrice * 100
	return fmt.Sprintf("(-£%.2f | %.2f%% off)", discount, percentage)
}

func (s *SuccessScrape) GetCheapestPriceString(product *Product) string {
	return s.formatPriceString(product, "Best price: ", true)
}

func (s *SuccessScrape) GetOtherPriceString(product *Product) string {
	return s.formatPriceString(product, "- ", false)
}

func (s *SuccessScrape) formatPriceString(product *Product, prefix string, bold bool) string {
	var output strings.Builder

	output.WriteString(prefix)

	switch {
	case s.CachedPrice == nil:
		output.WriteString("🆕 ")
	case s.Price > *s.CachedPrice:
		output.WriteString("🔺 ")
	}

	priceFormat := "£%.2f"
	if bold {
		priceFormat = "**£%.2f**"
	}

	output.WriteString(fmt.Sprintf(priceFormat+" at [%s](%s) %s", s.Price, s.Retailer.Name, s.Url, product.getDiscountString(s.Price)))

	return output.String()
}
