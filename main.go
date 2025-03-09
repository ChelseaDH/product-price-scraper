package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancelCtx := context.WithCancel(context.Background())
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancelCtx()
	}()

	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	cache, err := NewCache(config.General.Database)
	if err != nil {
		fmt.Println("Error instantiating cache:", err)
		return
	}

	client, err := getClient(ctx, config)
	if err != nil {
		fmt.Println("Error getting client:", err)
		return
	}

	retailers := GetRetailers()
	products := GetProducts(config, retailers)

	err = products.FindPricesAndNotify(ctx, client, cache, config.General.MinDiscount)
	if err != nil {
		fmt.Println("Error finding prices and notifying:", err)
	}

loop:
	for {
		next := time.Now().Add(config.General.Interval)
		interval := time.Until(next)
		fmt.Printf("Next scrape at %s (in %s)\n", next, interval)

		select {
		case <-time.After(interval):
			err = products.FindPricesAndNotify(ctx, client, cache, config.General.MinDiscount)
			if err != nil {
				fmt.Println("Error finding prices and notifying:", err)
			}
		case <-ctx.Done():
			break loop
		}
	}

	err = client.Stop()
	if err != nil {
		fmt.Println("Error stopping client:", err)
	}
}
