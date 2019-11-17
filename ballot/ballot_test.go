package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/db"
	"github.com/papito/ballot/ballot/hub"
	"github.com/papito/ballot/ballot/model"
	"github.com/papito/ballot/ballot/model/request"
	"github.com/papito/ballot/ballot/model/response"
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
		users[i] = user
	}

	clearHubEvents()
	return session, users
}

func clearHubEvents() {
	testHub.Connect(nil)
}

func TestHealthEndpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {t.Error(err)}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.HealthHttpHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t,  http.StatusOK, rr.Code)

	var health = response.HealthResponse{Status: "OK"}
	var data, _ = json.Marshal(health)

	expected := fmt.Sprintf("%s", data)
	assert.Equal(t, expected, rr.Body.String())
}

func TestCreateSessionEndpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {t.Error(err)}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.CreateSessionHttpHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t,  http.StatusOK, rr.Code)

	var session model.Session
	err = json.Unmarshal([]byte(rr.Body.String()), &session)
	if err != nil {t.Errorf("%s. Recevied: %s", err, rr.Body.String())}

	match, _ := regexp.MatchString("[a-z0-9]", session.SessionId)

	assert.True(t, match)
	assert.Len(t, session.SessionId, 36)

	key := fmt.Sprintf(db.Const.SessionState, session.SessionId)
	sessionState, err := srv.Service().Store().GetInt(key)
	assert.Equal(t, sessionState, model.NotVoting)

	key = fmt.Sprintf(db.Const.UserCount, session.SessionId)
	userCount, err := srv.Service().Store().GetInt(key)
	assert.Equal(t, 0, userCount)

	key = fmt.Sprintf(db.Const.VoteCount, session.SessionId)
	voteCount, err := srv.Service().Store().GetInt(key)
	assert.Equal(t, 0, voteCount)
}

func TestCreateUserEndpoint(t *testing.T) {
	clearHubEvents()

	session, err  := srv.Service().CreateSession()
	if err != nil {t.Errorf("Could not create session: %s", err)}

	userName := "  Player 1  "
	reqObj := request.CreateUserRequest{
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
	assert.Equal(t, http.StatusOK, rr.Code)

	var user model.User
	err = json.Unmarshal([]byte(rr.Body.String()), &user)

	assert.Equal(t, "Player 1", user.Name)
	assert.Equal(t, model.NoEstimate, user.Estimate)
	assert.NotNil(t, user.UserId)

	msg := testHub.Emitted[0]
	var userAddedWsEvent response.WsNewUser
	err = json.Unmarshal([]byte(msg), &userAddedWsEvent)
	assert.Equal(t, response.UserAddedEvent, userAddedWsEvent.Event)
	assert.Equal(t, user.Name, userAddedWsEvent.Name)

	userCountKey := fmt.Sprintf(db.Const.UserCount, session.SessionId)
	userCount, err := srv.Service().Store().GetInt(userCountKey)
	assert.Equal(t, 1, userCount)
}

func TestStartVoteEndpoint(t *testing.T) {
	session, _ := createSessionAndUsers(2, t)

	// force vote count to make sure it's reset
	voteCountKey := fmt.Sprintf(db.Const.VoteCount, session.SessionId)
	err := srv.Service().Store().Set(voteCountKey, 2)

	reqObj := request.StartVoteRequest{SessionId: session.SessionId}

	body, err := json.Marshal(reqObj)
	if err != nil {t.Error(err)}

	req, err := http.NewRequest("PUT", "/api/vote/start", bytes.NewBufferString(string(body)))
	if err != nil {t.Error(err)}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.StartVoteHttpHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t,  http.StatusOK, rr.Code)

	sessionStateKey := fmt.Sprintf(db.Const.SessionState, session.SessionId)
	sessionState, err := srv.Service().Store().GetInt(sessionStateKey)
	assert.Equal(t, model.Voting, sessionState)

	msg := testHub.Emitted[0]
	var voteStartedWsEvent response.WsVoteStarted
	err = json.Unmarshal([]byte(msg), &voteStartedWsEvent)
	assert.Equal(t, response.VoteStartedEVent, voteStartedWsEvent.Event)

	voteCount, err := srv.Service().Store().GetInt(voteCountKey)
	assert.Equal(t,0, voteCount)
}

func TestFinishVoteEndpoint(t *testing.T) {
	session, _ := createSessionAndUsers(2, t)

	reqObj := request.FinishVoteRequest{SessionId: session.SessionId}

	body, err := json.Marshal(reqObj)
	if err != nil {t.Error(err)}

	req, err := http.NewRequest("PUT", "/api/vote/finish", bytes.NewBufferString(string(body)))
	if err != nil {t.Error(err)}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.FinishVoteHttpHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t,  http.StatusOK, rr.Code)

	// get last event - it should be the vote results as we are done
	msg := testHub.Emitted[len(testHub.Emitted) - 1]
	var voteResultsWsEvent response.WsVoteFinished
	err = json.Unmarshal([]byte(msg), &voteResultsWsEvent)
	assert.Equal(t, response.VoteFinishedEvent, voteResultsWsEvent.Event)
}

func TestCastVoteForInactiveSession(t *testing.T) {
	session, users := createSessionAndUsers(2, t)

	_, err := srv.Service().CastVote(session.SessionId, users[0].UserId, "8")
	assert.NotNil(t, err)

	key := fmt.Sprintf(db.Const.VoteCount, session.SessionId)
	voteCount, err := srv.Service().Store().GetInt(key)
	assert.Equal(t, 0, voteCount)
}

func TestCastOneVote(t *testing.T) {
	userCount := 3
	session, users := createSessionAndUsers(userCount, t)
	err := srv.Service().StartVote(session.SessionId)
	if err != nil {t.Error(err)}

	clearHubEvents()

	vote, err := srv.Service().CastVote(session.SessionId, users[0].UserId, "8")
	if err != nil {t.Error(err)}
	assert.Equal(t, vote.UserId, users[0].UserId)

	voteCountKey := fmt.Sprintf(db.Const.VoteCount, session.SessionId)
	voteCount, err := srv.Service().Store().GetInt(voteCountKey)
	assert.Equal(t, 1, voteCount)

	storedUsers, err := srv.Service().Store().GetSessionUsers(session.SessionId)
	if err != nil {t.Error(err)}

	for i := 0; i < userCount; i++ {
		user := storedUsers[i]
		if user.UserId == users[0].UserId {
			assert.Equal(t, "8", user.Estimate)
			assert.Equal(t, true, user.Voted)
			break
		}
	}

	// vote not done so we should be getting the expected event
	msg := testHub.Emitted[0]
	var userVotedEvent response.WsUserVote
	err = json.Unmarshal([]byte(msg), &userVotedEvent)
	assert.Equal(t, response.UserVotedEVent, userVotedEvent.Event)
}

func TestCastAllVotes(t *testing.T) {
	numOfUsers := 3
	session, users := createSessionAndUsers(numOfUsers, t)
	err := srv.Service().StartVote(session.SessionId)
	if err != nil {t.Error(err)}

	for i := 0; i < numOfUsers; i++ {
		_, err := srv.Service().CastVote(session.SessionId, users[i].UserId, "3")
		if err != nil {t.Error(err)}
	}

	key := fmt.Sprintf(db.Const.VoteCount, session.SessionId)
	voteCount, err := srv.Service().Store().GetInt(key)
	assert.Equal(t, numOfUsers, voteCount)

	// get last event - it should be the vote results as we are done
	msg := testHub.Emitted[len(testHub.Emitted) - 1]
	var voteResultsWsEvent response.WsVoteFinished
	err = json.Unmarshal([]byte(msg), &voteResultsWsEvent)
	assert.Equal(t, response.VoteFinishedEvent, voteResultsWsEvent.Event)
	assert.Equal(t, numOfUsers, len(voteResultsWsEvent.Users))

	key = fmt.Sprintf(db.Const.SessionState, session.SessionId)
	sessionState, err := srv.Service().Store().GetInt(key)
	assert.Equal(t, sessionState, model.NotVoting)
}

/*
We want to make sure that all users in the session start with a "clean record"
 */
func TestNewVoteState(t *testing.T) {
	numOfUsers := 2
	session, users := createSessionAndUsers(numOfUsers, t)
	err := srv.Service().StartVote(session.SessionId)
	if err != nil {t.Error(err)}

	for i := 0; i < numOfUsers; i++ {
		_, err := srv.Service().CastVote(session.SessionId, users[i].UserId, "3")
		if err != nil {t.Error(err)}
	}

	err = srv.Service().StartVote(session.SessionId)
	if err != nil {t.Error(err)}

	usersForNewSession, err := srv.Service().Store().GetSessionUsers(session.SessionId)
	if err != nil {t.Error(err)}

	for i := 0; i < numOfUsers; i++ {
		user := usersForNewSession[i]
		assert.Equal(t, model.NoEstimate, user.Estimate)
		assert.Equal(t, false, user.Voted)
	}

}

func TestRepeatedVote(t *testing.T) {
	numOfUsers := 2
	session, users := createSessionAndUsers(numOfUsers, t)
	err := srv.Service().StartVote(session.SessionId)
	if err != nil {t.Error(err)}

	_, err = srv.Service().CastVote(session.SessionId, users[0].UserId, "3")
	if err != nil {t.Error(err)}
	_, err = srv.Service().CastVote(session.SessionId, users[0].UserId, "3")
	if err != nil {t.Error(err)}

	// vote count should still be 1 - one user voted
	key := fmt.Sprintf(db.Const.VoteCount, session.SessionId)
	voteCount, err := srv.Service().Store().GetInt(key)
	assert.Equal(t, 1, voteCount)

}

func TestGetUserById(t *testing.T) {
	_, users := createSessionAndUsers(1, t)
	createdUser := users[0]
	user, err := srv.Service().GetUser(createdUser.UserId)
	if err != nil {t.Error(err)}

	assert.Equal(t, createdUser.UserId, user.UserId)
	assert.Equal(t, createdUser.Name, user.Name)
	assert.Equal(t, false, user.Voted)
	assert.Equal(t, model.NoEstimate, user.Estimate)
}

func TestStateUserLeft(t *testing.T) {
	numOfUsers := 3
	session, users := createSessionAndUsers(numOfUsers, t)
	createdUser := users[0]

	err := srv.Service().RemoveUser(session.SessionId, createdUser.UserId)
	if err != nil {t.Error(err)}

	newNumOfUsers := numOfUsers - 1
	userIds, err := srv.Service().Store().GetSessionUserIds(session.SessionId)
	assert.Len(t, userIds, newNumOfUsers)

	key := fmt.Sprintf(db.Const.UserCount, session.SessionId)
	userCount, err := srv.Service().Store().GetInt(key)
	assert.Equal(t, newNumOfUsers, userCount)

	user, err := srv.Service().GetUser(createdUser.UserId)
	assert.Empty(t, user)
}

func TestVoteFinishedAfterUserLeft(t *testing.T) {
	numOfUsers := 3
	session, users := createSessionAndUsers(numOfUsers, t)
	flakeUser := users[0]

	err := srv.Service().StartVote(session.SessionId)
	if err != nil {t.Error(err)}

	/* Two users vote. Vote is not finished.
	 */
	_, err = srv.Service().CastVote(session.SessionId, users[1].UserId, "3")
	if err != nil {t.Error(err)}
	_, err = srv.Service().CastVote(session.SessionId, users[2].UserId, "8")
	if err != nil {t.Error(err)}

	// flake user does not vote and bails or gets disconnected
	err = srv.Service().RemoveUser(session.SessionId, flakeUser.UserId)
	if err != nil {t.Error(err)}

	// vote should be finished
	// get last event - it should be the vote results as we are done
	msg := testHub.Emitted[len(testHub.Emitted) - 1]
	var voteResultsWsEvent response.WsVoteFinished
	err = json.Unmarshal([]byte(msg), &voteResultsWsEvent)
	assert.Equal(t, response.VoteFinishedEvent, voteResultsWsEvent.Event)
	assert.Equal(t, numOfUsers - 1, len(voteResultsWsEvent.Users))

	key := fmt.Sprintf(db.Const.SessionState, session.SessionId)
	sessionState, err := srv.Service().Store().GetInt(key)
	assert.Equal(t, sessionState, model.NotVoting)
}

func TestSessionClearAfterAllUsersLeave(t *testing.T) {
	numOfUsers := 3
	session, users := createSessionAndUsers(numOfUsers, t)

	err := srv.Service().StartVote(session.SessionId)
	if err != nil {t.Error(err)}

	for i := 0; i < numOfUsers; i++ {
		err = srv.Service().RemoveUser(session.SessionId, users[i].UserId)
	}

	userCountKey := fmt.Sprintf(db.Const.UserCount, session.SessionId)
	_, err = srv.Service().Store().GetInt(userCountKey)
	assert.NotNil(t, err)

	sessionUserIds, err := srv.Service().Store().GetSessionUserIds(session.SessionId)
	if err != nil {t.Error(err)}
	assert.Empty(t, sessionUserIds)

	sessionStateKey := fmt.Sprintf(db.Const.SessionState, session.SessionId)
	_, err = srv.Service().Store().GetInt(sessionStateKey)
	assert.NotNil(t, err)

	voteCountKey := fmt.Sprintf(db.Const.VoteCount, session.SessionId)
	_, err = srv.Service().Store().GetInt(voteCountKey)
	assert.NotNil(t, err)
}

func TestEmptyUsername(t *testing.T) {
	session, err  := srv.Service().CreateSession()
	if err != nil {t.Errorf("Could not create session: %s", err)}

	_, err = srv.Service().CreateUser(session.SessionId, "")
	assert.NotNil(t, err)

	_, err = srv.Service().CreateUser(session.SessionId, "   ")
	assert.NotNil(t, err)

	_, err = srv.Service().CreateUser(session.SessionId, "  \n\n\t\t")
	assert.NotNil(t, err)
}

func TestDuplicateUsername(t *testing.T) {
	session, err  := srv.Service().CreateSession()
	if err != nil {t.Errorf("Could not create session: %s", err)}

	_, err = srv.Service().CreateUser(session.SessionId, "username")
	if err != nil {t.Error(err)}
	_, err = srv.Service().CreateUser(session.SessionId, "username")
	assert.NotNil(t, err)
}
