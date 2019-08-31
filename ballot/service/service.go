package service

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/db"
	"github.com/papito/ballot/ballot/hub"
	"github.com/papito/ballot/ballot/model"
	"log"
	"strings"
)

type Service struct  {
	store *db.Store
	hub *hub.Hub
}

func NewService(config config.Config) (Service, error) {
	service := Service{
		store: &db.Store{},
		hub: &hub.Hub{},
	}

	service.store.Connect(config.RedisUrl)

	var err error
	/* Initiate the hub that connects sessions and sockets
	 */
	log.Println("Creating hub")
	err = service.hub.Connect(config.RedisUrl)
	if err != nil {return service, fmt.Errorf("error creating hub: %s", err)}

	return service, nil
}

func (s *Service) Release() {
	log.Print("Releasing service resources")
	s.hub.Release()
	log.Print("Service done")
}

func (s *Service) Hub() *hub.Hub {
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

	return session, nil
}

func (s *Service) CreateUser(sessionId string, userName string) (model.User, error) {
	log.Printf("Creating user [%s]", userName)

	userName = strings.TrimSpace(userName)

	if len(userName) < 1 {
		valErr := model.ValidationError{
			Field: "name",
			ErrorStr: "This field cannot be empty"}
		return model.User{}, valErr
	}

	userUUID, _ := uuid.NewRandom()
	userId := userUUID.String()

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

	wsUser := model.WsUser{}
	wsUser.Event = "USER_ADDED"
	wsUser.Name = user.Name
	wsUser.UserId = user.UserId
	wsUser.Estimate = user.Estimate

	wsResp, err := json.Marshal(wsUser)
	if err != nil {return model.User{}, fmt.Errorf("error marshalling data. %v", err)}

	err = s.hub.Emit(sessionId, string(wsResp))
	if err != nil {return model.User{}, fmt.Errorf("error imitting data. %v", err)}

	return user, nil
}

func (s *Service) CastVote(sessionId string, userId string, estimate uint8) (model.Vote, error) {
	log.Printf("Voting for session ID [%s]", sessionId)

	// cannot vote on session that is inactive
	sessionKey := fmt.Sprintf(db.Const.SessionVoting, sessionId)
	sessionState, err := s.store.GetInt(sessionKey)

	if sessionState == model.NotVoting {
		return model.Vote{}, fmt.Errorf("not voting yet for session [%s]", sessionId)
	}

	log.Printf("Voting for user ID [%s] with estimate [%d]", userId, estimate)

	userKey := fmt.Sprintf("user:%s", userId)
	err = s.store.SetHashKey(userKey, "estimate", estimate)
	if err != nil {return model.Vote{}, fmt.Errorf("error saving data. %v", err)}

	wsUserVote := model.WsUserVote {
		Event:"USER_VOTED",
		UserId:userId,
		Estimate:estimate,
	}

	data, err := json.Marshal(wsUserVote)
	if err != nil {return model.Vote{}, fmt.Errorf("error imitting data. %v", err)}

	err = s.hub.Emit(sessionId, string(data))

	if err != nil {return model.Vote{}, fmt.Errorf("error imitting data. %v", err)}

	vote := model.Vote{
		SessionId: sessionId,
		UserId: userId,
		Estimate: estimate,
	}

	return vote, nil
}

func (s *Service) StartVote(sessionId string) error {
	log.Printf("Starting vote for session ID [%s]", sessionId)
	key := fmt.Sprintf(db.Const.SessionVoting, sessionId)
	err := s.store.SetKey(key, model.Voting)
	if err != nil {return fmt.Errorf("error saving data. %v", err)}

	session := model.WsSession{
		Event: "VOTING",
	}

	data, err := json.Marshal(session)
	if err != nil {return fmt.Errorf("error marshalling data. %v", err)}

	err = s.hub.Emit(sessionId, string(data))
	if err != nil {return fmt.Errorf("error imitting data. %v", err)}

	return nil
}