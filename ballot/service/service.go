package service

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/db"
	"github.com/papito/ballot/ballot/hub"
	"github.com/papito/ballot/ballot/models"
	"log"
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

//func (s service) CreateUser() (models.User, error) {
//
//}