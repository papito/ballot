package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/db"
	"github.com/papito/ballot/ballot/hub"
	"github.com/papito/ballot/ballot/jsonutil"
	"github.com/papito/ballot/ballot/logutil"
	"github.com/papito/ballot/ballot/model"
	"github.com/papito/ballot/ballot/service"
	"html/template"
	"log"
	"math/rand"
	"net/http"
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

	var health = model.HealthResponse{Status: "OK"}
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
	reqBody, err := jsonutil.GetRequestBody(r)
	var reqObj model.StartVoteRequest
	err = json.Unmarshal([]byte(reqBody), &reqObj)

	if err != nil {
		log.Print(err)
		http.Error(w, "Error serializing request JSON", http.StatusBadRequest)
		return
	}

	err = p.service.StartVote(reqObj.SessionId)

	if err != nil {
		http.Error(w, "Error saving data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	logutil.Logger(fmt.Fprint(w,"{}"))
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

	if sessionState == model.NotVoting {
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

	vote := model.Vote{
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
	reqBody, err := jsonutil.GetRequestBody(r)
	var reqObj model.CreateUserRequest
	err = json.Unmarshal([]byte(reqBody), &reqObj)

	if err != nil {
		log.Print(err)
		http.Error(w, "Error serializing request JSON", http.StatusBadRequest)
		return
	}

	var user model.User
	user, err = p.service.CreateUser(reqObj.SessionId, reqObj.UserName)

	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	var httpResp, _  = json.Marshal(user)
	w.Header().Set("Content-Type", "application/json")
	logutil.Logger(fmt.Fprintf(w, "%s", httpResp))
}
