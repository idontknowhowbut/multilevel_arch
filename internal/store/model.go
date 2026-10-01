package store

import "time"

type board [3][3]int

type game struct {
	id           string    `db:"id"`
	board        board     `db:"board"`
	gameType     string    `db:"type"`
	status       string    `db:"status"`
	movePlayerId string    `db:"move_player_id"`
	ownerId      string    `db:"owner_player_id"`
	createdAt    time.Time `db:"created_at"`
}

type gameRow struct {
	Id           string    `db:"id"`
	Board        board     `db:"board"`
	GameType     string    `db:"type"`
	Status       string    `db:"status"`
	MovePlayerId string    `db:"move_player_id"`
	OwnerId      string    `db:"owner_player_id"`
	CreatedAt    time.Time `db:"created_at"`
}

type scoreboardRow struct {
	Id      string  `db:"id"`
	Login   string  `db:"login"`
	Winrate float32 `db:"winrate"`
}

type user struct {
	id       string `db:"id"`
	password string `db:"password"`
}
