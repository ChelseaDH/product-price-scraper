package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type Cache struct {
	db *sql.DB
}

type CacheKey struct {
	Retailer, Product string
}

func NewCache(dbPath string) (*Cache, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS scrape_cache (provider TEXT, product TEXT, price INTEGER, last_scrape INTEGER, PRIMARY KEY (provider, product))")
	if err != nil {
		return nil, err
	}

	return &Cache{db: db}, nil
}

func (c *Cache) GetScrapes() (map[CacheKey]float64, error) {
	rows, err := c.db.Query("SELECT provider, product, price FROM scrape_cache")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scrapes := make(map[CacheKey]float64)
	for rows.Next() {
		var (
			provider, product string
			price             int
		)
		err = rows.Scan(&provider, &product, &price)
		if err != nil {
			return nil, err
		}

		scrapes[CacheKey{
			Retailer: provider,
			Product:  product,
		}] = float64(price) / 100
	}

	return scrapes, nil
}

func (c *Cache) SetScrapes(scrapes map[*Product][]SuccessScrape) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO scrape_cache (provider, product, price, last_scrape) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}

	for product, successScrapes := range scrapes {
		for _, successScrape := range successScrapes {
			_, err = stmt.Exec(successScrape.Retailer.Name, product.Name, int(successScrape.Price*100), time.Now().Unix())
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}
