package main

import (
	"ballot/ballot/models"
	"ballot/ballot/server"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	srv := server.NewServer()
	defer srv.Release()

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