package main

import (
	"ballot/ballot/server"
	"log"
	"net/http"
)

// FIXME: env  var

func main() {
	srv := server.NewServer()
	defer srv.Release()

	log.Println("Starting server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}