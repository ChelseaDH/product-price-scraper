package main

type Product struct {
	Name          string
	BasePrice     float64
	RetailerLinks map[*Retailer]string
}

func getProducts(config Config, retailers map[string]*Retailer) []Product {
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
