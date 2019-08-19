package main

import (
	"encoding/json"
	"fmt"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/models"
	"github.com/papito/ballot/ballot/server"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
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

	log.SetOutput(ioutil.Discard)

	envConfig = config.LoadConfig()

	srv = server.NewServer(envConfig)
	code := m.Run()
	srv.Release()

	os.Exit(code)
}

func TestHealth(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.HealthHttpHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var health = models.Health{Status:"OK"}
	var data, _ = json.Marshal(health)

	expected := fmt.Sprintf("%s", data)

	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestCreateSession(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.CreateSessionHttpHandler)
	handler.ServeHTTP(rr, req)

	// TODO: utility function
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var session models.Session
	err = json.Unmarshal([]byte(rr.Body.String()), &session)
	if err != nil {
		t.Errorf("%s. Recevied: %s", err, rr.Body.String())
	}
	match, _ := regexp.MatchString("[a-z0-9]", session.SessionId)

	if !match {
		t.Errorf("ID [%s] is not valid UUID", session.SessionId)
	}
	if len(session.SessionId) != 36 { // FIXME: this can be done with the regex above
		t.Errorf("ID [%s] is not a valid UUID", session.SessionId)
	}
}
