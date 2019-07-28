package config

import "os"

const (
	TEST = "test"
	DEV = "development"
	PROD = "production"
)

type Config struct {
	Environment string
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

	config.HttpPort = ":" + os.Getenv("HTTP_PORT")
	if config.HttpPort == ":" {
		panic("Specify HTTP_PORT environment variable")
	}

	config.RedisUrl = os.Getenv("REDIS_URL")
	if config.RedisUrl == ":" {
		panic("Specify REDIS_URL environment variable")
	}

	return config
}
