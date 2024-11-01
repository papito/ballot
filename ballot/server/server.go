/*
 * The MIT License
 *
 * Copyright (c) 2020,  Andrei Taranchenko
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
	"github.com/shurcooL/httpgzip"
	"html/template"
	"log"
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
	service   *service.Service
	templates *template.Template
}

func NewServer(config config.Config) Server {
	log.Println("Creating server")
	ballotService := service.NewService(config)

	server := server{
		service: &ballotService,
	}

	// Serve static files
	http.Handle("/p/", http.StripPrefix("/p/", httpgzip.FileServer(
		http.Dir("../ballot-ui/dist/"),
		httpgzip.FileServerOptions{
			IndexHTML: true,
		},
	)))

	http.Handle("/p/vote", http.StripPrefix("/p/vote", httpgzip.FileServer(
		http.Dir("../ballot-ui/dist/"),
		httpgzip.FileServerOptions{
			IndexHTML: true,
		},
	)))

	http.Handle("/assets/", http.StripPrefix("/assets/", httpgzip.FileServer(
		http.Dir("../ballot-ui/dist/assets"),
		httpgzip.FileServerOptions{
			IndexHTML: false,
		},
	)))

	// Handlers
	r := mux.NewRouter()
	r.HandleFunc("/", server.indexHttpHandler).Methods("GET")
	r.HandleFunc("/vote/{sessionId}", server.gotoVoteHandler).Methods("GET")
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
	http.Redirect(w, r, "/p/", http.StatusPermanentRedirect)
}

func (p server) gotoVoteHandler(w http.ResponseWriter, r *http.Request) {
	err := p.templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Fatalf("Error getting index view %+v", errorx.EnsureStackTrace(err))
	}
}

func (p server) CreateSessionHttpHandler(w http.ResponseWriter, r *http.Request) {
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
