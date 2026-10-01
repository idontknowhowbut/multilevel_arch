package handlers

import (
	"net/http"
	"tictactoe/internal/handlers/middleware"
)

func NewRouter(h *Handler, auth *middleware.AuthService) http.Handler {
	mux := http.NewServeMux()

	//auth not requiered
	mux.HandleFunc("POST /signUp", h.SignUpRequest)

	// JWT implementation
	mux.HandleFunc("POST /auth/login", h.GetTokenPair)
	mux.HandleFunc("POST /auth/refresh", h.RefreshTokenPair)

	//auth requiered
	mux.HandleFunc("POST /game/{id}", auth.Auth(h.PostGame))
	mux.HandleFunc("POST /game", auth.Auth(h.CreateGame))
	mux.HandleFunc("POST /game/{id}/join", auth.Auth(h.JoinGame))
	mux.HandleFunc("GET /game/{id}", auth.Auth(h.GetGame))

	mux.HandleFunc("GET /games", auth.Auth(h.GetAvailableGames))
	mux.HandleFunc("GET /userInfo", auth.Auth(h.GetUserInfo))
	mux.HandleFunc("GET /user", auth.Auth(h.GetUserId))
	mux.HandleFunc("GET /user/{id}/games", auth.Auth(h.GetUserFinishedGames))
	mux.HandleFunc("GET /userGames", auth.Auth(h.GetCallerFinishedGames))
	mux.HandleFunc("GET /scoreboard", auth.Auth(h.GetScoreBoard))

	return mux
}
