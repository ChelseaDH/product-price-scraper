package main

import "testing"

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

func TestNotify(t *testing.T) {
	retailer := &Retailer{
		Name: "Test Retailer",
	}
	prices := map[*Product][]SuccessScrape{
		&Product{
			Name:      "Test Product",
			BasePrice: 100.00,
		}: {
			{
				Retailer: retailer,
				Price:    90.00,
				Url:      "https://test.com/1",
			},
			{
				Retailer: retailer,
				Price:    100.00,
				Url:      "https://test.com/2",
			},
		},
		&Product{
			Name:      "Test Product 2",
			BasePrice: 100.00,
		}: {
			{
				Retailer: retailer,
				Price:    95.00,
				Url:      "https://test.com/2",
			},
			{
				Retailer: retailer,
				Price:    90.01,
				Url:      "https://test.com/2",
			},
		},
	}
	client := &TestClient{}

	err := notify(prices, client, 0.1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "🛍️ **New prices found** 🤑\n\n**Test Product**\nBase price: £100.00\nBest price: **£90." +
		"00** at [Test Retailer](https://test.com/1) (-£10.00 | 10.00% off)\n\n"
	if client.message != expected {
		t.Errorf("unexpected message: expected %s, got %s", expected, client.message)
	}
}

func TestShouldNotify(t *testing.T) {
	product := &Product{
		Name:      "Test Product",
		BasePrice: 100.00,
	}
	retailer := &Retailer{
		Name: "Test Retailer",
	}

	type args struct {
		prices       map[*Product][]SuccessScrape
		cachedPrices map[CacheKey]float64
		minDiscount  float64
	}
	tests := []struct {
		name     string
		args     args
		expected bool
	}{
		{
			name: "should notify if price is different to cache",
			args: args{
				prices: map[*Product][]SuccessScrape{
					product: {
						{retailer, product.BasePrice - 10, "https://test.com/1"},
					},
				},
				cachedPrices: map[CacheKey]float64{
					CacheKey{retailer.Name, product.Name}: product.BasePrice,
				},
				minDiscount: 0.1,
			},
			expected: true,
		},
		{
			name: "should not notify if price is the same as cache",
			args: args{
				prices: map[*Product][]SuccessScrape{
					product: {
						{retailer, product.BasePrice - 10, "https://test.com/1"},
					},
				},
				cachedPrices: map[CacheKey]float64{
					CacheKey{retailer.Name, product.Name}: product.BasePrice - 10,
				},
				minDiscount: 0.1,
			},
			expected: false,
		},
		{
			name: "should not notify if cache does not exist and price is not lower",
			args: args{
				prices: map[*Product][]SuccessScrape{
					product: {
						{retailer, product.BasePrice, "https://test.com/1"},
					},
				},
				cachedPrices: nil,
				minDiscount:  0.1,
			},
			expected: false,
		},
		{
			name: "should not notify if cache does not exist and price not lower than min discount",
			args: args{
				prices: map[*Product][]SuccessScrape{
					product: {
						{retailer, product.BasePrice * 0.95, "https://test.com/1"},
					},
				},
				cachedPrices: nil,
				minDiscount:  0.1,
			},
			expected: false,
		},
		{
			name: "should notify if cache does not exist and price is lower by the exact min discount",
			args: args{
				prices: map[*Product][]SuccessScrape{
					product: {
						{retailer, product.BasePrice * 0.9, "https://test.com/1"},
					},
				},
				cachedPrices: nil,
				minDiscount:  0.1,
			},
			expected: true,
		},
		{
			name: "should notify if cache does not exist and price is lower",
			args: args{
				prices: map[*Product][]SuccessScrape{
					product: {
						{retailer, product.BasePrice * 0.8, "https://test.com/1"},
					},
				},
				cachedPrices: nil,
				minDiscount:  0.1,
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldNotify(tt.args.prices, tt.args.cachedPrices, tt.args.minDiscount); got != tt.expected {
				t.Errorf("shouldNotify() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
