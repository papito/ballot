package service

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/db"
	. "github.com/papito/ballot/ballot/hub"
	"github.com/papito/ballot/ballot/model"
	"github.com/papito/ballot/ballot/model/response"
	"log"
	"strings"
)

type Service struct  {
	store *db.Store
	hub IHub
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
	}

	service.store.Connect(config.RedisUrl)

	var err error
	/* Initiate the hub that connects sessions and sockets
	 */
	log.Println("Creating hub")
	err = service.hub.Connect(service.store)
	if err != nil {return service, fmt.Errorf("error creating hub: %s", err)}

	return service, nil
}

func (s *Service) Release() {
	log.Print("Releasing service resources")
	s.hub.Release()
	log.Print("Service done")
}

func (s *Service) Hub() IHub {
	return s.hub
}

func (s *Service) Store() *db.Store {
	return s.store
}

func (s *Service) CreateSession() (model.Session, error) {
	sessionUUID, _ := uuid.NewRandom()
	sessionId := sessionUUID.String()
	session := model.Session{SessionId: sessionId}

	key := fmt.Sprintf(db.Const.SessionVoting, sessionId)
	err := s.store.SetKey(key, model.NotVoting)
	if err != nil {return model.Session{}, fmt.Errorf("error saving data: %s", err)}

	key = fmt.Sprintf(db.Const.UserCount, sessionId)
	err = s.store.SetKey(key, 0)
	if err != nil {return model.Session{}, fmt.Errorf("error saving data: %s", err)}

	key = fmt.Sprintf(db.Const.VoteCount, sessionId)
	err = s.store.SetKey(key, 0)
	if err != nil {return model.Session{}, fmt.Errorf("error saving data: %s", err)}

	return session, nil
}

func (s *Service) CreateUser(sessionId string, userName string) (model.User, error) {
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
	err := s.store.SetHashKey(
		userKey,
		"name", user.Name,
		"id", user.UserId,
		"estimate", user.Estimate,)

	if err != nil {return model.User{}, err}

	sessionUserKey := fmt.Sprintf(db.Const.SessionUsers, sessionId)
	err = s.store.AddToSet(sessionUserKey, userId)
	if err != nil {return model.User{}, fmt.Errorf("error saving data. %v", err)}

	userCountKey  := fmt.Sprintf(db.Const.UserCount, sessionId)
	err = s.store.Incr(userCountKey, 1)
	if err != nil {return model.User{}, err}

	wsUser := response.WsNewUser{}
	wsUser.Event = response.UserAddedEvent
	wsUser.Name = user.Name
	wsUser.UserId = user.UserId
	wsUser.Estimate = user.Estimate

	wsResp, err := json.Marshal(wsUser)
	if err != nil {return model.User{}, fmt.Errorf("error marshalling data. %v", err)}

	err = s.hub.Emit(sessionId, string(wsResp))
	if err != nil {return model.User{}, fmt.Errorf("error emitting data. %v", err)}

	return user, nil
}

func (s *Service) CastVote(sessionId string, userId string, estimate int) (model.PendingVote, error) {
	log.Printf("Voting for session ID [%s] and user ID [%s]", sessionId, userId)

	// cannot vote on session that is inactive
	sessionKey := fmt.Sprintf(db.Const.SessionVoting, sessionId)
	sessionState, err := s.store.GetInt(sessionKey)

	if sessionState == model.NotVoting {
		return model.PendingVote{},
			fmt.Errorf("not voting yet for session [%s]", sessionId)
	}
	log.Printf("Voting for user ID [%s] with estimate [%d]", userId, estimate)

	userKey := fmt.Sprintf(db.Const.User, userId)
	err = s.store.SetHashKey(userKey, "estimate", estimate)
	if err != nil {
		return model.PendingVote{}, fmt.Errorf("error saving data. %v", err)
	}

	voteCountKey := fmt.Sprintf(db.Const.VoteCount, sessionId)
	err = s.store.Incr(voteCountKey, 1)
	if err != nil {
		return model.PendingVote{}, fmt.Errorf("error saving data. %v", err)
	}

	voteFinished, err := s.IsVoteFinished(sessionId)

	if voteFinished == false {
		wsUserVote := response.WsUserVote{
			Event:response.UserVotedEVent,
			UserId:userId,
		}

		data, err := json.Marshal(wsUserVote)
		if err != nil {return model.PendingVote{}, fmt.Errorf("error emitting data. %v", err)}

		err = s.hub.Emit(sessionId, string(data))
		if err != nil {return model.PendingVote{}, fmt.Errorf("error emitting data. %v", err)}
	} else {
		err = s.FinishVote(sessionId)
		if err != nil {return model.PendingVote{}, fmt.Errorf("error finishing vote. %v", err)}
	}

	vote := model.PendingVote{
		SessionId: sessionId,
		UserId: userId,
	}

	return vote, nil
}

func (s *Service) StartVote(sessionId string) error {
	log.Printf("Starting vote for session ID [%s]", sessionId)
	key := fmt.Sprintf(db.Const.SessionVoting, sessionId)
	err := s.store.SetKey(key, model.Voting)
	if err != nil {return fmt.Errorf("error saving data: %v", err)}

	key = fmt.Sprintf(db.Const.VoteCount, sessionId)
	err = s.store.SetKey(key, 0)
	if err != nil {return fmt.Errorf("error saving data: %v", err)}

	session := response.WsVoteStarted{
		Event: response.VoteStartedEVent,
	}

	data, err := json.Marshal(session)
	if err != nil {return fmt.Errorf("error marshalling data: %v", err)}

	err = s.hub.Emit(sessionId, string(data))
	if err != nil {return fmt.Errorf("error emitting data: %v", err)}

	return nil
}

func (s *Service) FinishVote(sessionId string) error {
	users, err := s.store.GetSessionUsers(sessionId)
	if err != nil {return fmt.Errorf("error getting users: %v", err)}

	session := response.WsVoteFinished{
		Event: response.VoteFinishedEvent,
		Users: users,
	}

	data, err := json.Marshal(session)
	if err != nil {return fmt.Errorf("error marshalling data: %v", err)}

	err = s.hub.Emit(sessionId, string(data))
	if err != nil {return fmt.Errorf("error emitting data: %v", err)}

	return nil
}

func (s * Service) IsVoteFinished(sessionId string) (bool, error) {
	voteCountKey := fmt.Sprintf(db.Const.VoteCount, sessionId)
	voteCount, err := s.store.GetInt(voteCountKey)
	if err != nil {
		return false, fmt.Errorf("error: %v", err)
	}

	userCountKey := fmt.Sprintf(db.Const.UserCount, sessionId)
	userCount, err := s.store.GetInt(userCountKey)
	if err != nil {
		return false, fmt.Errorf("error: %v", err)
	}

	return voteCount == userCount, nil
}