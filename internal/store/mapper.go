package store

import (
	"tictactoe/internal/domain"

	"github.com/google/uuid"
)

func fromDomainToStore(domainGame domain.Game) (storeGame game) {
	storeGame.id = domainGame.Id.String()
	storeGame.board = board(domainGame.Board)
	storeGame.movePlayerId = domainGame.MovePlayerId.String()
	storeGame.ownerId = domainGame.OwnerId.String()
	storeGame.createdAt = domainGame.CreatedAt

	domainStatus := domainGame.Status
	switch domainStatus {
	case domain.StatusCreated:
		storeGame.status = "CREATED"
	case domain.StatusPlayerMove:
		storeGame.status = "PLAYER_MOVE"
	case domain.StatusWaitingForPlayer:
		storeGame.status = "WAITING_FOR_PLAYER"
	case domain.StatusDraw:
		storeGame.status = "DRAW"
	case domain.StatusPlayerWon:
		storeGame.status = "PLAYER_WON"
	default:
		storeGame.status = "UNKNOWN"
	}

	domainGameType := domainGame.Type
	switch domainGameType {
	case domain.TypePVE:
		storeGame.gameType = "PVE"
	case domain.TypePVP:
		storeGame.gameType = "PVP"
	}

	return storeGame
}

func fromStoreToDomain(storeGame game) (domainGame domain.Game) {
	domainGame.Id, _ = uuid.Parse(storeGame.id)
	domainGame.Board = domain.Board(storeGame.board)
	domainGame.MovePlayerId, _ = uuid.Parse(storeGame.movePlayerId)
	domainGame.OwnerId, _ = uuid.Parse(storeGame.ownerId)
	domainGame.Type = domain.GameType(storeGame.gameType)
	domainGame.CreatedAt = storeGame.createdAt

	storeStatus := storeGame.status
	switch storeStatus {
	case "CREATED":
		domainGame.Status = domain.StatusCreated
	case "PLAYER_MOVE":
		domainGame.Status = domain.StatusPlayerMove
	case "WAITING_FOR_PLAYER":
		domainGame.Status = domain.StatusWaitingForPlayer
	case "DRAW":
		domainGame.Status = domain.StatusDraw
	case "PLAYER_WON":
		domainGame.Status = domain.StatusPlayerWon
	default:
		domainGame.Status = "unknown"
	}

	return domainGame
}
