package response

import "github.com/papito/ballot/ballot/model"

type HealthResponse struct {
	Status string `json:"status"`
}

type WsVoteStarted struct {
	Event string `json:"event"`
}

type WsVoteFinished struct {
	Users []model.User `json:"users"`
	Tally string       `json:"tally"`
	Event string       `json:"event"`
}

type WsNewUser struct {
	model.User
	Event string `json:"event"`
}

type WsUserVote struct {
	Event  string `json:"event"`
	UserId string `json:"user_id"`
}

type WsSession struct {
	Event        string       `json:"event"`
	SessionState int          `json:"status"`
	Users        []model.User `json:"users"`
	Observers    []model.User `json:"observers"`
	Tally        string       `json:"tally"`
}

type WsUserLeftEvent struct {
	Event     string `json:"event"`
	SessionId string `json:"session_id"`
	UserId    string `json:"user_id"`
}

type WsObserverLeftEvent struct {
	Event     string `json:"event"`
	SessionId string `json:"session_id"`
	UserId    string `json:"user_id"`
}

const (
	UserAddedEvent     = "USER_ADDED"
	ObserverAddedEvent = "OBSERVER_ADDED"
	UserVotedEVent     = "USER_VOTED"
	VoteStartedEVent   = "VOTING"
	VoteFinishedEvent  = "VOTE_FINISHED"
)
