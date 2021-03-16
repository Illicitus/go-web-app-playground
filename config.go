package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (pc PostgresConfig) ConnectionInfo() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		pc.Host, pc.Port, pc.User, pc.Password, pc.Name,
	)
}

func (pc PostgresConfig) Dialect() string {
	return "postgres"
}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "password",
		Name:     "go_playground",
	}
}

type AppConfig struct {
	Port     int            `json:"port"`
	Env      string         `json:"env"`
	Pepper   string         `json:"pepper"`
	HMACKey  string         `json:"hmac_key"`
	Database PostgresConfig `json:"database"`
}

func (c AppConfig) IsProd() bool {
	return c.Env == "prod"
}

func DefaultAppConfig() AppConfig {
	return AppConfig{
		Port:     3000,
		Env:      "dev",
		Pepper:   "secret-random-string",
		HMACKey:  "secret-hmac-key",
		Database: DefaultPostgresConfig(),
	}
}

func LoadConfig() AppConfig {
	f, err := os.Open("config.json")
	if err != nil {
		fmt.Println("Using the default config...")
		return DefaultAppConfig()
	}
	var ac AppConfig
	dec := json.NewDecoder(f)
	err = dec.Decode(&ac)
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully upload config...")
	return ac
}
