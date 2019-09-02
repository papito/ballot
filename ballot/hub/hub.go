package hub

import (
	"encoding/json"
	"fmt"
	"github.com/desertbit/glue"
	"github.com/gomodule/redigo/redis"
	"github.com/papito/ballot/ballot/jsonutil"
	"github.com/papito/ballot/ballot/model"
	"log"
	"net/http"
	"strconv"
	"sync"
)

/* Modeled after https://github.com/hjr265/tonesa/blob/master/hub/hub.go */

type IHub interface {
	Connect(url string) error
	HandleWebSockets(url string)
	Emit(session string, data string) error
	EmitLocal(session string, data string)
	Release()
}

type Hub struct {
	socketsMap  map[*glue.Socket]map[string]bool
	sessionsMap map[string]map[*glue.Socket]bool

	pubConn redis.Conn
	subConn redis.PubSubConn
	rwMutex sync.RWMutex
	glueSrv *glue.Server
}

func (p *Hub) Connect(url string) error {
	// FIXME: this seems to be reassigned!
	c, err := redis.DialURL(url)
	if err != nil {return err}

	p.pubConn = c
	c, err = redis.DialURL(url)
	if err != nil {return err}
	p.subConn = redis.PubSubConn{Conn:c}

	p.socketsMap = map[*glue.Socket]map[string]bool{}
	p.sessionsMap = map[string]map[*glue.Socket]bool{}

	go func() {
		for {
			switch v := p.subConn.Receive().(type) {
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
		err := p.subConn.Subscribe(session)
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
			err := p.subConn.Unsubscribe(session)
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
	_, err := p.pubConn.Do("PUBLISH", session, data)
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
			if err != nil {
				log.Print(err)
			}

			// spit out all the current users
			key := fmt.Sprintf("session:%s:users", sessionId)
			userIds, err := redis.Strings(p.pubConn.Do("SMEMBERS", key))
			if err != nil {log.Print(err); return}

			log.Println("Current session voters: ", userIds)

			// OPTIMIZE: batch this
			for _, userId := range userIds {
				key = fmt.Sprintf("user:%s", userId)
				_ = p.pubConn.Send("HGETALL", key)
			}

			res, err := redis.Values(p.pubConn.Do(""))

			if err != nil {
				log.Printf("ERROR ", err)
			}

			var users []model.User

			for i, r := range res {
				switch t := r.(type) {
				case redis.Error:
					fmt.Printf("res[%d] is redis.Error %v\n", i, r)
				case []interface{}:
					m, _ := redis.StringMap(r, nil)

					estimate, err := strconv.Atoi(m["estimate"])
					if err != nil {
						fmt.Println("ERROR ", err)
					}

					user := model.User{
						UserId: m["id"],
						Name: m["name"],
						Estimate: estimate,
					}
					users = append(users, user)
				default:
					log.Printf("UNEXPECTED TYPE: %T", t)
				}
			}

			// get session state - voting, not voting
			key = fmt.Sprintf("session:%s:voting", sessionId)
			isVoting, err := redis.Int(p.pubConn.Do("GET", key))

			sessionState := model.NotVoting
			if isVoting == 1 {
				sessionState = model.Voting
			}

			session := model.WsSession{
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
func (p *VoidHub) Connect(url string) error {
	p.Emitted = p.Emitted[:0]
	p.LocalEmitted = p.LocalEmitted[:0]
	return nil
}

func (p *VoidHub) HandleWebSockets(url string) {return}
func (p *VoidHub) Release() {return}
