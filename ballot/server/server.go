package server

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/db"
	"github.com/papito/ballot/ballot/hub"
	"github.com/papito/ballot/ballot/jsonutil"
	"github.com/papito/ballot/ballot/logutil"
	"github.com/papito/ballot/ballot/models"
	"github.com/papito/ballot/ballot/service"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

type Server interface {
	Release()
	HealthHttpHandler(w http.ResponseWriter, r *http.Request)
	CreateSessionHttpHandler(w http.ResponseWriter, r *http.Request)
	CreateUserHttpHandler(w http.ResponseWriter, r *http.Request)
	Store() *db.Store
	Hub() *hub.Hub
	Service() *service.Service
}

type server struct {
	service *service.Service
	templates *template.Template
	store *db.Store
	hub *hub.Hub
}

func NewServer(config config.Config) Server {
	log.Println("Creating server")
	log.Printf("Environment is %s", config.Environment)

	ballotService, err := service.NewService(config)

	server := server{
		service: &ballotService,
		templates: template.Must(template.ParseGlob("../ui/templates/*")),
		store: &db.Store{},
		hub: &hub.Hub{},
	}

	server.store.Connect(config.RedisUrl)

	/* Initiate the hub that connects sessions and sockets
	 */
	log.Println("Creating hub")
	err = server.hub.Connect(config.RedisUrl)
	if err != nil {
		log.Fatal(err)
	}

	// Serve static files
	fs := http.FileServer(http.Dir("../ui/dist/js"))
	http.Handle("../ui/js/",http.StripPrefix("../ui/js/", fs))

	// Handlers
	r := mux.NewRouter()
	r.HandleFunc("/", server.indexHttpHandler).Methods("GET")
	r.HandleFunc("/health", server.HealthHttpHandler).Methods("GET")
	r.HandleFunc("/api/session", server.CreateSessionHttpHandler).Methods("POST")
	r.HandleFunc("/api/user", server.CreateUserHttpHandler).Methods("POST")
	r.HandleFunc("/api/vote/start", server.startVoteHttpHandler).Methods("PUT")
	r.HandleFunc("/api/vote/cast", server.castVoteHttpHandler).Methods("PUT")
	http.Handle("/", r)

	server.hub.HandleWebSockets("/glue/ws")

	return server
}

func (p server) Store() *db.Store {
	return p.store
}

func (p server) Hub() *hub.Hub {
	return p.hub
}

func (p server) Service() *service.Service {
	return p.service
}

func (p server) Release() {
	log.Print("Releasing server resources")
	p.hub.Release()
	log.Print("Server done")
}

func (p server) HealthHttpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type Health struct {
		Status string `json:"status"`
	}

	var health = Health{Status:"OK"}
	var data, _ = json.Marshal(health)

	logutil.Logger(fmt.Fprintf(w, "%s", data))
}

func (p server) indexHttpHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving Index")

	nocache := rand.Intn(1000000)
	err := p.templates.ExecuteTemplate(w, "index.html", nocache)
	if err != nil {
		log.Fatal("Error getting index view ", err)
	}
}

func (p server) CreateSessionHttpHandler(w http.ResponseWriter, r *http.Request) {
	session, err := p.service.CreateSession()

	if err != nil {
		log.Printf("Error creating session: %s", err)
		http.Error(w, "Error saving data", http.StatusInternalServerError)
		return
	}

	var data, _  = json.Marshal(session)
	log.Printf("API session with %+v", session)
	w.Header().Set("Content-Type", "application/json")
	logutil.Logger(fmt.Fprintf(w, "%s", data))
}

func (p server) startVoteHttpHandler(w http.ResponseWriter, r *http.Request)  {
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

	key := fmt.Sprintf(db.Const.SessionVoting, sessionId)
	err = p.store.SetKey(key, models.Voting)

	if err != nil {
		http.Error(w, "Error saving data", http.StatusInternalServerError)
		return
	}

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

	err = p.hub.Emit(sessionId, string(data))

	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "application/json")
	logutil.Logger(fmt.Fprintf(w, "%s", data))
}

func (p server) castVoteHttpHandler(w http.ResponseWriter, r *http.Request)  {
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
	sessionKey := fmt.Sprintf(db.Const.SessionVoting, sessionId)

	sessionState, err := p.store.GetInt(sessionKey)

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
	err = p.store.SetHashKey(userKey, "estimate", estimate)
	if err != nil {
		http.Error(w, "Error saving data", http.StatusInternalServerError)
		return
	}

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

	err = p.hub.Emit(sessionId, string(data))

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
	logutil.Logger(fmt.Fprintf(w, "%s", data))
}

func (p server) CreateUserHttpHandler(w http.ResponseWriter, r *http.Request) {
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
	err = p.store.SetHashKey(
		userKey,
		"name", user.Name,
		"id", user.UserId,
		"estimate", user.Estimate,)

	if err != nil {
		http.Error(w, "Error saving data", http.StatusInternalServerError)
		return
	}

	sessionUserKey := fmt.Sprintf("session:%s:users", sessionId)
	err = p.store.AddToSet(sessionUserKey, userId)

	if err != nil {
		http.Error(w, "Error saving data", http.StatusInternalServerError)
		return
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

	err = p.hub.Emit(sessionId, string(wsResp))

	if err != nil {
		log.Println(err)
	}

	var httpResp, _  = json.Marshal(user)
	w.Header().Set("Content-Type", "application/json")
	logutil.Logger(fmt.Fprintf(w, "%s", httpResp))
}
