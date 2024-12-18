package hub

import (
	"encoding/json"
	"fmt"
	"github.com/desertbit/glue"
	"github.com/gomodule/redigo/redis"
	"github.com/joomcode/errorx"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/db"
	"github.com/papito/ballot/ballot/jsonutil"
	"github.com/papito/ballot/ballot/model"
	"github.com/papito/ballot/ballot/model/response"
	"log"
	"net/http"
	"os"
	"sync"
)

/* Modeled after https://github.com/hjr265/tonesa/blob/master/hub/hub.go */

type IHub interface {
	Connect(store *db.Store)
	HandleWebSockets(url string)
	Emit(session string, data string) error
	EmitLocal(session string, data string)
	Release()
}

var Event = struct {
	Start        string
	Restart      string
	Vote         string
	Watch        string
	Watching     string
	UserLeft     string
	ObserverLeft string
}{
	"START",
	"RESTART",
	"VOTE",
	"WATCH",
	"WATCHING",
	"USER_LEFT",
	"OBSERVER_LEFT",
}

type Hub struct {
	store       *db.Store
	socketsMap  map[*glue.Socket]string
	sessionsMap map[string]map[*glue.Socket]bool
	userMap     map[*glue.Socket]string

	rwMutex sync.RWMutex
	glueSrv *glue.Server
}

func (p *Hub) Connect(store *db.Store) {
	p.store = store
	p.socketsMap = map[*glue.Socket]string{}
	p.sessionsMap = map[string]map[*glue.Socket]bool{}
	p.userMap = map[*glue.Socket]string{}

	go func() {
		for {
			p.store.SubConn = redis.PubSubConn{Conn: p.store.Pool.Get()}

			for p.store.SubConn.Conn.Err() == nil {
				switch v := p.store.SubConn.Receive().(type) {
				case redis.Message:
					log.Printf(
						"Subscribe connection received [%s] on channel [%s]", v.Data, v.Channel)
					p.EmitLocal(v.Channel, string(v.Data))
				case error:
					log.Print("PubSub err...or?")
					fmt.Printf(p.store.SubConn.Conn.Err().Error())
				}
			}
			_ = p.store.SubConn.Close()

			log.Print("Heroically getting a new connection!")
			p.store.SubConn = redis.PubSubConn{Conn: p.store.Pool.Get()}
		}
	}()

	/* Create the Glue server */
	env := os.Getenv("ENV")
	p.glueSrv = glue.NewServer(glue.Options{
		HTTPSocketType: glue.HTTPSocketTypeNone,
		CheckOrigin: func(r *http.Request) bool {
			return env == config.DEV
		},
	})

	p.glueSrv.OnNewSocket(p.handleSocket)
}

func (p *Hub) HandleWebSockets(url string) {
	http.Handle(url, p.glueSrv)
}

func (p *Hub) Release() {
	log.Print("Releasing Hub resources...")
	p.glueSrv.Release()
	log.Print("Hub done")
}

func (p *Hub) Subscribe(sock *glue.Socket, sessionId string) error {
	log.Printf("Subscribing socket %s to sessionId %s", sock.ID(), sessionId)
	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()

	p.socketsMap[sock] = sessionId

	_, ok := p.sessionsMap[sessionId]

	if !ok {
		p.sessionsMap[sessionId] = map[*glue.Socket]bool{}
		err := p.store.SubConn.Subscribe(sessionId)
		if err != nil {
			return errorx.EnsureStackTrace(err)
		}
		err = p.store.ServiceSubCon.Subscribe(sessionId)
		if err != nil {
			return errorx.EnsureStackTrace(err)
		}
	}
	p.sessionsMap[sessionId][sock] = true

	return nil
}

func (p *Hub) associateSocketWithUser(sock *glue.Socket, userId string) {
	log.Printf("Associating user [%s] with socket [%s]", userId, sock.ID())
	p.userMap[sock] = userId
}

func (p *Hub) disassociateSocketWithUser(sock *glue.Socket) {
	if userId, ok := p.userMap[sock]; ok {
		log.Printf("Disassociating user [%s] with socket [%s]", userId, sock.ID())
		delete(p.userMap, sock)
	}
}

func (p *Hub) unsubscribeAll(sock *glue.Socket) error {
	log.Printf("Unsubscribing all from socket %s", sock.ID())
	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()

	if sessionId, ok := p.socketsMap[sock]; ok {
		delete(p.sessionsMap[sessionId], sock)

		if len(p.sessionsMap[sessionId]) == 0 {
			delete(p.sessionsMap, sessionId)
			log.Printf("Unsubscribing from sessionId [%s] - no sockets connecting", sessionId)
			err := p.store.SubConn.Unsubscribe(sessionId)
			if err != nil {
				return errorx.EnsureStackTrace(err)
			}
		}

		userId, _ := p.userMap[sock]

		user, err := p.store.GetUser(userId)
		if err != nil {
			return errorx.EnsureStackTrace(err)
		}

		var data []byte

		if user.IsObserver {
			event := response.WsObserverLeftEvent{
				Event:     Event.ObserverLeft,
				SessionId: sessionId,
				UserId:    userId,
			}
			data, err = json.Marshal(event)
		} else {
			event := response.WsUserLeftEvent{
				Event:     Event.UserLeft,
				SessionId: sessionId,
				UserId:    userId,
			}
			data, err = json.Marshal(event)
		}

		if err != nil {
			return errorx.EnsureStackTrace(err)
		}
		err = p.Emit(sessionId, string(data))
		if err != nil {
			return errorx.EnsureStackTrace(err)
		}
	}
	delete(p.socketsMap, sock)
	p.disassociateSocketWithUser(sock)
	return nil
}

func (p *Hub) Emit(session string, data string) error {
	log.Printf("EMIT. Session %s - %s", session, data)
	c := p.store.Pool.Get()
	defer p.store.Close(c)
	_, err := c.Do("PUBLISH", session, data)

	if err != nil {
		return errorx.EnsureStackTrace(err)
	} else {
		return nil
	}
}

func (p *Hub) EmitLocal(session string, data string) {
	log.Printf("EMIT LOCAL. Session %s - %s", session, data)
	p.rwMutex.RLock()
	defer p.rwMutex.RUnlock()

	// write to socketsMap interested in this session
	for socket := range p.sessionsMap[session] {
		socket.Write(data)
	}
}

func (p *Hub) emitSocket(sock *glue.Socket, data string) {
	log.Printf("EMIT SOCKET. Socket %s - %s", sock.ID(), data)
	p.rwMutex.RLock()
	defer p.rwMutex.RUnlock()

	sock.Write(data)
}

func (p *Hub) handleSocket(sock *glue.Socket) {
	log.Printf("Handling socket %s", sock.ID())

	sock.OnClose(func() {
		log.Printf("Socket %s closed", sock.ID())

		err := p.unsubscribeAll(sock)
		if err != nil {
			log.Printf("%+v", errorx.EnsureStackTrace(err))
			return
		}
	})

	sock.OnRead(func(data string) {
		log.Printf("Reading from socket %s: %s", sock.ID(), data)

		c := p.store.Pool.Get()
		defer p.store.Close(c)

		jsonData, err := jsonutil.GetJsonFromString(data)
		if err != nil {
			log.Printf("%+v", err)
			return
		}

		var sessionId = jsonData["session_id"].(string)
		var action = jsonData["action"].(string)

		switch action {
		/*
		   Emit the WATCHING event, as well as a list of current users in this session
		*/
		case Event.Watch:
			log.Printf("WS. Watching session %s", sessionId)
			err := p.Subscribe(sock, sessionId)
			if err != nil {
				log.Printf("%+v", err)
				return
			}

			// get session state - voting, not voting
			key := fmt.Sprintf(db.Const.SessionState, sessionId)
			isVoting, err := redis.Int(c.Do("GET", key))
			if err != nil {
				log.Printf("%+v", err)
				return
			}

			sessionState := model.NotVoting
			if isVoting == 1 {
				sessionState = model.Voting
			}

			if userId, ok := jsonData["user_id"].(string); ok {
				p.associateSocketWithUser(sock, userId)

				user, err := p.store.GetUser(userId)
				if err != nil {
					log.Printf("%+v", err)
					return
				}

				if user.IsObserver {
					sessionObserverKey := fmt.Sprintf(db.Const.SessionObservers, sessionId)
					log.Printf("Adding observer [%s] to session [%s]", userId, sessionId)
					err = p.store.AddToSet(sessionObserverKey, userId)
					if err != nil {
						return
					}

				} else {
					sessionUserKey := fmt.Sprintf(db.Const.SessionUsers, sessionId)
					log.Printf("Adding voter [%s] to session [%s]", userId, sessionId)
					err = p.store.AddToSet(sessionUserKey, userId)
					if err != nil {
						return
					}
				}

				wsUser := response.WsNewUser{}
				if user.IsObserver {
					wsUser.Event = response.ObserverAddedEvent

				} else {
					wsUser.Event = response.UserAddedEvent
				}

				wsUser.Name = user.Name
				wsUser.UserId = user.UserId
				wsUser.Joined = user.Joined
				wsUser.Voted = user.Voted
				wsUser.IsObserver = user.IsObserver
				wsUser.IsAdmin = user.IsAdmin

				// only expose votes when not voting
				if sessionState == model.NotVoting {
					wsUser.Estimate = user.Estimate
				}

				wsResp, err := json.Marshal(wsUser)
				if err != nil {
					log.Printf("%+v", err)
					return
				}

				err = p.Emit(sessionId, string(wsResp))
				if err != nil {
					log.Printf("%+v", err)
					return
				}
			}

			users, err := p.store.GetSessionVoters(sessionId)
			if err != nil {
				log.Printf("%+v", err)
				return
			}

			// null out estimates if still voting
			for idx := range users {
				if sessionState == model.Voting {
					users[idx].Estimate = model.NoEstimate
				}
			}

			observers, err := p.store.GetSessionObservers(sessionId)
			if err != nil {
				log.Printf("%+v", err)
				return
			}

			key = fmt.Sprintf(db.Const.Tally, sessionId)
			tally, err := p.store.GetStr(key)
			if err != nil {
				log.Printf("%+v", err)
			}

			session := response.WsSession{
				Event:        Event.Watching,
				SessionState: sessionState,
				Users:        users,
				Observers:    observers,
				Tally:        tally,
			}

			data, err := json.Marshal(session)
			if err != nil {
				log.Printf("%+v", err)
				return
			}

			p.emitSocket(sock, string(data))

		case Event.Start:
			err := p.Emit(sessionId, "{}")
			if err != nil {
				log.Printf("%+v", err)
				return
			}

		case Event.Restart:
			err := p.Emit(sessionId, "{}")
			if err != nil {
				log.Printf("%+v", err)
				return
			}

		case Event.Vote:
			err := p.Emit(sessionId, "{}")
			if err != nil {
				log.Printf("%+v", err)
				return
			}
		}
	})
}

// A hub implementation used for testing.

type VoidHub struct {
	Emitted      []string
	LocalEmitted []string
}

func (p *VoidHub) Emit(_ string, data string) error {
	p.Emitted = append(p.Emitted, data)
	return nil
}

func (p *VoidHub) EmitLocal(_ string, data string) {
	p.LocalEmitted = append(p.LocalEmitted, data)
}

// Connect This VOID version resets the state
func (p *VoidHub) Connect(_ *db.Store) {
	p.Emitted = p.Emitted[:0]
	p.LocalEmitted = p.LocalEmitted[:0]
}

func (p *VoidHub) HandleWebSockets(_ string) { return }
func (p *VoidHub) Release()                  { return }
