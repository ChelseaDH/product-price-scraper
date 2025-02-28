package main

import (
	"github.com/pelletier/go-toml"
	"os"
)

type Config struct {
	Products []ProductTOML `toml:"products"`
}

type ProductTOML struct {
	Name      string            `toml:"name"`
	BasePrice float64           `toml:"base_price"`
	Links     map[string]string `toml:"links"`
}

func loadConfig() (Config, error) {
	var config Config
	b, err := os.ReadFile("config.toml")
	if err != nil {
		return config, err
	}

	err = toml.Unmarshal(b, &config)
	return config, nil
}

func getProductsFromConfig(config Config, retailers map[string]*Retailer) []Product {
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
