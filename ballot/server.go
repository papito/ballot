package main

import (
	"ballot/ballot/hub"
	"ballot/ballot/jsonutil"
	"ballot/ballot/models"
	"encoding/json"
	"fmt"
	"github.com/desertbit/glue"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

type Server interface {}

type server struct {
	redisConn redis.Conn
	redisUrl string
	templates *template.Template
}

func NewServer(glueServer *glue.Server) Server {
	redisConn, err := redis.DialURL(redisUrl)
	if err != nil {
		log.Fatal("Error getting index view ", err)
	}

	server := server{
		redisUrl: redisUrl,
		redisConn: redisConn,
		templates: template.Must(template.ParseGlob("ui/templates/*")),
	}

	http.Handle("/glue/ws", glueServer)

	/* Serve static files
	 */
	fs := http.FileServer(http.Dir("ui/dist/js"))
	http.Handle("/ui/js/",http.StripPrefix("/ui/js/", fs))


	/* Handlers
	 */
	r := mux.NewRouter()
	r.HandleFunc("/", server.indexHttpHandler).Methods("GET")
	r.HandleFunc("/api/session", server.createSessionHttpHandler).Methods("POST")
	r.HandleFunc("/api/user", server.createUserHttpHandler).Methods("POST")
	r.HandleFunc("/api/vote/start", server.startVoteHttpHandler).Methods("PUT")
	r.HandleFunc("/api/vote/cast", server.castVoteHttpHandler).Methods("PUT")
	http.Handle("/", r)

	log.Println("Starting server")
	log.Fatal(http.ListenAndServe(":8080", nil))

	return server
}


func (s server) indexHttpHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving Index")

	nocache := rand.Intn(1000000)
	err := s.templates.ExecuteTemplate(w, "index.html", nocache)
	if err != nil {
		log.Fatal("Error getting index view ", err)
	}
}

func (s server) createSessionHttpHandler(w http.ResponseWriter, r *http.Request) {
	sessionUUID, _ := uuid.NewRandom()
	sessionId := sessionUUID.String()
	session := models.Session{SessionId: sessionId}

	key := fmt.Sprintf("session:%s:voting", sessionId)
	_, err := s.redisConn.Do("SET", key, models.NotVoting)
	log.Printf("Session %s saved", sessionId)

	if err != nil {
		log.Println(err)
		http.Error(w, "Error saving session", http.StatusInternalServerError)
		return
	}

	var data, _  = json.Marshal(session)
	log.Printf("API session with %+v", session)
	w.Header().Set("Content-Type", "application/json")
	logerr(fmt.Fprintf(w, "%s", data))
}

func (s server) startVoteHttpHandler(w http.ResponseWriter, r *http.Request)  {
	jsonData, err := jsonutil.GetRequestJson(r)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	// get session id
	if jsonData["session_id"] == nil {
		http.Error(w, "Must specify 'SessionId'", http.StatusBadRequest)
		return
	}
	sessionId := jsonData["session_id"].(string)
	log.Printf("Starting vote for session ID [%s]", sessionId)

	key := fmt.Sprintf("session:%s:voting", sessionId)
	_, err = s.redisConn.Do("SET", key, 1)
	log.Printf("Session %s voting", sessionId)

	type WsSession struct {
		Event string `json:"event"`
	}

	session := WsSession{
		Event: "VOTING",
	}

	data, err := json.Marshal(session)
	if err != nil {
		log.Println(err)
	}

	err = hub.Emit(sessionId, string(data))

	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "application/json")
	logerr(fmt.Fprintf(w, "%s", data))
}

func (s server) castVoteHttpHandler(w http.ResponseWriter, r *http.Request)  {
	jsonData, err := jsonutil.GetRequestJson(r)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	// get session id
	if jsonData["session_id"] == nil {
		http.Error(w, "Must specify 'SessionId'", http.StatusBadRequest)
		return
	}
	sessionId := jsonData["session_id"].(string)
	log.Printf("Voting for session ID [%s]", sessionId)

	// cannot vote on session that is inactive
	sessionKey := fmt.Sprintf("session:%s:voting", sessionId)
	sessionState, err := redis.Int(s.redisConn.Do("GET", sessionKey))

	if sessionState == models.NotVoting {
		http.Error(w, "Not voting yet for session " + sessionId, http.StatusBadRequest)
		return
	}

	// get user id
	if jsonData["user_id"] == nil {
		http.Error(w, "Must specify 'UserId'", http.StatusBadRequest)
		return
	}
	userId := jsonData["user_id"].(string)

	// get user vote value
	if jsonData["estimate"] == nil {
		http.Error(w, "Must specify 'Estimate'", http.StatusBadRequest)
		return
	}

	estimate := int(jsonData["estimate"].(float64))
	log.Printf("Estimate: %d", estimate)

	if err != nil {
		log.Println(err)
		http.Error(w, "Error voting on session", http.StatusInternalServerError)
		return
	}

	log.Printf("Voting for user ID [%s] with estimate [%d]", userId, estimate)

	userKey := fmt.Sprintf("user:%s", userId)
	_, err = s.redisConn.Do(
		"HSET", userKey,
		"estimate", estimate)
	log.Printf("User %s voted with %d", userId, estimate)

	type WsUserVote struct {
		Event string `json:"event"`
		UserId string `json:"user_id"`
		Estimate int `json:"estimate"`
	}

	wsUserVote := WsUserVote {
		Event:"USER_VOTED",
		UserId:userId,
		Estimate:estimate,
	}

	data, err := json.Marshal(wsUserVote)
	if err != nil {
		log.Println(err)
	}

	err = hub.Emit(sessionId, string(data))

	if err != nil {
		log.Print(err)
	}

	vote := models.Vote{
		SessionId: sessionId,
		UserId: userId,
		Estimate: estimate,
	}

	data, _  = json.Marshal(vote)
	w.Header().Set("Content-Type", "application/json")
	log.Println(string(data))
	logerr(fmt.Fprintf(w, "%s", data))
}

func (s server) createUserHttpHandler(w http.ResponseWriter, r *http.Request) {
	jsonData, err := jsonutil.GetRequestJson(r)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	// get session id
	if jsonData["session_id"] == nil {
		http.Error(w, "Must specify 'SessionId'", http.StatusBadRequest)
		return
	}

	sessionId := jsonData["session_id"].(string)
	log.Printf("Session ID [%s]", sessionId)

	// get user name
	if jsonData["name"] == nil {
		http.Error(w, "Must specify 'Name'", http.StatusBadRequest)
		return
	}
	name := jsonData["name"].(string)
	name = strings.TrimSpace(name)
	log.Printf("Creating user [%s]", name)

	if len(name) < 1 {
		valErr := models.ValidationError{
			Field: "name",
			Error: "This field cannot be empty"}

		data, err := json.Marshal(valErr)
		log.Println(string(data))

		if err != nil {
			http.Error(w, "Error creating response", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusBadRequest)
		return
	}

	userUUID, _ := uuid.NewRandom()
	userId := userUUID.String()

	user := models.User{
		UserId: userId,
		Name: name,
		Estimate: models.NoEstimate,
	}
	log.Println(user)

	userKey := fmt.Sprintf("user:%s", userId)
	_, err = s.redisConn.Do(
		"HSET", userKey,
		"name", user.Name,
		"id", user.UserId,
		"estimate", user.Estimate)
	log.Printf("User %s saved", user.Name)

	sessionUserKey := fmt.Sprintf("session:%s:users", sessionId)
	_, err = s.redisConn.Do("SADD", sessionUserKey, userId)
	log.Printf("Added user %s to session %s", userId, sessionId)

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

	err = hub.Emit(sessionId, string(wsResp))

	if err != nil {
		log.Println(err)
	}

	var httpResp, _  = json.Marshal(user)
	w.Header().Set("Content-Type", "application/json")
	logerr(fmt.Fprintf(w, "%s", httpResp))
}
