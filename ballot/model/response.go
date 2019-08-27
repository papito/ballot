package model

type HealthResponse struct {
	Status string `json:"status"`
}

type WsSession struct {
	Event string `json:"event"`
}

type WsUser struct {
	User
	Event  string `json:"event"`
}

