package main

import (
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/server"
	"log"
	"net/http"
)

func main() {
	envConfig := config.LoadConfig()
	srv := server.NewServer(envConfig)
	defer srv.Release()

	log.Printf("Starting server on port %s", envConfig.HttpPort)
	log.Fatal(http.ListenAndServe(envConfig.HttpPort, nil))
}
