package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joomcode/errorx"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/errors"
	"github.com/papito/ballot/ballot/jsonutil"
	"github.com/papito/ballot/ballot/logutil"
	"github.com/papito/ballot/ballot/model"
	"github.com/papito/ballot/ballot/model/request"
	"github.com/papito/ballot/ballot/model/response"
	"github.com/papito/ballot/ballot/service"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
}

// Lifted straight from Gorilla Mux documentation
// https://github.com/gorilla/mux?tab=readme-ov-file#serving-single-page-applications
type spaHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Join internally call path.Clean to prevent directory traversal
	path := filepath.Join(h.staticPath, r.URL.Path)

	// check whether a file exists or is a directory at the given path
	fi, err := os.Stat(path)
	if os.IsNotExist(err) || fi.IsDir() {
		// file does not exist or path is a directory, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	}

	if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static file
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func NewServer(config config.Config) Server {
	log.Println("Creating server")
	ballotService := service.NewService(config)

	server := server{
		service: &ballotService,
	}

	// Handlers
	r := mux.NewRouter()
	r.HandleFunc("/health", server.HealthHttpHandler).Methods("GET")
	r.HandleFunc("/api/session", server.CreateSessionHttpHandler).Methods("POST")
	r.HandleFunc("/api/user/{id}", server.GetUserHttpHandler).Methods("GET")
	r.HandleFunc("/api/user", server.CreateUserHttpHandler).Methods("POST")
	r.HandleFunc("/api/vote/start", server.StartVoteHttpHandler).Methods("PUT")
	r.HandleFunc("/api/vote/finish", server.FinishVoteHttpHandler).Methods("PUT")
	r.HandleFunc("/api/vote/cast", server.CastVoteHttpHandler).Methods("PUT")

	spa := spaHandler{staticPath: "../ballot-ui/dist", indexPath: "index.html"}
	r.PathPrefix("/").Handler(spa)

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

func (p server) HealthHttpHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var health = response.HealthResponse{Status: "OK"}
	var data, _ = json.Marshal(health)

	logutil.Logger(fmt.Fprintf(w, "%s", data))
}

func (p server) CreateSessionHttpHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	session, err := p.service.CreateSession()
	if err != nil {
		log.Printf("%+v", err)
		err = errors.CriticalError{Message: "Error saving data"}
		var data, _ = json.Marshal(err)
		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	var data, _ = json.Marshal(session)
	log.Printf("API session with %+v", session)
	logutil.Logger(fmt.Fprintf(w, "%s", data))
}

func (p server) StartVoteHttpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reqBody, err := jsonutil.GetRequestBody(r)
	var reqObj request.StartVoteRequest
	err = json.Unmarshal([]byte(reqBody), &reqObj)
	if err != nil {
		log.Printf("%+v", err)
		err = errors.CriticalError{Message: "Error serializing request JSON"}
		var data, _ = json.Marshal(err)
		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	err = p.service.StartVote(reqObj.SessionId)
	if err != nil {
		log.Printf("%+v", err)
		err = errors.CriticalError{Message: "Error starting vote"}
		var data, _ = json.Marshal(err)
		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	logutil.Logger(fmt.Fprint(w, "{}"))
}

func (p server) FinishVoteHttpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reqBody, err := jsonutil.GetRequestBody(r)
	var reqObj request.StartVoteRequest
	err = json.Unmarshal([]byte(reqBody), &reqObj)
	if err != nil {
		log.Printf("%+v", err)
		err = errors.CriticalError{Message: "Error serializing request JSON"}
		var data, _ = json.Marshal(err)
		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	err = p.service.FinishVote(reqObj.SessionId)
	if err != nil {
		log.Printf("%+v", err)
		err = errors.CriticalError{Message: "Error finishing vote"}
		var data, _ = json.Marshal(err)
		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	logutil.Logger(fmt.Fprint(w, "{}"))
}

func (p server) CastVoteHttpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reqBody, err := jsonutil.GetRequestBody(r)
	var reqObj request.CastVoteRequest
	err = json.Unmarshal([]byte(reqBody), &reqObj)

	if err != nil {
		log.Printf("%+v", err)
		err = errors.CriticalError{Message: "Error serializing request JSON"}
		var data, _ = json.Marshal(err)
		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	vote, err := p.service.CastVote(reqObj.SessionId, reqObj.UserId, reqObj.Estimate)

	if err != nil {
		log.Printf("%+v", err)
		err = errors.CriticalError{Message: "Error casting vote"}
		var data, _ = json.Marshal(err)
		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(vote)
	logutil.Logger(fmt.Fprintf(w, "%s", data))
}

func (p server) CreateUserHttpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reqBody, err := jsonutil.GetRequestBody(r)
	var reqObj request.CreateUserRequest
	err = json.Unmarshal([]byte(reqBody), &reqObj)

	if err != nil {
		log.Printf("%+v", err)
		err = errors.CriticalError{Message: "Error serializing request JSON"}
		var data, _ = json.Marshal(err)
		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	var user model.User
	user, err = p.service.CreateUser(
		reqObj.SessionId,
		reqObj.UserName,
		reqObj.IsAdmin == 1,
		reqObj.IsObserver == 1)

	if err != nil {
		log.Printf("%+v", errorx.EnsureStackTrace(err))

		switch err.(type) {
		case errors.ValidationError:
			data, _ := json.Marshal(err)
			http.Error(w, string(data), http.StatusBadRequest)
		default:
			http.Error(w, "{}", http.StatusInternalServerError)
		}

		return
	}

	var httpResp, _ = json.Marshal(user)
	w.Header().Set("Content-Type", "application/json")
	logutil.Logger(fmt.Fprintf(w, "%s", httpResp))
}

func (p server) GetUserHttpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	userId := vars["id"]

	user, err := p.service.GetUser(userId)

	if err != nil {
		log.Printf("%+v", err)
		err = errors.CriticalError{Message: "Error creating user"}
		var data, _ = json.Marshal(err)
		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(user)
	logutil.Logger(fmt.Fprintf(w, "%s", data))
}
