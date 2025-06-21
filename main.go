package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancelCtx := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancelCtx()
	}()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey && a.Value.Any() == levelFatal {
				return slog.String(slog.LevelKey, "FATAL")
			}
			return a
		},
	}))

	config, err := loadConfig()
	if err != nil {
		LogFatal(ctx, logger, "Failed to load config", err)
		return
	}

	cache, err := NewCache(config.General.Database)
	if err != nil {
		LogFatal(ctx, logger, "Failed to instantiate cache", err)
		return
	}

	client, err := getClient(ctx, logger, config)
	if err != nil {
		LogFatal(ctx, logger, "Failed to get client", err)
		return
	}

	retailers := GetRetailers()
	products := GetProducts(config, retailers)

	err = products.FindPricesAndNotify(ctx, logger, client, cache, config.General.MinDiscount)
	if err != nil {
		LogError(logger, "Failed to find prices and notify", err)
	}

loop:
	for {
		next := time.Now().Add(config.General.Interval)
		interval := time.Until(next)
		logger.Info(fmt.Sprintf("Next scrape at %s (in %s)\n", next, interval))

		select {
		case <-time.After(interval):
			err = products.FindPricesAndNotify(ctx, logger, client, cache, config.General.MinDiscount)
			if err != nil {
				LogError(logger, "Failed to find prices and notify", err)
			}
		case <-ctx.Done():
			break loop
		}
	}

	err = client.Stop()
	if err != nil {
		LogError(logger, "Failed to stop client", err)
	}
}

const levelFatal = slog.Level(12)

func LogFatal(ctx context.Context, logger *slog.Logger, msg string, err error) {
	logger.LogAttrs(ctx, levelFatal, msg, slog.Any("err", err))
}

func LogError(logger *slog.Logger, msg string, err error) {
	logger.Error(msg, slog.Any("err", err))
}
