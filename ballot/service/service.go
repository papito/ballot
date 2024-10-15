/*
 * The MIT License
 *
 * Copyright (c) 2020,  Andrei Taranchenko
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

package service

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/joomcode/errorx"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/db"
	"github.com/papito/ballot/ballot/errors"
	. "github.com/papito/ballot/ballot/hub"
	"github.com/papito/ballot/ballot/jsonutil"
	"github.com/papito/ballot/ballot/model"
	"github.com/papito/ballot/ballot/model/response"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	store  *db.Store
	hub    IHub
	config config.Config
}

func getHub(config config.Config) IHub {
	var hubImpl IHub = nil
	if config.Environment == "test" {
		hubImpl = &VoidHub{}
	} else {
		hubImpl = &Hub{}
	}

	return hubImpl
}

func NewService(config config.Config) Service {
	hubImpl := getHub(config)
	service := Service{
		store:  &db.Store{},
		hub:    hubImpl,
		config: config,
	}

	service.store.Pool = db.NewPool(config.RedisUrl)

	go func() {
		for {
			service.store.ServiceSubCon = redis.PubSubConn{Conn: service.store.Pool.Get()}

			for service.store.ServiceSubCon.Conn.Err() == nil {
				switch v := service.store.ServiceSubCon.Receive().(type) {
				case redis.Message:
					log.Printf(
						"Service subscriber connection received [%s] on channel [%s]", v.Data, v.Channel)
					service.processSubscriberEvent(v.Channel, string(v.Data))
				case error:
					log.Print("PubSub err...or?")
					fmt.Printf(service.store.ServiceSubCon.Conn.Err().Error())
				}
			}
			_ = service.store.ServiceSubCon.Close()

			log.Print("Heroically getting a new connection!")
			service.store.ServiceSubCon = redis.PubSubConn{Conn: service.store.Pool.Get()}
		}
	}()

	/* Initiate the hub that connects sessions and sockets
	 */
	log.Println("Creating hub")
	service.hub.Connect(service.store)

	return service
}

func (p *Service) Release() {
	log.Print("Releasing service resources")
	p.hub.Release()
	log.Print("Service done")
}

func (p *Service) Hub() IHub {
	return p.hub
}

func (p *Service) Config() config.Config {
	return p.config
}

func (p *Service) Store() *db.Store {
	return p.store
}

func (p *Service) CreateSession() (model.Session, error) {
	sessionUUID, _ := uuid.NewRandom()
	sessionId := sessionUUID.String()
	session := model.Session{SessionId: sessionId}

	key := fmt.Sprintf(db.Const.SessionState, sessionId)
	println("1")
	err := p.store.Set(key, model.NotVoting)
	println("2")
	if err != nil {
		log.Printf("%+v", err)
		return model.Session{}, err
	}

	key = fmt.Sprintf(db.Const.VoteCount, sessionId)
	err = p.store.Set(key, 0)
	if err != nil {
		log.Printf("%+v", err)
		return model.Session{}, err
	}

	return session, nil
}

func (p *Service) CreateUser(sessionId string, userName string, isAdmin bool, isObserver bool) (model.User, error) {
	userName = strings.TrimSpace(userName)

	if len(userName) < 1 {
		valErr := errors.ValidationError{
			Field:    "user.name",
			ErrorStr: "This field cannot be empty"}
		return model.User{}, valErr
	}

	// check for a duplicate user in this session
	currentUsers, err := p.store.GetSessionVoters(sessionId)
	if err != nil {
		log.Printf("%+v", err)
		return model.User{}, err
	}

	for _, user := range currentUsers {
		if strings.ToLower(user.Name) == strings.ToLower(userName) {
			valErr := errors.ValidationError{
				Field:    "user.name",
				ErrorStr: "This user name already taken for this session"}
			return model.User{}, valErr
		}
	}

	userUUID, _ := uuid.NewRandom()
	userId := userUUID.String()
	joined := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	log.Printf("Creating user [%s] and id [%s]", userName, userId)

	user := model.User{
		UserId:     userId,
		Name:       userName,
		Estimate:   model.NoEstimate,
		Joined:     joined,
		IsObserver: isObserver,
		IsAdmin:    isAdmin,
	}

	userKey := fmt.Sprintf(db.Const.User, userId)
	err = p.store.SetHashKey(
		userKey,
		"name", user.Name,
		"id", user.UserId,
		"estimate", user.Estimate,
		"joined", user.Joined,
		"is_admin", user.IsAdmin,
		"is_observer", isObserver)

	if err != nil {
		log.Printf("%+v", err)
		return model.User{}, err
	}

	return user, nil
}

func (p *Service) AddUserToSession(sessionId string, userId string) error {
	sessionUserKey := fmt.Sprintf(db.Const.SessionUsers, sessionId)
	log.Printf("Adding user [%s] to session [%s]", userId, sessionId)
	err := p.store.AddToSet(sessionUserKey, userId)
	if err != nil {
		return err
	}

	return nil
}

func (p *Service) RemoveUserFromSession(sessionId string, userId string) error {
	log.Printf("Removing user [%s] from session [%s]", userId, sessionId)

	sessionUserKey := fmt.Sprintf(db.Const.SessionUsers, sessionId)
	err := p.store.RemoveFromSet(sessionUserKey, userId)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	return nil
}

func (p *Service) RemoveObserver(sessionId string, userId string) error {
	log.Printf("Removing observer [%s] from session [%s]", userId, sessionId)

	sessionObserverKey := fmt.Sprintf(db.Const.SessionObservers, sessionId)
	err := p.store.RemoveFromSet(sessionObserverKey, userId)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	return nil
}

func (p *Service) GetUser(userId string) (model.User, error) {
	user, err := p.store.GetUser(userId)
	if err != nil {
		log.Printf("%+v", err)
		return model.User{}, err
	}
	return user, nil
}

func (p *Service) CastVote(sessionId string, userId string, estimate string) (model.PendingVote, error) {
	log.Printf("Voting for session ID [%s] and user ID [%s]", sessionId, userId)

	// cannot vote on session that is inactive
	sessionKey := fmt.Sprintf(db.Const.SessionState, sessionId)
	sessionState, err := p.store.GetInt(sessionKey)

	if sessionState == model.NotVoting {
		return model.PendingVote{},
			fmt.Errorf("not voting yet for session [%s]", sessionId)
	}
	log.Printf("Voting for user ID [%s] with estimate [%s]", userId, estimate)

	userKey := fmt.Sprintf(db.Const.User, userId)

	previousEstimate, err := p.store.GetHashKey(userKey, "estimate")
	if err != nil {
		log.Printf("%+v", err)
		return model.PendingVote{}, err
	}

	err = p.store.SetHashKey(userKey, "estimate", estimate)
	if err != nil {
		log.Printf("%+v", err)
		return model.PendingVote{}, err
	}

	// increment vote count IF this is a brand new vote for the user this session
	if previousEstimate == model.NoEstimate {
		voteCountKey := fmt.Sprintf(db.Const.VoteCount, sessionId)
		err = p.store.Incr(voteCountKey, 1)
		if err != nil {
			log.Printf("%+v", err)
			return model.PendingVote{}, err
		}
	}

	wsUserVote := response.WsUserVote{
		Event:  response.UserVotedEVent,
		UserId: userId,
	}

	data, err := json.Marshal(wsUserVote)
	if err != nil {
		log.Printf("%+v", errorx.EnsureStackTrace(err))
		return model.PendingVote{}, errorx.EnsureStackTrace(err)
	}

	err = p.hub.Emit(sessionId, string(data))
	if err != nil {
		log.Printf("%+v", errorx.EnsureStackTrace(err))
		return model.PendingVote{}, errorx.EnsureStackTrace(err)
	}

	voteFinished, err := p.IsVoteFinished(sessionId)
	if voteFinished == true {
		err = p.FinishVote(sessionId)
		if err != nil {
			log.Printf("%+v", err)
			return model.PendingVote{}, err
		}
	}

	vote := model.PendingVote{
		SessionId: sessionId,
		UserId:    userId,
	}

	return vote, nil
}

func (p *Service) StartVote(sessionId string) error {
	log.Printf("Starting vote for session ID [%s]", sessionId)
	key := fmt.Sprintf(db.Const.SessionState, sessionId)
	err := p.store.Set(key, model.Voting)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	key = fmt.Sprintf(db.Const.VoteCount, sessionId)
	err = p.store.Set(key, 0)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	key = fmt.Sprintf(db.Const.Tally, sessionId)
	err = p.store.Set(key, "")
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// reset user state
	userIds, err := p.store.GetSessionVoterIds(sessionId)

	for i := 0; i < len(userIds); i++ {
		userId := userIds[i]
		userKey := fmt.Sprintf(db.Const.User, userId)
		err = p.store.SetHashKey(userKey, "estimate", model.NoEstimate)
	}

	session := response.WsVoteStarted{
		Event: response.VoteStartedEVent,
	}

	data, err := json.Marshal(session)
	if err != nil {
		log.Printf("%+v", errorx.EnsureStackTrace(err))
		return errorx.EnsureStackTrace(err)
	}

	err = p.hub.Emit(sessionId, string(data))
	if err != nil {
		log.Printf("%+v", errorx.EnsureStackTrace(err))
		return errorx.EnsureStackTrace(err)
	}

	return nil
}

func (p *Service) FinishVote(sessionId string) error {
	key := fmt.Sprintf(db.Const.SessionState, sessionId)
	err := p.store.Set(key, model.NotVoting)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	users, err := p.store.GetSessionVoters(sessionId)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	estimates := make([]string, 0)

	for _, user := range users {
		estimates = append(estimates, user.Estimate)
	}

	tally, err := p.GetVoteResult(estimates)
	if err != nil {
		return err
	}

	key = fmt.Sprintf(db.Const.Tally, sessionId)
	err = p.store.Set(key, tally)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	session := response.WsVoteFinished{
		Event: response.VoteFinishedEvent,
		Users: users,
		Tally: tally,
	}

	data, err := json.Marshal(session)
	if err != nil {
		log.Printf("%+v", errorx.EnsureStackTrace(err))
		return errorx.EnsureStackTrace(err)
	}

	err = p.hub.Emit(sessionId, string(data))
	if err != nil {
		log.Printf("%+v", errorx.EnsureStackTrace(err))
		return errorx.EnsureStackTrace(err)
	}

	return nil
}

func (p *Service) IsVoteFinished(sessionId string) (bool, error) {
	voteCountKey := fmt.Sprintf(db.Const.VoteCount, sessionId)
	voteCount, err := p.store.GetInt(voteCountKey)
	if err != nil {
		log.Printf("%+v", err)
		return false, err
	}

	userCount, err := p.store.GetSetLength(fmt.Sprintf(db.Const.SessionUsers, sessionId))
	if err != nil {
		log.Printf("%+v", err)
		return false, err
	}

	return voteCount == userCount, nil
}

func (p *Service) processSubscriberEvent(sessionId string, data string) {
	jsonData, err := jsonutil.GetJsonFromString(data)
	if err != nil {
		log.Printf("%+v", err)
	}

	event, ok := jsonData["event"].(string)
	if !ok {
		log.Printf("no event found in: %v", data)
		return
	}

	if event == Event.UserLeft {
		userId, ok := jsonData["user_id"].(string)
		if !ok {
			log.Printf("no user_id found in: %v", data)
			return
		}

		err = p.RemoveUserFromSession(sessionId, userId)
		if err != nil {
			log.Printf("Error removnig user: %+v", err)
		}
	}

	if event == Event.ObserverLeft {
		userId, ok := jsonData["user_id"].(string)
		if !ok {
			log.Printf("no user_id found in: %v", data)
			return
		}

		err = p.RemoveObserver(sessionId, userId)
		if err != nil {
			log.Printf("Error removing observer: %+v", err)
		}
	}
}

func (p *Service) GetVoteResult(sInputs []string) (string, error) {
	if len(sInputs) == 0 {
		return "?", nil
	}

	inputs := make([]int, 0)

	var counts = map[int]int{}

	// convert to ints first
	for _, sInput := range sInputs {
		if sInput == "" || sInput == "?" {
			continue
		}

		iInput, err := strconv.Atoi(sInput)
		if err != nil {
			log.Printf("%+v", errorx.EnsureStackTrace(err))
			return "", err
		}
		inputs = append(inputs, iInput)
	}

	if len(inputs) == 0 {
		return "?", nil
	}

	// group the same estimates, key is the estimate, value is the count
	for _, input := range inputs {
		counts[input] = counts[input] + 1
	}

	counters := make([]int, len(counts))

	for _, count := range counts {
		counters = append(counters, count)
	}
	sort.Ints(counters)

	// and the highest count of same estimates is?
	frequentEstimateCount := counters[len(counters)-1]

	// get the estimates with the same highest count, could be be 1 or more
	mostFreqEstimates := make([]int, 0)

	for estimate, count := range counts {
		if count == frequentEstimateCount {
			mostFreqEstimates = append(mostFreqEstimates, estimate)
		}
	}

	sort.Ints(mostFreqEstimates)

	// the end result is the range of the most frequent estimates
	if len(mostFreqEstimates) == 1 {
		return fmt.Sprintf("%d", mostFreqEstimates[0]), nil
	} else if len(mostFreqEstimates) > 1 {
		return fmt.Sprintf("%d - %d", mostFreqEstimates[0], mostFreqEstimates[len(mostFreqEstimates)-1]), nil
	} else {
		return "?", nil
	}
}
