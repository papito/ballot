/*
 * The MIT License
 *
 * Copyright (c) 2019,  Andrei Taranchenko
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package config

import (
	"log"
	"os"
)

const SessionTtl = 172800 // 48H

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
