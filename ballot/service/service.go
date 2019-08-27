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
	if err != nil {return service, err}

	return service, nil
}

func (p *Service) CreateSession() (model.Session, error) {
	sessionUUID, _ := uuid.NewRandom()
	sessionId := sessionUUID.String()
	session := model.Session{SessionId: sessionId}

	key := fmt.Sprintf(db.Const.SessionVoting, sessionId)
	err := p.store.SetKey(key, model.NotVoting)
	if err != nil {return model.Session{}, err}

	return session, nil
}

func (p *Service) CreateUser(sessionId string, userName string) (model.User, error) {
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
	err := p.store.SetHashKey(
		userKey,
		"name", user.Name,
		"id", user.UserId,
		"estimate", user.Estimate,)
	if err != nil {return model.User{}, err}

	sessionUserKey := fmt.Sprintf(db.Const.SessionUsers, sessionId)
	err = p.store.AddToSet(sessionUserKey, userId)
	if err != nil {return model.User{}, err}

	wsUser := model.WsUser{}
	wsUser.Event = "USER_ADDED"
	wsUser.Name = user.Name
	wsUser.UserId = user.UserId
	wsUser.Estimate = user.Estimate

	wsResp, err := json.Marshal(wsUser)
	if err != nil {return model.User{}, err}

	err = p.hub.Emit(sessionId, string(wsResp))
	if err != nil {return model.User{}, err}

	return user, nil
}

func (p *Service) StartVote(sessionId string) error {
	log.Printf("Starting vote for session ID [%s]", sessionId)

	key := fmt.Sprintf(db.Const.SessionVoting, sessionId)
	err := p.store.SetKey(key, model.Voting)
	if err != nil {return err}

	session := model.WsSession{
		Event: "VOTING",
	}

	data, err := json.Marshal(session)
	if err != nil {return err}

	err = p.hub.Emit(sessionId, string(data))
	if err != nil {return err}

	return nil
}