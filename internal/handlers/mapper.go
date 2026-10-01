package handlers

import (
	"tictactoe/internal/domain"

	"github.com/google/uuid"
)

func fromDomainToHandler(domainBoard domain.Board) (handlerBoard Board) {
	handlerBoard = Board(domainBoard)
	return handlerBoard
}

func fromHandlerToDomain(handlerBoard Board) (domainBoard domain.Board) {
	domainBoard = domain.Board(handlerBoard)
	return domainBoard
}

type userGameResponse struct {
	gameResponse
	Result string `json:"result"`
}

func gameFromDomain(game domain.Game) gameResponse {
	return gameResponse{
		Id:         game.Id.String(),
		Board:      fromDomainToHandler(game.Board),
		OwnerId:    game.OwnerId.String(),
		GameType:   string(game.Type),
		MoveUserId: game.MovePlayerId.String(),
		Status:     string(game.Status),
		CreatedAt:  game.CreatedAt,
	}
}

func gamesFromDomain(games []domain.Game) []gameResponse {
	result := make([]gameResponse, 0, len(games))
	for _, game := range games {
		result = append(result, gameFromDomain(game))
	}
	return result
}

func userGamesFromDomain(games []domain.Game, userId uuid.UUID) []userGameResponse {
	result := make([]userGameResponse, 0, len(games))
	for _, game := range games {
		result = append(result, userGameResponse{
			gameResponse: gameFromDomain(game),
			Result:       string(domain.ResultForUser(game, userId)),
		})
	}
	return result
}

func scoreBoardFromDomain(scores []domain.UserScore) []userScoreResponse {
	result := make([]userScoreResponse, 0, len(scores))
	for _, score := range scores {
		result = append(result, userScoreResponse{
			UserId:    score.UserId,
			UserLogin: score.UserLogin,
			WinRate:   score.Winrate,
		})
	}
	return result
}
