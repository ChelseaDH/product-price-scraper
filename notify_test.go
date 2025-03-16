package main

import (
	"testing"
)

type TestClient struct {
	message string
}

func (t *TestClient) SendMessage(markdown string) error {
	t.message = markdown
	return nil
}

func (t *TestClient) Stop() error {
	return nil
}

func floatPtr(f float64) *float64 {
	return &f
}

func TestNotify(t *testing.T) {
	retailer := &Retailer{
		Name: "Test Retailer",
	}
	retailer2 := &Retailer{
		Name: "Test Retailer 2",
	}
	prices := map[*Product][]SuccessScrape{
		&Product{
			Name:      "Test Product",
			BasePrice: 100.00,
			Category:  "Category 1",
		}: {
			{
				Retailer:    retailer,
				Price:       80.00,
				Url:         "https://test.com/1",
				CachedPrice: floatPtr(80.00),
			},
			{
				Retailer:    retailer2,
				Price:       90.00,
				Url:         "https://test2.com/1",
				CachedPrice: floatPtr(80.00),
			},
		},
		&Product{
			Name:      "Test Product 2",
			BasePrice: 90.00,
			Category:  "Category 2",
		}: {
			{
				Retailer:    retailer,
				Price:       60.00,
				Url:         "https://test.com/2",
				CachedPrice: nil,
			},
		},
		&Product{
			Name:      "Test Product 3",
			BasePrice: 100.00,
			Category:  "Category 1",
		}: {
			{
				Retailer:    retailer,
				Price:       95.00,
				Url:         "https://test.com/3",
				CachedPrice: floatPtr(95.00),
			},
			{
				Retailer: retailer,
				Price:    90.01,
				Url:      "https://test.com/4",
			},
		},
		&Product{
			Name:      "Test Product 4",
			BasePrice: 100.00,
		}: {
			{
				Retailer:    retailer,
				Price:       75.00,
				Url:         "https://test.com/4",
				CachedPrice: floatPtr(95.00),
			},
		},
	}
	client := &TestClient{}

	err := notify(prices, client)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "🛍️ **Cheaper prices found** 🤑\n\n" +
		"**Category 1**\n\n" +
		"**Test Product**\nBase price: £100.00\nBest price: **£80.00** at [Test Retailer](https://test.com/1) (-£20.00 | 20.00% off)\n" +
		"Other prices:\n- 🔺 £90.00 at [Test Retailer 2](https://test2.com/1) (-£10.00 | 10.00% off)\n\n" +
		"**Test Product 3**\nBase price: £100.00\nBest price: 🆕 **£90.01** at [Test Retailer](https://test.com/4) (-£9.99 | 9.99% off)\n" +
		"Other prices:\n- £95.00 at [Test Retailer](https://test.com/3) (-£5.00 | 5.00% off)\n\n" +
		"**Category 2**\n\n" +
		"**Test Product 2**\nBase price: £90.00\nBest price: 🆕 **£60.00** at [Test Retailer](https://test.com/2) (-£30.00 | 33.33% off)\n\n" +
		"**Other**\n\n" +
		"**Test Product 4**\nBase price: £100.00\nBest price: **£75.00** at [Test Retailer](https://test.com/4) (-£25.00 | 25.00% off)\n\n"
	if client.message != expected {
		t.Errorf("unexpected message: expected %s\n\ngot: %s", expected, client.message)
	}
}

func TestGetNotifiablePrices(t *testing.T) {
	product := &Product{
		Name:      "Test Product",
		BasePrice: 100.00,
	}
	retailer := &Retailer{
		Name: "Test Retailer",
	}
	prices := map[*Product][]SuccessScrape{
		product: {
			// Prices without a cached price:
			// Price is the same as the base price => should not be included
			{retailer, product.BasePrice, "https://test.com/1", nil},
			// Price is lower than the base price by less the min discount => should not be included
			{retailer, product.BasePrice * 0.95, "https://test.com/2", nil},
			// Price is lower than the base price by the min discount => should be included
			{retailer, product.BasePrice * 0.9, "https://test.com/3", nil},
			// Price is lower than the base price by more than the min discount => should be included
			{retailer, product.BasePrice * 0.8, "https://test.com/4", nil},

			// Prices with a cached price:
			// Price is the same as the cached price => should not be included
			{retailer, product.BasePrice * 0.9, "https://test.com/5", floatPtr(product.BasePrice * 0.9)},
			// Price is below the base threshold but higher than the lower cache threshold => should not be included
			{retailer, product.BasePrice * 0.85, "https://test.com/6", floatPtr(product.BasePrice * 0.9)},
			// Price falls below the base threshold but is higher than the lower cache threshold  => should be included
			{retailer, product.BasePrice * 0.85, "https://test.com/7", floatPtr(product.BasePrice * 0.91)},
			// Price is below the base threshold and at the lower cache threshold => should be included
			{retailer, product.BasePrice * 0.81, "https://test.com/8", floatPtr(product.BasePrice * 0.9)},
			// Price is below the base and cache price thresholds => should be included
			{retailer, product.BasePrice * 0.79, "https://test.com/9", floatPtr(product.BasePrice * 0.9)},
			// Price is below the base threshold but has increased by less than the upper cache threshold => should not be included
			{retailer, product.BasePrice * 0.83, "https://test.com/10", floatPtr(product.BasePrice * 0.8)},
			// Price is below the base threshold and has increased to the upper cache threshold => should be included
			{retailer, product.BasePrice * 0.88, "https://test.com/11", floatPtr(product.BasePrice * 0.8)},
			// Price is below the base threshold and has increased beyond the upper cache threshold => should be included
			{retailer, product.BasePrice * 0.89, "https://test.com/12", floatPtr(product.BasePrice * 0.8)},
			// Price is above the base threshold but is below the upper cache threshold => should not be included
			{retailer, product.BasePrice * 0.91, "https://test.com/13", floatPtr(product.BasePrice * 0.8)},
		},
	}
	expected := map[*Product][]SuccessScrape{
		product: {
			{retailer, product.BasePrice * 0.9, "https://test.com/3", nil},
			{retailer, product.BasePrice * 0.8, "https://test.com/4", nil},
			{retailer, product.BasePrice * 0.85, "https://test.com/7", floatPtr(product.BasePrice * 0.91)},
			{retailer, product.BasePrice * 0.81, "https://test.com/8", floatPtr(product.BasePrice * 0.9)},
			{retailer, product.BasePrice * 0.79, "https://test.com/9", floatPtr(product.BasePrice * 0.9)},
			{retailer, product.BasePrice * 0.88, "https://test.com/11", floatPtr(product.BasePrice * 0.8)},
			{retailer, product.BasePrice * 0.89, "https://test.com/12", floatPtr(product.BasePrice * 0.8)},
		},
	}

	filteredPrices := GetNotifiablePrices(prices, 0.1)

	actualScrapes := filteredPrices[product]
	expectedScrapes := expected[product]
	if len(expectedScrapes) != len(actualScrapes) {
		t.Errorf("unexpected length: expected %d, got %d", len(expectedScrapes), len(actualScrapes))
	}

	for i, expectedScrape := range expectedScrapes {
		actualScrape := actualScrapes[i]
		if expectedScrape.Price != actualScrape.Price {
			t.Errorf("unexpected price: expected %.2f, got %.2f", expectedScrape.Price, actualScrape.Price)
		}
		if expectedScrape.Url != actualScrape.Url {
			t.Errorf("unexpected url: expected %s, got %s", expectedScrape.Url, actualScrape.Url)
		}
	}
}
