package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/db"
	"github.com/papito/ballot/ballot/models"
	"github.com/papito/ballot/ballot/requests"
	"github.com/papito/ballot/ballot/server"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, rr.Code, http.StatusOK)

	var health = models.Health{Status: "OK"}
	var data, _ = json.Marshal(health)

	expected := fmt.Sprintf("%s", data)
	assert.Equal(t, rr.Body.String(), expected)
}

func TestCreateSession(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.CreateSessionHttpHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)

	var session models.Session
	err = json.Unmarshal([]byte(rr.Body.String()), &session)
	if err != nil {
		t.Errorf("%s. Recevied: %s", err, rr.Body.String())
	}
	match, _ := regexp.MatchString("[a-z0-9]", session.SessionId)

	assert.True(t, match)
	assert.Len(t, session.SessionId, 36)

	sessionKey := fmt.Sprintf(db.Const.SessionVoting, session.SessionId)
	sessionState, err := srv.Store().GetInt(sessionKey)
	assert.Equal(t, sessionState, models.NotVoting)
}

func TestCreateUser(t *testing.T) {
	session, err  := srv.Service().CreateSession()
	if err != nil {
		t.Errorf("Could not create session: %s", err)
	}

	userName := "  Player 1  "

	reqData := requests.CreateUserRequest{
		UserName:  userName,
		SessionId: session.SessionId,
	}

	body, err := json.Marshal(reqData)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/api/user", bytes.NewBufferString(string(body)))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.CreateUserHttpHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)

	var user models.User
	err = json.Unmarshal([]byte(rr.Body.String()), &user)

	assert.Equal(t, user.Name, "Player 1")
	assert.Equal(t, user.Estimate, models.NoEstimate)
	assert.NotNil(t, user.UserId)
}
