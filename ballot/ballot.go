package main

import (
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/server"
	"log"
	"net/http"
	"os"
)

func main() {
	envConfig := config.LoadConfig()
	srv := server.NewServer(envConfig)
	defer srv.Release()

	serverPort := ":" + os.Getenv("HTTP_PORT")

	if serverPort == ":" {

		panic("Specify HTTP_PORT environment variable")
	}

	log.Printf("Starting server on port %s", serverPort)
	log.Fatal(http.ListenAndServe(serverPort, nil))
}