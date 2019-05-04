package hub

import (
	"ballot/ballot/jsonutil"
	"ballot/ballot/models"
	"encoding/json"
	"fmt"
	"github.com/desertbit/glue"
	"github.com/gomodule/redigo/redis"
	"log"
	"strconv"
	"sync"
)

/* Borrowed from https://github.com/hjr265/tonesa/blob/master/hub/hub.go */

var (
	socketsMap  = map[*glue.Socket]map[string]bool{}
	sessionsMap = map[string]map[*glue.Socket]bool{}

	pubConn redis.Conn
	subConn redis.PubSubConn
	l       sync.RWMutex
)

func InitHub(url string) error {
	c, err := redis.DialURL(url)
	if err != nil {
		return err
	}
	pubConn = c

	c, err = redis.DialURL(url)
	if err != nil {
		return err
	}
	subConn = redis.PubSubConn{c}

	go func() {
		for {
			switch v := subConn.Receive().(type) {
			case redis.Message:
				log.Printf(
					"Subscribe connection received [%s] on channel [%s]", v.Data, v.Channel)
				EmitLocal(v.Channel, string(v.Data))

			case error:
				panic(v)
			}
		}
	}()

	return nil
}

func Subscribe(sock *glue.Socket, session string) error {
	log.Printf("Subscribing socket %p to session %s", sock, session)
	l.Lock()
	defer l.Unlock()

	_, ok := socketsMap[sock]
	if !ok {
		socketsMap[sock] = map[string]bool{}
	}
	socketsMap[sock][session] = true

	_, ok = sessionsMap[session]

	if !ok {
		sessionsMap[session] = map[*glue.Socket]bool{}
		err := subConn.Subscribe(session)
		if err != nil {
			return err
		}
	}
	sessionsMap[session][sock] = true

	return nil
}

func UnsubscribeAll(sock *glue.Socket) error {
	log.Printf("Unsubscribing all from socket %p", sock)
	l.Lock()
	defer l.Unlock()

	for session := range socketsMap[sock] {
		delete(sessionsMap[session], sock)
		if len(sessionsMap[session]) == 0 {
			delete(sessionsMap, session)
			err := subConn.Unsubscribe(session)
			if err != nil {
				return err
			}
		}
	}
	delete(socketsMap, sock)

	return nil
}

func Emit(session string, data string) error {
	log.Printf("EMIT. Session %s - %s", session, data)
	_, err := pubConn.Do("PUBLISH", session, data)
	return err
}


func EmitLocal(session string, data string) {
	log.Printf("EMIT LOCAL. Session %s - %s", session, data)
	l.RLock()
	defer l.RUnlock()

	// write to socketsMap interested in this session
	for socket := range sessionsMap[session] {
		socket.Write(data)
	}
}

func EmitSocket(sock *glue.Socket,data string) {
	log.Printf("EMIT SOCKET. Socket %s - %s", sock.ID(), data)
	l.RLock()
	defer l.RUnlock()

	sock.Write(data)
}

func HandleSocket(sock *glue.Socket) {
	log.Printf("Handling socket %p", sock)

	sock.OnClose(func() {
		log.Printf("Socket %p closed", sock)
		err := UnsubscribeAll(sock)
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
			err := Subscribe(sock, sessionId)
			if err != nil {
				log.Print(err)
			}

			// spit out all the current users
			key := fmt.Sprintf("session:%s:users", sessionId)
			userIds, err := redis.Strings(pubConn.Do("SMEMBERS", key))
			if err != nil {
				log.Print(err)
				return
			}

			log.Println("Current session voters: ", userIds)

			// OPTIMIZE: batch this
			for _, userId := range userIds {
				key = fmt.Sprintf("user:%s", userId)
				_ = pubConn.Send("HGETALL", key)
			}

			res, err := redis.Values(pubConn.Do(""))

			if err != nil {
				fmt.Println("ERROR ", err)
			}

			var users []models.User

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

					user := models.User{
						UserId: m["id"],
						Name: m["name"],
						Estimate: estimate,
					}
					users = append(users, user)
				default:
					fmt.Printf("UNEXPECTED TYPE: %T", t)
				}
			}

			// get session state - voting, not voting
			key = fmt.Sprintf("session:%s:voting", sessionId)
			isVoting, err := redis.Int(pubConn.Do("GET", key))

			sessionState := models.NotVoting
			if isVoting == 1 {
				sessionState = models.Voting
			}

			type WsSession struct {
				Event string `json:"event"`
				SessionState int `json:"session_state"`
				Users []models.User `json:"users"`
			}

			session := WsSession{
				Event: "WATCHING",
				SessionState: sessionState,
				Users: users,
			}

			data, err := json.Marshal(session)
			if err != nil {
				log.Println(err)
			}

			EmitSocket(sock, string(data))

		case "START":
			err := Emit(sessionId, "{}")
			if err != nil {
				log.Print(err)
			}

		case "RESTART":
			err := Emit(sessionId, "{}")
			if err != nil {
				log.Print(err)
			}

		case "VOTE":
			err := Emit(sessionId, "{}")
			if err != nil {
				log.Print(err)
			}
		}
	})
}