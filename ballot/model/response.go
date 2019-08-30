package model

type HealthResponse struct {
	Status string `json:"status"`
}

type WsSession struct {
	Event string `json:"event"`
}

type WsUserVote struct {
	Event string `json:"event"`
	UserId string `json:"user_id"`
	Estimate uint8 `json:"estimate"`
}
