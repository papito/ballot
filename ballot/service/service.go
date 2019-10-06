package service

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/db"
	. "github.com/papito/ballot/ballot/hub"
	"github.com/papito/ballot/ballot/jsonutil"
	"github.com/papito/ballot/ballot/model"
	"github.com/papito/ballot/ballot/model/response"
	"log"
	"strings"
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

func NewService(config config.Config) (Service, error) {
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

			service.processSubscriberEvent(string(v.Data))

			case error:
				panic(v)
			}
		}
	}()

	/* Initiate the hub that connects sessions and sockets
	 */
	log.Println("Creating hub")
	err := service.hub.Connect(service.store)
	if err != nil {return service, fmt.Errorf("error creating hub: %s", err)}

	return service, nil
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
	if err != nil {return model.Session{}, fmt.Errorf("error saving data: %p", err)}

	key = fmt.Sprintf(db.Const.UserCount, sessionId)
	err = p.store.Set(key, 0)
	if err != nil {return model.Session{}, fmt.Errorf("error saving data: %p", err)}

	key = fmt.Sprintf(db.Const.VoteCount, sessionId)
	err = p.store.Set(key, 0)
	if err != nil {return model.Session{}, fmt.Errorf("error saving data: %p", err)}

	return session, nil
}

func (p *Service) CreateUser(sessionId string, userName string) (model.User, error) {
	userName = strings.TrimSpace(userName)

	if len(userName) < 1 {
		valErr := model.ValidationError{
			Field: "name",
			ErrorStr: "This field cannot be empty"}
		return model.User{}, valErr
	}

	userUUID, _ := uuid.NewRandom()
	userId := userUUID.String()
	log.Printf("Creating user [%s] and id [%s]", userName, userId)

	user := model.User{
		UserId:   userId,
		Name:     userName,
		Estimate: model.NoEstimate,
	}

	userKey := fmt.Sprintf(db.Const.User, userId)
	err := p.store.SetHashKey(
		userKey,
		"name", user.Name,
		"id", user.UserId,
		"estimate", user.Estimate,)

	if err != nil {return model.User{}, err}

	sessionUserKey := fmt.Sprintf(db.Const.SessionUsers, sessionId)
	err = p.store.AddToSet(sessionUserKey, userId)
	if err != nil {return model.User{}, fmt.Errorf("error saving data. %v", err)}

	userCountKey  := fmt.Sprintf(db.Const.UserCount, sessionId)
	err = p.store.Incr(userCountKey, 1)
	if err != nil {return model.User{}, err}

	wsUser := response.WsNewUser{}
	wsUser.Event = response.UserAddedEvent
	wsUser.Name = user.Name
	wsUser.UserId = user.UserId
	wsUser.Estimate = user.Estimate

	wsResp, err := json.Marshal(wsUser)
	if err != nil {return model.User{}, fmt.Errorf("error marshalling data. %v", err)}

	err = p.hub.Emit(sessionId, string(wsResp))
	if err != nil {return model.User{}, fmt.Errorf("error emitting data. %v", err)}

	return user, nil
}

func (p *Service) RemoveUser(sessionId string, userId string) error {
	userKey := fmt.Sprintf(db.Const.User, userId)

	err := p.store.Del(userKey)
	if err != nil {return err}

	sessionUserKey := fmt.Sprintf(db.Const.SessionUsers, sessionId)
	err = p.store.RemoveFromSet(sessionUserKey, userId)

	userCountKey  := fmt.Sprintf(db.Const.UserCount, sessionId)
	err = p.store.Decr(userCountKey, 1)
	if err != nil {return err}

	voteFinished, err := p.IsVoteFinished(sessionId)
	if voteFinished == true {
		err = p.FinishVote(sessionId)
		if err != nil {return fmt.Errorf("error finishing vote. %v", err)}
	}

	return nil
}

func (p *Service) GetUser(userId string) (model.User, error) {
	user, err := p.store.GetUser(userId)
	if err != nil {return model.User{}, err}
	return user, nil
}

func (p *Service) CastVote(sessionId string, userId string, estimate string) (model.PendingVote, error) {
	log.Printf("Voting for session ID [%p] and user ID [%p]", sessionId, userId)

	// cannot vote on session that is inactive
	sessionKey := fmt.Sprintf(db.Const.SessionState, sessionId)
	sessionState, err := p.store.GetInt(sessionKey)

	if sessionState == model.NotVoting {
		return model.PendingVote{},
			fmt.Errorf("not voting yet for session [%p]", sessionId)
	}
	log.Printf("Voting for user ID [%p] with estimate [%p]", userId, estimate)

	userKey := fmt.Sprintf(db.Const.User, userId)

	previousEstimate, err := p.store.GetHashKey(userKey, "estimate")
	if err != nil {
		return model.PendingVote{}, fmt.Errorf("error getting data. %v", err)
	}

	err = p.store.SetHashKey(userKey, "estimate", estimate)
	if err != nil {
		return model.PendingVote{}, fmt.Errorf("error saving data. %v", err)
	}

	// increment vote count IF this is a brand new vote for the user this session
	if previousEstimate == model.NoEstimate {
		voteCountKey := fmt.Sprintf(db.Const.VoteCount, sessionId)
		err = p.store.Incr(voteCountKey, 1)
		if err != nil {
			return model.PendingVote{}, fmt.Errorf("error saving data. %v", err)
		}
	}

	wsUserVote := response.WsUserVote{
		Event:response.UserVotedEVent,
		UserId:userId,
	}

	data, err := json.Marshal(wsUserVote)
	if err != nil {return model.PendingVote{}, fmt.Errorf("error emitting data. %v", err)}

	err = p.hub.Emit(sessionId, string(data))
	if err != nil {return model.PendingVote{}, fmt.Errorf("error emitting data. %v", err)}

	voteFinished, err := p.IsVoteFinished(sessionId)
	if voteFinished == true {
		err = p.FinishVote(sessionId)
		if err != nil {return model.PendingVote{}, fmt.Errorf("error finishing vote. %v", err)}
	}

	vote := model.PendingVote{
		SessionId: sessionId,
		UserId: userId,
	}

	return vote, nil
}

func (p *Service) StartVote(sessionId string) error {
	log.Printf("Starting vote for session ID [%p]", sessionId)
	key := fmt.Sprintf(db.Const.SessionState, sessionId)
	err := p.store.Set(key, model.Voting)
	if err != nil {return fmt.Errorf("error saving data: %v", err)}

	key = fmt.Sprintf(db.Const.VoteCount, sessionId)
	err = p.store.Set(key, 0)
	if err != nil {return fmt.Errorf("error saving data: %v", err)}

	// reset user state
	userIds, err := p.store.GetSessionUserIds(sessionId)

	for i := 0; i < len(userIds); i++ {
		userId := userIds[i]
		userKey := fmt.Sprintf(db.Const.User, userId)
		err = p.store.SetHashKey(userKey, "estimate", model.NoEstimate)
	}

	session := response.WsVoteStarted{
		Event: response.VoteStartedEVent,
	}

	data, err := json.Marshal(session)
	if err != nil {return fmt.Errorf("error marshalling data: %v", err)}

	err = p.hub.Emit(sessionId, string(data))
	if err != nil {return fmt.Errorf("error emitting data: %v", err)}

	return nil
}

func (p *Service) FinishVote(sessionId string) error {
	key := fmt.Sprintf(db.Const.SessionState, sessionId)
	err := p.store.Set(key, model.NotVoting)
	if err != nil {return fmt.Errorf("error saving data: %v", err)}

	users, err := p.store.GetSessionUsers(sessionId)
	if err != nil {return fmt.Errorf("error getting users: %v", err)}

	session := response.WsVoteFinished{
		Event: response.VoteFinishedEvent,
		Users: users,
	}

	data, err := json.Marshal(session)
	if err != nil {return fmt.Errorf("error marshalling data: %v", err)}

	err = p.hub.Emit(sessionId, string(data))
	if err != nil {return fmt.Errorf("error emitting data: %v", err)}

	return nil
}

func (p * Service) IsVoteFinished(sessionId string) (bool, error) {
	voteCountKey := fmt.Sprintf(db.Const.VoteCount, sessionId)
	voteCount, err := p.store.GetInt(voteCountKey)
	if err != nil {
		return false, fmt.Errorf("error: %v", err)
	}

	userCountKey := fmt.Sprintf(db.Const.UserCount, sessionId)
	userCount, err := p.store.GetInt(userCountKey)
	if err != nil {
		return false, fmt.Errorf("error: %v", err)
	}

	return voteCount == userCount, nil
}

func(p *Service) processSubscriberEvent(data string) {
	jsonData, err := jsonutil.GetJsonFromString(data)
	if err != nil {log.Print(err)}

	event, ok := jsonData["event"].(string)
	if !ok {
		log.Printf("no event found in: %v", data)
		return
	}

	sessionId, ok := jsonData["session_id"].(string)
	if !ok {
		log.Printf("no session_id found in: %v", data)
		return
	}

	userId, ok := jsonData["user_id"].(string)
	if !ok {
		log.Printf("no user_id found in: %v", data)
		return
	}

	if event == Event.UserLeft {
		err = p.RemoveUser(sessionId, userId)
	}

}