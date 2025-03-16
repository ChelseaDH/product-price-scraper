package main

import (
	"github.com/pelletier/go-toml"
	"os"
	"time"
)

type Config struct {
	General  General       `toml:"general"`
	Matrix   *Matrix       `toml:"matrix"`
	Products []ProductTOML `toml:"products"`
}

type General struct {
	Database    string        `toml:"database"`
	Interval    time.Duration `toml:"interval"`
	MinDiscount float64       `toml:"min_discount"`
}

type Matrix struct {
	HomeServer  string `toml:"home_server"`
	UserName    string `toml:"username"`
	AccessToken string `toml:"access_token"`
	RoomID      string `toml:"room_id"`
}

type ProductTOML struct {
	Name      string            `toml:"name"`
	BasePrice float64           `toml:"base_price"`
	Category  string            `toml:"category"`
	Links     map[string]string `toml:"links"`
}

func loadConfig() (Config, error) {
	var config Config
	file, err := os.Open("config.toml")
	if err != nil {
		return config, err
	}
	defer file.Close()

	return config, toml.NewDecoder(file).Decode(&config)
}
