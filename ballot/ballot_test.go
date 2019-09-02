package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/db"
	"github.com/papito/ballot/ballot/hub"
	"github.com/papito/ballot/ballot/model"
	"github.com/papito/ballot/ballot/server"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
)

var envConfig config.Config
var srv server.Server
var testHub *hub.VoidHub

// setup/teardown
func TestMain(m *testing.M) {
	err := os.Setenv("ENV", config.TEST)
	if err != nil {panic(err)}

	// remove logs in test
	log.SetOutput(ioutil.Discard)

	envConfig = config.LoadConfig()

	srv = server.NewServer(envConfig)
	testHub = srv.Service().Hub().(*hub.VoidHub)

	code := m.Run()

	srv.Release()
	os.Exit(code)
}

func RandString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func createSessionAndUsers(numOfUsers int, t *testing.T) (session model.Session, userIds []model.User) {
	session, err  := srv.Service().CreateSession()
	if err != nil {t.Errorf("Could not create session: %s", err)}

	users := make([]model.User, numOfUsers)
	for i := 0; i < numOfUsers; i++ {
		user, err := srv.Service().CreateUser(session.SessionId, RandString(20))
		if err != nil {t.Errorf("Could not create user: %s", err)}
		users = append(users, user)
	}

	return session, users
}

func TestHealthEndpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {t.Error(err)}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.HealthHttpHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)

	var health = model.HealthResponse{Status: "OK"}
	var data, _ = json.Marshal(health)

	expected := fmt.Sprintf("%s", data)
	assert.Equal(t, rr.Body.String(), expected)
}

func TestCreateSessionEndpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {t.Error(err)}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.CreateSessionHttpHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)

	var session model.Session
	err = json.Unmarshal([]byte(rr.Body.String()), &session)
	if err != nil {t.Errorf("%s. Recevied: %s", err, rr.Body.String())}

	// FIXME: length can also be checked with a regex
	match, _ := regexp.MatchString("[a-z0-9]", session.SessionId)

	assert.True(t, match)
	assert.Len(t, session.SessionId, 36)

	sessionKey := fmt.Sprintf(db.Const.SessionVoting, session.SessionId)
	sessionState, err := srv.Service().Store().GetInt(sessionKey)
	assert.Equal(t, sessionState, model.NotVoting)
}

func TestCreateUserEndpoint(t *testing.T) {
	session, err  := srv.Service().CreateSession()
	if err != nil {t.Errorf("Could not create session: %s", err)}

	userName := "  Player 1  "
	reqObj := model.CreateUserRequest{
		UserName:  userName,
		SessionId: session.SessionId,
	}

	body, err := json.Marshal(reqObj)
	if err != nil {t.Error(err)}

	req, err := http.NewRequest("POST", "/api/user", bytes.NewBufferString(string(body)))
	if err != nil {t.Error(err)}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.CreateUserHttpHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)

	var user model.User
	err = json.Unmarshal([]byte(rr.Body.String()), &user)

	assert.Equal(t, user.Name, "Player 1")
	assert.Equal(t, user.Estimate, model.NoEstimate)
	assert.NotNil(t, user.UserId)
}

func TestStartVoteEndpoint(t *testing.T) {
	session, _ := createSessionAndUsers(2, t)

	reqObj := model.StartVoteRequest{SessionId:session.SessionId}

	body, err := json.Marshal(reqObj)
	if err != nil {t.Error(err)}

	req, err := http.NewRequest("PUT", "/api/vote/start", bytes.NewBufferString(string(body)))
	if err != nil {t.Error(err)}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.StartVoteHttpHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)

	sessionKey := fmt.Sprintf(db.Const.SessionVoting, session.SessionId)
	sessionState, err := srv.Service().Store().GetInt(sessionKey)
	assert.Equal(t, sessionState, model.Voting)
}