package handlers

import "time"

type Board [3][3]int

type createGameRequest struct {
	GameType string `json:"gameType"`
}

type gameRequest struct {
	Board Board `json:"board" validate:"required"`
}

type gameResponse struct {
	Id         string    `json:"id"`
	Board      Board     `json:"board"`
	OwnerId    string    `json:"ownerId"`
	GameType   string    `json:"gameType"`
	MoveUserId string    `json:"moveUserId"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type credentialsRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type refreshTokensRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type tokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type getUserIdResponse struct {
	UserID string `json:"userId"`
}

type userScoreResponse struct {
	UserId    string  `json:"userId"`
	UserLogin string  `json:"userLogin"`
	WinRate   float32 `json:"winRate"`
}
