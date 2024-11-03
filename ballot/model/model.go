package model

type Session struct {
	SessionId string `json:"id"`
}

type User struct {
	UserId     string `json:"id"`
	Name       string `json:"name"`
	Estimate   string `json:"estimate"`
	Voted      bool   `json:"voted"`
	Joined     string `json:"joined"`
	IsObserver bool   `json:"is_observer"`
	IsAdmin    bool   `json:"is_admin"`
}

type PendingVote struct {
	SessionId string `json:"session_id"`
	UserId    string `json:"user_id"`
	// Pending vote has no estimate since we are hiding it while the vote is going on
}

const (
	NotVoting = iota
	Voting
)

const NoEstimate = ""
