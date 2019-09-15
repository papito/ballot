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

type Hub struct {
	store *db.Store
	socketsMap  map[*glue.Socket]map[string]bool
	sessionsMap map[string]map[*glue.Socket]bool

	rwMutex sync.RWMutex
	glueSrv *glue.Server
}

func (p *Hub) Connect(store *db.Store) error {
	p.store = store
	p.socketsMap = map[*glue.Socket]map[string]bool{}
	p.sessionsMap = map[string]map[*glue.Socket]bool{}

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

func (p* Hub) Subscribe(sock *glue.Socket, session string) error {
	log.Printf("Subscribing socket %p to session %s", sock, session)
	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()

	_, ok := p.socketsMap[sock]
	if !ok {
		p.socketsMap[sock] = map[string]bool{}
	}
	p.socketsMap[sock][session] = true

	_, ok = p.sessionsMap[session]

	if !ok {
		p.sessionsMap[session] = map[*glue.Socket]bool{}
		err := p.store.SubConn.Subscribe(session)
		if err != nil {
			return err
		}
	}
	p.sessionsMap[session][sock] = true

	return nil
}

func (p * Hub) unsubscribeAll(sock *glue.Socket) error {
	log.Printf("Unsubscribing all from socket %p", sock)
	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()

	for session := range p.socketsMap[sock] {
		delete(p.sessionsMap[session], sock)
		if len(p.sessionsMap[session]) == 0 {
			delete(p.sessionsMap, session)
			err := p.store.SubConn.Unsubscribe(session)
			if err != nil {
				return err
			}
		}
	}
	delete(p.socketsMap, sock)

	return nil
}

func (p *Hub) Emit(session string, data string) error {
	log.Printf("EMIT. Session %s - %s", session, data)
	_, err := p.store.RedisConn.Do("PUBLISH", session, data)
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
	log.Printf("Handling socket %p", sock)

	sock.OnClose(func() {
		log.Printf("Socket %p closed", sock)
		err := p.unsubscribeAll(sock)
		if err != nil {
			log.Print(err)
		}

		// TODO: let other users in this session know the user left
	})

	sock.OnRead(func(data string) {
		log.Printf("Reading from socket %p: %s", sock, data)

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
		case "WATCH":
			log.Printf("WS. Watching session %s", sessionId)
			err := p.Subscribe(sock, sessionId)
			if err != nil {log.Print(err)}

			// get session state - voting, not voting
			key := fmt.Sprintf(db.Const.SessionState, sessionId)
			isVoting, err := redis.Int(p.store.RedisConn.Do("GET", key))

			sessionState := model.NotVoting
			if isVoting == 1 {
				sessionState = model.Voting
			}

			users, err := p.store.GetSessionUsers(sessionId)
			if err != nil {log.Print(err)}

			session := response.WsSession{
				Event: "WATCHING",
				SessionState: sessionState,
				Users: users,
			}

			data, err := json.Marshal(session)
			if err != nil {
				log.Println(err)
			}

			p.emitSocket(sock, string(data))

		case "START":
			err := p.Emit(sessionId, "{}")
			if err != nil {
				log.Print(err)
			}

		case "RESTART":
			err := p.Emit(sessionId, "{}")
			if err != nil {
				log.Print(err)
			}

		case "VOTE":
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
