package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"tictactoe/internal/app"
	"tictactoe/internal/domain"
	jwtservice "tictactoe/internal/jwt"

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

	writeJSON(w, http.StatusOK, gameFromDomain(newGameState))
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

	writeJSON(w, http.StatusOK, gameFromDomain(newGame))
}

func (h *Handler) SignUpRequest(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.SignUpUser(req.Login, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Registration immediately authenticates the newly created user,
	// so signUp and login return the same token pair contract.
	userId, err := h.service.Authenticate(req.Login, req.Password)
	if err != nil {
		http.Error(w, "Auth failed", http.StatusInternalServerError)
		return
	}

	resp, err := h.createTokenResponse(userId)
	if err != nil {
		http.Error(w, "Auth failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetUserId(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId")

	resp := getUserIdResponse{
		UserID: userId.(string),
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetAvailableGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId := ctx.Value("userId").(string)

	userIdUuid, err := uuid.Parse(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	games, err := h.service.GetAvailableGames(userIdUuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, gamesFromDomain(games))
}

func (h *Handler) JoinGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId := ctx.Value("userId").(string)

	userIdUuid, err := uuid.Parse(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	gameId, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = h.service.JoinGame(userIdUuid, gameId); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	game, err := h.service.GetGame(gameId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, gameFromDomain(game))
}

func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
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

	writeJSON(w, http.StatusOK, gameFromDomain(game))
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
	req := credentialsRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Auth failed", http.StatusUnauthorized)
		return
	}

	userId, err := h.service.Authenticate(req.Login, req.Password)
	if err != nil {
		http.Error(w, "Auth failed", http.StatusUnauthorized)
		return
	}

	resp, err := h.createTokenResponse(userId)
	if err != nil {
		http.Error(w, "Auth failed", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) RefreshTokenPair(w http.ResponseWriter, r *http.Request) {
	req := refreshTokensRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	resp, err := h.createTokenResponse(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetUserFinishedGames(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")

	userIdUuid, err := uuid.Parse(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	games, err := h.service.GetUserFinishedGames(userIdUuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, gamesFromDomain(games))
}

func (h *Handler) GetCallerFinishedGames(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(string)
	userIdUuid, err := uuid.Parse(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	games, err := h.service.GetUserFinishedGames(userIdUuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, gamesFromDomain(games))
}

func (h *Handler) GetScoreBoard(w http.ResponseWriter, r *http.Request) {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	scores, err := h.service.GetScoreBoard(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, scoreBoardFromDomain(scores))
}

func (h *Handler) createTokenResponse(userId string) (tokenResponse, error) {
	tokenAccess, tokenRefresh, err := h.service.CreateTokenPair(userId)
	if err != nil {
		return tokenResponse{}, err
	}

	return tokenResponse{
		AccessToken:  tokenAccess,
		RefreshToken: tokenRefresh,
	}, nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		fmt.Println(err)
	}
}
