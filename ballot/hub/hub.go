/*
 * The MIT License
 *
 * Copyright (c) 2019,  Andrei Taranchenko
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package hub

import (
	"encoding/json"
	"fmt"
	"github.com/desertbit/glue"
	"github.com/gomodule/redigo/redis"
	"github.com/joomcode/errorx"
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
	Connect(store *db.Store)
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

func (p *Hub) Connect(store *db.Store) {
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
				log.Printf("%+v", v)
			}
		}
	}()

	/* Create the Glue server */
	p.glueSrv = glue.NewServer(glue.Options{
		HTTPSocketType: glue.HTTPSocketTypeNone,
	})

	p.glueSrv.OnNewSocket(p.handleSocket)
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
		if err != nil {return errorx.EnsureStackTrace(err)}
		err = p.store.ServiceSubCon.Subscribe(sessionId)
		if err != nil {return errorx.EnsureStackTrace(err)}
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
			if err != nil {return errorx.EnsureStackTrace(err)}
		}

		userId, _ := p.userMap[sock]
		event := response.WsUserLeftEvent{
			Event:     Event.UserLeft,
			SessionId: sessionId,
			UserId:    userId,
		}

		data, err := json.Marshal(event)
		if err != nil {return errorx.EnsureStackTrace(err)}
		err = p.Emit(sessionId, string(data))
		if err != nil {return errorx.EnsureStackTrace(err)}
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
		if err != nil {log.Printf("%+v", errorx.EnsureStackTrace(err)); return}
	})

	sock.OnRead(func(data string) {
		log.Printf("Reading from socket %s: %s", sock.ID(), data)

		c  := p.store.Pool.Get()
		defer p.store.Close(c)

		jsonData, err := jsonutil.GetJsonFromString(data)
		if err != nil {log.Printf("%+v", err); return}

		var sessionId = jsonData["session_id"].(string)
		var action = jsonData["action"].(string)

		switch action {
		/*
		Emit the WATCHING event, as well as a list of current users in this session
		 */
		case Event.Watch:
			log.Printf("WS. Watching session %s", sessionId)
			err := p.Subscribe(sock, sessionId)
			if err != nil {log.Printf("%+v", err); return}

			// get session state - voting, not voting
			key := fmt.Sprintf(db.Const.SessionState, sessionId)
			isVoting, err := redis.Int(c.Do("GET", key))
			if err != nil {log.Printf("%+v", err); return}

			sessionState := model.NotVoting
			if isVoting == 1 {
				sessionState = model.Voting
			}

			if userId, ok := jsonData["user_id"].(string); ok {
				p.associateSocketWithUser(sock, userId)

				sessionUserKey := fmt.Sprintf(db.Const.SessionUsers, sessionId)
				log.Printf("Adding user [%s] to session [%s]", userId, sessionId)
				err := p.store.AddToSet(sessionUserKey, userId)
				if err != nil {return}

				user, err := p.store.GetUser(userId)
				if err != nil {log.Printf("%+v", err); return}

				wsUser := response.WsNewUser{}
				wsUser.Event = response.UserAddedEvent
				wsUser.Name = user.Name
				wsUser.UserId = user.UserId
				wsUser.Joined = user.Joined
				wsUser.Voted = user.Voted

				// only expose votes when not voting
				if sessionState == model.NotVoting {
					wsUser.Estimate = user.Estimate
				}

				wsResp, err := json.Marshal(wsUser)
				if err != nil {log.Printf("%+v", err); return}

				err = p.Emit(sessionId, string(wsResp))
				if err != nil {log.Printf("%+v", err); return}
			}

			users, err := p.store.GetSessionUsers(sessionId)
			if err != nil {log.Printf("%+v", err); return}

			// null out estimates if still voting
			for idx := range users {
				if sessionState == model.Voting {
					users[idx].Estimate = model.NoEstimate
				}
			}

			key = fmt.Sprintf(db.Const.Tally, sessionId)
			tally, err := p.store.GetStr(key)
			if err != nil {log.Printf("%+v", err)}

			session := response.WsSession{
				Event: Event.Watching,
				SessionState: sessionState,
				Users: users,
				Tally: tally,
			}

			data, err := json.Marshal(session)
			if err != nil {log.Printf("%+v", err); return}

			p.emitSocket(sock, string(data))

		case Event.Start:
			err := p.Emit(sessionId, "{}")
			if err != nil {log.Printf("%+v", err); return}

		case Event.Restart:
			err := p.Emit(sessionId, "{}")
			if err != nil {log.Printf("%+v", err); return}

		case Event.Vote:
			err := p.Emit(sessionId, "{}")
			if err != nil {log.Printf("%+v", err); return}
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
func (p *VoidHub) Connect(db *db.Store) {
	p.Emitted = p.Emitted[:0]
	p.LocalEmitted = p.LocalEmitted[:0]
}

func (p *VoidHub) HandleWebSockets(url string) {return}
func (p *VoidHub) Release() {return}
