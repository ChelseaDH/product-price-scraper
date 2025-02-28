package main

import (
	"fmt"
)

func main() {
	boots := Retailer{Name: "Boots", Scraper: &BootsScraper{}}
	amazon := Retailer{Name: "Amazon", Scraper: &AmazonScraper{}}
	lookFantastic := Retailer{Name: "LookFantastic", Scraper: &LookFantasticScraper{}}

	products := []Product{
		{
			Name:      "Byoma Moisturizing Gel Cream",
			BasePrice: 11.99,
			RetailerLinks: map[*Retailer]string{
				&boots:  "https://www.boots.com/byoma-moisturizing-gel-cream-50ml-10307026",
				&amazon: "https://www.amazon.co.uk/BYOMA-Moisturizing-Gel-Cream-50ml/dp/B0BJ74DJXG",
			},
		},
		{
			Name:      "Byoma Balancing Face Mist",
			BasePrice: 11.99,
			RetailerLinks: map[*Retailer]string{
				&boots:  "https://www.boots.com/byoma-balancing-face-mist-100ml-10307029",
				&amazon: "https://www.amazon.co.uk/BYOMA-Balancing-Hydrating-Face-100ml/dp/B0C7C8D9LS",
			},
		},
		{
			Name:      "INKEY List Oat Cleansing Balm",
			BasePrice: 12.00,
			RetailerLinks: map[*Retailer]string{
				&boots:         "https://www.boots.com/the-inkey-list-oat-cleansing-balm-150ml-10278182",
				&amazon:        "https://www.amazon.co.uk/INKEY-List-Cleansing-Removes-Sensitive/dp/B09MRD1648",
				&lookFantastic: "https://www.lookfantastic.com/p/the-inkey-list-oat-cleansing-balm-150ml/12435694/",
			},
		},
		{
			Name:      "INKEY List Q10 Serum",
			BasePrice: 9.00,
			RetailerLinks: map[*Retailer]string{
				&amazon:        "https://www.amazon.co.uk/INKEY-List-Antioxidant-Serum-Protect-dp-B09N9ZKWT8/dp/B09N9ZKWT8",
				&lookFantastic: "https://www.lookfantastic.com/p/the-inkey-list-q10-serum-30ml/12208008/",
			},
		},
	}

	cheaperPrices := make(map[*Product]map[*Retailer]float64)

	for _, product := range products {
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

	// Print out the cheaper prices
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
