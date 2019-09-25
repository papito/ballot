package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/jsonutil"
	"github.com/papito/ballot/ballot/logutil"
	"github.com/papito/ballot/ballot/model"
	"github.com/papito/ballot/ballot/model/request"
	"github.com/papito/ballot/ballot/model/response"
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
	GetUserHttpHandler(w http.ResponseWriter, r *http.Request)
	StartVoteHttpHandler(w http.ResponseWriter, r *http.Request)
	FinishVoteHttpHandler(w http.ResponseWriter, r *http.Request)
	CastVoteHttpHandler(w http.ResponseWriter, r *http.Request)
	Service() *service.Service
}

type server struct {
	service *service.Service
	templates *template.Template
}

func NewServer(config config.Config) Server {
	log.Println("Creating server")
	ballotService, err := service.NewService(config)

	if err !=  nil {
		log.Fatalf("Could not initialize service. %v", err)
	}
	server := server{
		service: &ballotService,
		templates: template.Must(template.ParseGlob("../ui/templates/*")),
	}

	// Serve static files
	fs := http.FileServer(http.Dir("../ui/dist/"))
	http.Handle("/ui/",http.StripPrefix("/ui/", fs))

	// Handlers
	r := mux.NewRouter()
	r.HandleFunc("/", server.indexHttpHandler).Methods("GET")
	r.HandleFunc("/health", server.HealthHttpHandler).Methods("GET")
	r.HandleFunc("/api/session", server.CreateSessionHttpHandler).Methods("POST")
	r.HandleFunc("/api/user/{id}", server.GetUserHttpHandler).Methods("GET")
	r.HandleFunc("/api/user", server.CreateUserHttpHandler).Methods("POST")
	r.HandleFunc("/api/vote/start", server.StartVoteHttpHandler).Methods("PUT")
	r.HandleFunc("/api/vote/finish", server.FinishVoteHttpHandler).Methods("PUT")
	r.HandleFunc("/api/vote/cast", server.CastVoteHttpHandler).Methods("PUT")
	http.Handle("/", r)

	server.service.Hub().HandleWebSockets("/glue/ws")

	return server
}

func (p server) Service() *service.Service {
	return p.service
}

func (p server) Release() {
	log.Print("Releasing server resources")
	p.service.Release()
	log.Print("Server done")
}

func (p server) HealthHttpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var health = response.HealthResponse{Status: "OK"}
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

func (p server) StartVoteHttpHandler(w http.ResponseWriter, r *http.Request)  {
	reqBody, err := jsonutil.GetRequestBody(r)
	var reqObj request.StartVoteRequest
	err = json.Unmarshal([]byte(reqBody), &reqObj)
	if err != nil {
		log.Printf("Error serializing request JSON. %v", err)
		http.Error(w, "Error serializing request JSON", http.StatusBadRequest)
		return
	}

	err = p.service.StartVote(reqObj.SessionId)
	if err != nil {
		log.Printf("Error starting vote. %v", err)
		http.Error(w, "Error starting vote", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	logutil.Logger(fmt.Fprint(w, "{}"))
}

func (p server) FinishVoteHttpHandler(w http.ResponseWriter, r *http.Request)  {
	reqBody, err := jsonutil.GetRequestBody(r)
	var reqObj request.StartVoteRequest
	err = json.Unmarshal([]byte(reqBody), &reqObj)
	if err != nil {
		log.Printf("Error serializing request JSON. %v", err)
		http.Error(w, "Error serializing request JSON", http.StatusBadRequest)
		return
	}

	err = p.service.FinishVote(reqObj.SessionId)
	if err != nil {
		log.Printf("Error finishing vote. %v", err)
		http.Error(w, "Error finishing vote", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	logutil.Logger(fmt.Fprint(w, "{}"))
}

func (p server) CastVoteHttpHandler(w http.ResponseWriter, r *http.Request)  {
	reqBody, err := jsonutil.GetRequestBody(r)
	var reqObj request.CastVoteRequest
	err = json.Unmarshal([]byte(reqBody), &reqObj)

	if err != nil {
		log.Printf("Error serializing request JSON. %v", err)
		http.Error(w, "Error serializing request JSON", http.StatusBadRequest)
		return
	}

	vote, err := p.service.CastVote(reqObj.SessionId, reqObj.UserId, reqObj.Estimate)

	if err != nil {
		log.Printf("Error casting vote. %s", err)
		http.Error(w, "Error casting vote", http.StatusInternalServerError)
		return
	}

	data, _  := json.Marshal(vote)
	w.Header().Set("Content-Type", "application/json")
	logutil.Logger(fmt.Fprintf(w, "%s", data))
}

func (p server) CreateUserHttpHandler(w http.ResponseWriter, r *http.Request) {
	reqBody, err := jsonutil.GetRequestBody(r)
	var reqObj request.CreateUserRequest
	err = json.Unmarshal([]byte(reqBody), &reqObj)

	if err != nil {
		log.Printf("Error serializing request JSON. %v", err)
		http.Error(w, "Error serializing request JSON", http.StatusBadRequest)
		return
	}

	var user model.User
	user, err = p.service.CreateUser(reqObj.SessionId, reqObj.UserName)

	if err != nil {
		log.Printf("Error creating user. %s", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	var httpResp, _  = json.Marshal(user)
	w.Header().Set("Content-Type", "application/json")
	logutil.Logger(fmt.Fprintf(w, "%s", httpResp))
}

func (p server) GetUserHttpHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["id"]

	user, err := p.service.GetUser(userId)

	if err != nil {
		log.Printf("Error getting user. %s", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	data, _  := json.Marshal(user)
	w.Header().Set("Content-Type", "application/json")
	logutil.Logger(fmt.Fprintf(w, "%s", data))
}
