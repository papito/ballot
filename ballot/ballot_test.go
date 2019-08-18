package main

import (
	"encoding/json"
	"fmt"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/models"
	"github.com/papito/ballot/ballot/server"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var envConfig config.Config
var srv server.Server

// setup/teardown
func TestMain(m *testing.M) {
	err := os.Setenv("ENV", config.TEST)
	if err != nil {
		panic(err)
	}

	envConfig = config.LoadConfig()

	srv = server.NewServer(envConfig)
	code := m.Run()
	srv.Release()

	os.Exit(code)
}

func TestHealthEndpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(srv.HealthHttpHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	var health = models.Health{Status:"OK"}
	var data, _ = json.Marshal(health)

	expected := fmt.Sprintf("%s", data)

	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}