package service

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/db"
	"github.com/papito/ballot/ballot/hub"
	"github.com/papito/ballot/ballot/models"
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
	if err != nil {
		return service, fmt.Errorf("error creating hub: %s", err)
	}

	return service, nil
}

func (s *Service) CreateSession() (models.Session, error) {
	sessionUUID, _ := uuid.NewRandom()
	sessionId := sessionUUID.String()
	session := models.Session{SessionId: sessionId}

	key := fmt.Sprintf(db.Const.SessionVoting, sessionId)
	err := s.store.SetKey(key, models.NotVoting)

	if err != nil {
		return models.Session{}, fmt.Errorf("error saving data: %s", err)
	}

	return session, nil
}

func (s *Service) CreateUser(sessionId string, userName string) (models.User, error) {
	log.Printf("Creating user [%s]", userName)

	userName = strings.TrimSpace(userName)

	if len(userName) < 1 {
		valErr := models.ValidationError{
			Field: "name",
			ErrorStr: "This field cannot be empty"}


		return models.User{}, valErr
	}

	userUUID, _ := uuid.NewRandom()
	userId := userUUID.String()

	user := models.User{
		UserId:   userId,
		Name:     userName,
		Estimate: models.NoEstimate,
	}

	userKey := fmt.Sprintf(db.Const.User, userId)
	err := s.store.SetHashKey(
		userKey,
		"name", user.Name,
		"id", user.UserId,
		"estimate", user.Estimate,)

	if err != nil {
		return models.User{}, err
	}

	sessionUserKey := fmt.Sprintf(db.Const.SessionUsers, sessionId)
	err = s.store.AddToSet(sessionUserKey, userId)

	if err != nil {
		log.Printf("Error saving data: %s", err)
		return models.User{}, err
	}

	type WsUser struct {
		models.User
		Event  string `json:"event"`
	}

	wsUser := WsUser{}
	wsUser.Event = "USER_ADDED"
	wsUser.Name = user.Name
	wsUser.UserId = user.UserId
	wsUser.Estimate = user.Estimate

	wsResp, err := json.Marshal(wsUser)

	if err != nil {
		log.Println(err)
	}

	err = s.hub.Emit(sessionId, string(wsResp))

	if err != nil {
		log.Println(err)
	}

	return user, nil
}
