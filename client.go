package main

import (
	"context"
	"fmt"
)

type Client interface {
	SendMessage(markdown string) error
	Stop() error
}

type DefaultClient struct{}

func (d *DefaultClient) SendMessage(markdown string) error {
	_, err := fmt.Println(markdown)
	return err
}

func (d *DefaultClient) Stop() error {
	return nil
}

func getClient(ctx context.Context, config Config) (Client, error) {
	if config.Matrix == nil {
		return &DefaultClient{}, nil
	} else {
		return connectToMatrix(ctx, *config.Matrix, config.General.Database)
	}
}
