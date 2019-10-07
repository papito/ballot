package hub

import (
	"encoding/json"
	"fmt"
	"github.com/desertbit/glue"
	"github.com/gomodule/redigo/redis"
	"github.com/papito/ballot/ballot/db"
	"github.com/papito/ballot/ballot/jsonutil"
	"github.com/papito/ballot/ballot/model"
	"github.com/papito/ballot/ballot/model/response"
	"log"
	"net/http"
	"sync"
)

/* Modeled after https://github.com/hjr265/tonesa/blob/master/hub/hub.go */

type IHub interface {
	Connect(store *db.Store) error
	HandleWebSockets(url string)
	Emit(session string, data string) error
	EmitLocal(session string, data string)
	Release()
}

var Event = struct{
	Start    string
	Restart  string
	Vote     string
	Watch    string
	Watching string
	UserLeft string
}{
	"START",
	"RESTART",
	"VOTE",
	"WATCH",
	"WATCHING",
	"USER_LEFT",
}

type Hub struct {
	store *db.Store
	socketsMap  map[*glue.Socket]string
	sessionsMap map[string]map[*glue.Socket]bool
	userMap  map[*glue.Socket]string

	rwMutex sync.RWMutex
	glueSrv *glue.Server
}

func (p *Hub) Connect(store *db.Store) error {
	p.store = store
	p.socketsMap = map[*glue.Socket]string{}
	p.sessionsMap = map[string]map[*glue.Socket]bool{}
	p.userMap = map[*glue.Socket]string{}

	go func() {
		for {
			switch v := p.store.SubConn.Receive().(type) {
			case redis.Message:
				log.Printf(
					"Subscribe connection received [%s] on channel [%s]", v.Data, v.Channel)
				p.EmitLocal(v.Channel, string(v.Data))

			case error:
				panic(v)
			}
		}
	}()

	/* Create the Glue server */
	p.glueSrv = glue.NewServer(glue.Options{
		HTTPSocketType: glue.HTTPSocketTypeNone,
	})

	p.glueSrv.OnNewSocket(p.handleSocket)
	return nil
}

func (p* Hub) HandleWebSockets(url string) {
	http.Handle(url, p.glueSrv)
}

func (p* Hub) Release() {
	log.Print("Releasing Hub resources...")
	p.glueSrv.Release()
	log.Print("Hub done")
}

func (p* Hub) Subscribe(sock *glue.Socket, sessionId string) error {
	log.Printf("Subscribing socket %s to sessionId %s", sock.ID(), sessionId)
	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()

	p.socketsMap[sock] = sessionId

	_, ok := p.sessionsMap[sessionId]

	if !ok {
		p.sessionsMap[sessionId] = map[*glue.Socket]bool{}
		err := p.store.SubConn.Subscribe(sessionId)
		if err != nil {
			return err
		}
		err = p.store.ServiceSubCon.Subscribe(sessionId)
		if err != nil {
			return err
		}
	}
	p.sessionsMap[sessionId][sock] = true

	return nil
}

func (p* Hub) associateSocketWithUser(sock *glue.Socket, userId string) {
	log.Printf("Associating user [%s] with socket [%s]", userId, sock.ID())
	p.userMap[sock] = userId
}

func (p* Hub) disassociateSocketWithUser(sock *glue.Socket) {
	if userId, ok := p.userMap[sock]; ok {
		log.Printf("Disassociating user [%s] with socket [%s]", userId, sock.ID())
		delete(p.userMap, sock)
	}
}


func (p * Hub) unsubscribeAll(sock *glue.Socket) error {
	log.Printf("Unsubscribing all from socket %s", sock.ID())
	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()

	if sessionId, ok := p.socketsMap[sock]; ok {
		delete(p.sessionsMap[sessionId], sock)

		if len(p.sessionsMap[sessionId]) == 0 {
			delete(p.sessionsMap, sessionId)
			log.Printf("Unsubscribing from sessionId [%s] - no sockets connecting", sessionId)
			err := p.store.SubConn.Unsubscribe(sessionId)
			if err != nil {return err}
		}

		userId, _ := p.userMap[sock]
		event := response.WsUserLeftEvent{
			Event:     Event.UserLeft,
			SessionId: sessionId,
			UserId:    userId,
		}

		data, err := json.Marshal(event)
		if err != nil {log.Println(err)}
		err = p.Emit(sessionId, string(data))
		if err != nil {log.Println(err)}
	}
	delete(p.socketsMap, sock)
	p.disassociateSocketWithUser(sock)
	return nil
}

func (p *Hub) Emit(session string, data string) error {
	log.Printf("EMIT. Session %s - %s", session, data)
	c  := p.store.Pool.Get()
	defer p.store.Close(c)
	_, err := c.Do("PUBLISH", session, data)
	return err
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

func (p *Hub) emitSocket(sock *glue.Socket,data string) {
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
		if err != nil {log.Print(err)}
	})

	sock.OnRead(func(data string) {
		log.Printf("Reading from socket %s: %s", sock.ID(), data)

		c  := p.store.Pool.Get()
		defer p.store.Close(c)

		jsonData, err := jsonutil.GetJsonFromString(data)
		if err != nil {
			log.Print(err)
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
			if err != nil {log.Print(err)}

			if userId, ok := jsonData["user_id"].(string); ok {
				p.associateSocketWithUser(sock, userId)
			}

			// get session state - voting, not voting
			key := fmt.Sprintf(db.Const.SessionState, sessionId)
			isVoting, err := redis.Int(c.Do("GET", key))

			sessionState := model.NotVoting
			if isVoting == 1 {
				sessionState = model.Voting
			}

			users, err := p.store.GetSessionUsers(sessionId)
			if err != nil {log.Print(err)}

			session := response.WsSession{
				Event: Event.Watching,
				SessionState: sessionState,
				Users: users,
			}

			data, err := json.Marshal(session)
			if err != nil {
				log.Println(err)
			}

			p.emitSocket(sock, string(data))

		case Event.Start:
			err := p.Emit(sessionId, "{}")
			if err != nil {
				log.Print(err)
			}

		case Event.Restart:
			err := p.Emit(sessionId, "{}")
			if err != nil {
				log.Print(err)
			}

		case Event.Vote:
			err := p.Emit(sessionId, "{}")
			if err != nil {
				log.Print(err)
			}
		}
	})
}

/* A hub implementation used for testing. Poor man's mockery.
   Feel free to mock this.
   Or fake outrage.

   I can do this all day.
*/
// FIXME: mock wiring in tests instead of making the service aware of the test ENV
type VoidHub struct {
	Emitted []string
	LocalEmitted []string
}

func (p *VoidHub) Emit(session string, data string) error {
	p.Emitted = append(p.Emitted, data)
	return nil
}

func (p *VoidHub) EmitLocal(session string, data string) {
	p.LocalEmitted = append(p.LocalEmitted, data)
}

/* This VOID version resets the state */
func (p *VoidHub) Connect(db *db.Store) error {
	p.Emitted = p.Emitted[:0]
	p.LocalEmitted = p.LocalEmitted[:0]
	return nil
}

func (p *VoidHub) HandleWebSockets(url string) {return}
func (p *VoidHub) Release() {return}
