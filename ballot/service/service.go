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
	"strconv"
	"strings"
	"time"
)

type Service struct  {
	store *db.Store
	hub IHub
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
		store: &db.Store{},
		hub: hubImpl,
		config: config,
	}

	service.store.Connect(config.RedisUrl)

	go func() {
		for {
			switch v := service.store.ServiceSubCon.Receive().(type) {
			case redis.Message:
				log.Printf(
					"Service subscriber connection received [%s] on channel [%s]", v.Data, v.Channel)

				service.processSubscriberEvent(v.Channel, string(v.Data))

			case error:
				panic(v)
			}
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
	err := p.store.Set(key, model.NotVoting)
	if err != nil {log.Printf("%+v", err); return model.Session{}, err}

	key = fmt.Sprintf(db.Const.UserCount, sessionId)
	err = p.store.Set(key, 0)
	if err != nil {log.Printf("%+v", err); return model.Session{}, err}

	key = fmt.Sprintf(db.Const.VoteCount, sessionId)
	err = p.store.Set(key, 0)
	if err != nil {log.Printf("%+v", err); return model.Session{}, err}

	return session, nil
}

func (p *Service) CreateUser(sessionId string, userName string) (model.User, error) {
	userName = strings.TrimSpace(userName)

	if len(userName) < 1 {
		valErr := errors.ValidationError{
			Field: "user.name",
			ErrorStr: "This field cannot be empty"}
		return model.User{}, valErr
	}

	// check for a duplicate user in this session
	currentUsers, err := p.store.GetSessionUsers(sessionId)
	if err != nil {log.Printf("%+v", err); return model.User{}, err}

	for _, user := range currentUsers {
		if strings.ToLower(user.Name) == strings.ToLower(userName) {
			valErr := errors.ValidationError{
				Field: "user.name",
				ErrorStr: "This user name already taken for this session"}
			return model.User{}, valErr
		}
	}

	userUUID, _ := uuid.NewRandom()
	userId := userUUID.String()
	joined := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	log.Printf("Creating user [%s] and id [%s]", userName, userId)

	user := model.User{
		UserId:   userId,
		Name:     userName,
		Estimate: model.NoEstimate,
	    Joined: joined,
	}

	userKey := fmt.Sprintf(db.Const.User, userId)
	err = p.store.SetHashKey(
		userKey,
		"name", user.Name,
		"id", user.UserId,
		"estimate", user.Estimate,
		"joined", user.Joined)

	if err != nil {log.Printf("%+v", err); return model.User{}, err}

	userCountKey  := fmt.Sprintf(db.Const.UserCount, sessionId)
	err = p.store.Incr(userCountKey, 1)
	if err != nil {log.Printf("%+v", err); return model.User{}, err}

	return user, nil
}

func (p *Service) RemoveUserFromSession(sessionId string, userId string) error {
	log.Printf("Removing user [%s] from session [%s]", userId, sessionId)

	sessionUserKey := fmt.Sprintf(db.Const.SessionUsers, sessionId)
	err := p.store.RemoveFromSet(sessionUserKey, userId)
	if err != nil {log.Printf("%+v", err); return err}

	userCountKey := fmt.Sprintf(db.Const.UserCount, sessionId)

	// if the session is dead, this will blow up
	_, err = p.store.GetInt(userCountKey)
	if err != nil {log.Printf("%+v", err); return err}

	err = p.store.Decr(userCountKey, 1)
	if err != nil {log.Printf("%+v", err); return err}

	voteFinished, err := p.IsVoteFinished(sessionId)
	if voteFinished == true {
		err = p.FinishVote(sessionId)
		if err != nil {log.Printf("%+v", err); return err}
	}

	return nil
}

func (p *Service) GetUser(userId string) (model.User, error) {
	user, err := p.store.GetUser(userId)
	if err != nil {log.Printf("%+v", err); return model.User{}, err}
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
		Event:response.UserVotedEVent,
		UserId:userId,
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
		UserId: userId,
	}

	return vote, nil
}

func (p *Service) StartVote(sessionId string) error {
	log.Printf("Starting vote for session ID [%s]", sessionId)
	key := fmt.Sprintf(db.Const.SessionState, sessionId)
	err := p.store.Set(key, model.Voting)
	if err != nil {log.Printf("%+v", err); return err}

	key = fmt.Sprintf(db.Const.VoteCount, sessionId)
	err = p.store.Set(key, 0)
	if err != nil {log.Printf("%+v", err); return err}

	// reset user state
	userIds, err := p.store.GetSessionUserIds(sessionId)

	for i := 0; i < len(userIds); i++ {
		userId := userIds[i]
		userKey := fmt.Sprintf(db.Const.User, userId)
		err = p.store.SetHashKey(userKey,"estimate", model.NoEstimate)
	}

	session := response.WsVoteStarted{
		Event: response.VoteStartedEVent,
	}

	data, err := json.Marshal(session)
	if err != nil {log.Printf("%+v", errorx.EnsureStackTrace(err)); return errorx.EnsureStackTrace(err)}

	err = p.hub.Emit(sessionId, string(data))
	if err != nil {log.Printf("%+v", errorx.EnsureStackTrace(err)); return errorx.EnsureStackTrace(err)}

	return nil
}

func (p *Service) FinishVote(sessionId string) error {
	key := fmt.Sprintf(db.Const.SessionState, sessionId)
	err := p.store.Set(key, model.NotVoting)
	if err != nil {log.Printf("%+v", err); return err}

	users, err := p.store.GetSessionUsers(sessionId)
	if err != nil {log.Printf("%+v", err); return err}

	session := response.WsVoteFinished{
		Event: response.VoteFinishedEvent,
		Users: users,
	}

	data, err := json.Marshal(session)
	if err != nil {log.Printf("%+v", errorx.EnsureStackTrace(err)); return errorx.EnsureStackTrace(err)}

	err = p.hub.Emit(sessionId, string(data))
	if err != nil {log.Printf("%+v", errorx.EnsureStackTrace(err)); return errorx.EnsureStackTrace(err)}

	return nil
}

func (p * Service) IsVoteFinished(sessionId string) (bool, error) {
	voteCountKey := fmt.Sprintf(db.Const.VoteCount, sessionId)
	voteCount, err := p.store.GetInt(voteCountKey)
	if err != nil {log.Printf("%+v", err); return false, err}

	userCountKey := fmt.Sprintf(db.Const.UserCount, sessionId)
	userCount, err := p.store.GetInt(userCountKey)
	if err != nil {log.Printf("%+v", err); return false, err}

	return voteCount == userCount, nil
}

func(p *Service) processSubscriberEvent(sessionId string, data string) {
	jsonData, err := jsonutil.GetJsonFromString(data)
	if err != nil {log.Printf("%+v", err)}

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
}