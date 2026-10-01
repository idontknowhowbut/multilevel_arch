package domain

import (
	"time"

	"github.com/google/uuid"
)

type Board [3][3]int

type Game struct {
	Id           uuid.UUID
	Board        Board
	OwnerId      uuid.UUID
	Type         GameType
	Status       GameStatus
	MovePlayerId uuid.UUID
	CreatedAt    time.Time
}

type GameType string

const (
	TypePVE GameType = "PVE"
	TypePVP GameType = "PVP"
)

type GameStatus string

const (
	StatusCreated          GameStatus = "Game created"
	StatusPlayerMove       GameStatus = "Waiting for player move"
	StatusWaitingForPlayer GameStatus = "Waiting for player connection"
	StatusPlayerWon        GameStatus = "Player won"
	StatusBotWon           GameStatus = "Bot won"
	StatusDraw             GameStatus = "Draw"
)

type UserScore struct {
	UserId    string
	UserLogin string
	Winrate   float32
}
