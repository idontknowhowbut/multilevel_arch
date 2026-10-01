package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"tictactoe/internal/app"
	"tictactoe/internal/domain"
	jwtservice "tictactoe/internal/jwt"
	"time"

	"github.com/google/uuid"
)

type Handler struct {
	service *app.Service
}

func New(service *app.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) PostGame(w http.ResponseWriter, r *http.Request) {
	gameId := r.PathValue("id")

	var req gameRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	board := fromHandlerToDomain(req.Board)

	userId, ok := r.Context().Value("userId").(string)

	if !ok {
		http.Error(w, "prank gone wrong", http.StatusBadRequest)
		return
	}

	newGameState, err := h.service.ProcessMove(gameId, board, userId)
	if err == app.ErrWrongPlayer {
		http.Error(w, "You can't make a move now or this is not your game", http.StatusForbidden)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := gameResponse{
		Id:         newGameState.Id.String(),
		Board:      Board(newGameState.Board),
		GameType:   string(newGameState.Type),
		MoveUserId: newGameState.MovePlayerId.String(),
		Status:     string(newGameState.Status),
		CreatedAt:  newGameState.CreatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}

func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var req createGameRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	fmt.Println(err)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		http.Error(w, "prank gone wrong", http.StatusBadRequest)
		return
	}

	gameType := domain.GameType(req.GameType)

	newGame, err := h.service.CreateGame(userId, gameType)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := gameResponse{
		Id:         newGame.Id.String(),
		Board:      fromDomainToHandler(newGame.Board),
		GameType:   req.GameType,
		MoveUserId: userId,
		Status:     string(domain.StatusCreated),
		CreatedAt:  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) SignUpRequest(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.service.SignUpUser(req.Login, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetUserId(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId")

	resp := getUserIdResponse{
		UserID: userId.(string),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetAvailableGames(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	userId := ctx.Value("userId").(string)

	userIdUuid, _ := uuid.Parse(userId)

	games, _ := h.service.GetAvailableGames(userIdUuid)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}

func (h *Handler) JoinGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	userId := ctx.Value("userId").(string)

	userIdUuid, _ := uuid.Parse(userId)
	gameId, err := uuid.Parse(r.PathValue("id"))

	err = h.service.JoinGame(userIdUuid, gameId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	json.NewEncoder(w).Encode(gameId)
}

func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gameId, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	game, err := h.service.GetGame(gameId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp := gameResponse{
		Board:      fromDomainToHandler(game.Board),
		Id:         game.Id.String(),
		GameType:   string(game.Type),
		MoveUserId: game.MovePlayerId.String(),
	}

	json.NewEncoder(w).Encode(resp)

}

func (h *Handler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userId := r.PathValue("id")
	userLogin, err := h.service.GetUserInfo(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(userLogin)
}

func (h *Handler) GetTokenPair(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	req := credentialsRequest{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Auth failed", http.StatusUnauthorized)
		return
	}

	userId, err := h.service.Authenticate(req.Login, req.Password)
	if err != nil {
		http.Error(w, "Auth failed", http.StatusUnauthorized)
		return
	}

	tokenAccess, tokenRefresh, err := h.service.CreateTokenPair(userId)
	if err != nil {
		http.Error(w, "Auth failed", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	resp := tokenResponse{
		AccessToken:  tokenAccess,
		RefreshToken: tokenRefresh,
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) RefreshTokenPair(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	req := refreshTokensRequest{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userId, tokenType, err := h.service.Jwtservice.ValidateToken(req.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if tokenType != jwtservice.Refresh {
		http.Error(w, "Incorrect token type. Needs \"refresh\"", http.StatusUnauthorized)
		return
	}

	tokenAccess, tokenRefresh, err := h.service.CreateTokenPair(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	resp := tokenResponse{
		AccessToken:  tokenAccess,
		RefreshToken: tokenRefresh,
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetUserFinishedGames(w http.ResponseWriter, r *http.Request) {

	userId := r.PathValue("id")

	userIdUuid, _ := uuid.Parse(userId)

	games, _ := h.service.GetUserFinishedGames(userIdUuid)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}

func (h *Handler) GetCallerFinishedGames(w http.ResponseWriter, r *http.Request) {

	userId := r.Context().Value("userId").(string)
	userIdUuid, _ := uuid.Parse(userId)

	games, _ := h.service.GetUserFinishedGames(userIdUuid)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}

func (h *Handler) GetScoreBoard(w http.ResponseWriter, r *http.Request) {

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	games, _ := h.service.GetScoreBoard(limit)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}
