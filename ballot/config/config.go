package config

import (
	"log"
	"os"
)

const (
	TEST = "test"
	DEV = "development"
	PROD = "production"
)

type Config struct {
	Environment string
	HttpHost string
	HttpPort string
	RedisUrl string
}

func LoadConfig() Config {
	config := Config{}

	config.Environment = os.Getenv("ENV")
	if config.Environment == "" {
		config.Environment = DEV
		err := os.Setenv("ENV", config.Environment)

		if err != nil {
			panic(err)
		}
	}
	log.Printf("ENV: %s", config.Environment)

	config.HttpPort = ":" + os.Getenv("HTTP_PORT")
	if config.HttpPort == ":" {
		config.HttpPort = ":8080"
	}
	log.Printf("HTTP port %s", config.HttpPort)

	config.HttpHost = os.Getenv("HTTP_HOST")
	if config.HttpHost == "" {
		config.HttpHost = "http://localhost" + config.HttpPort
	}
	log.Printf("HTTP host %s", config.HttpHost)

	config.RedisUrl = os.Getenv("REDIS_URL")
	if config.RedisUrl == "" {
		config.RedisUrl = "localhost:6379"
	}
	log.Printf("Redis URL %s", config.RedisUrl)

	return config
}
