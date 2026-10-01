package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage() *Storage {
	newPool, err := pgxpool.New(context.Background(), "postgresql://go_service_user:69420@localhost:5432/tic_tac_toe")
	if err != nil {
		fmt.Printf("Failed to connect to DB %s", err)
		return nil
	}

	return &Storage{
		pool: newPool,
	}
}

func (s *Storage) save(storeGame game) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	boardJson, _ := json.Marshal(storeGame.board)

	query := `
	INSERT INTO games (id, board, status, move_player_id)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id) DO UPDATE SET board = $2, status = $3, move_player_id = $4
	`

	_, err := s.pool.Exec(ctx, query, storeGame.id, string(boardJson), storeGame.status, storeGame.movePlayerId)
	return err
}

func (s *Storage) load(id string) (game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	query := `SELECT id, board, status, type, move_player_id, owner_player_id, created_at
		FROM games
		WHERE id = $1`

	res, err := s.pool.Query(ctx, query, id)

	gameRow, err := pgx.CollectOneRow(res, pgx.RowToStructByNameLax[gameRow])

	storeGame := game{
		id:           gameRow.Id,
		board:        gameRow.Board,
		gameType:     gameRow.GameType,
		status:       gameRow.Status,
		movePlayerId: gameRow.MovePlayerId,
		ownerId:      gameRow.OwnerId,
		createdAt:    gameRow.CreatedAt,
	}

	if err != nil {
		return game{}, err
	}

	return storeGame, nil
}

func (s *Storage) checkUser(login string, pass string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	res, err := s.pool.Query(ctx, "SELECT id FROM users WHERE login = $1 AND password = $2", login, pass)
	if err != nil {
		return "", err
	}

	userId, err := pgx.CollectOneRow(res, pgx.RowTo[string])
	if err != nil {
		return "", err
	}

	return userId, err
}

func (s *Storage) createUser(login string, pass string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	newId := uuid.New()
	_, err := s.pool.Exec(ctx, "INSERT INTO users (id, login, password) VALUES ($1, $2, $3)", newId, login, pass)

	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) saveNewGame(storeGame game, userId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	boardJson, _ := json.Marshal(storeGame.board)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Failed to save new game: %w", err)
	}

	defer tx.Rollback(ctx)

	const saveGameQuery = "INSERT INTO games (id, board, status, type, move_player_id, owner_player_id) VALUES ($1, $2, $3, $4, $5, $5)"
	_, err = tx.Exec(ctx, saveGameQuery, storeGame.id, string(boardJson), storeGame.status, storeGame.gameType, storeGame.movePlayerId)
	if err != nil {
		return fmt.Errorf("Failed to save new game: %w", err)
	}

	const assignPlayerQuery = "INSERT INTO users_games (user_id, game_id, role) VALUES ($1, $2, $3)"
	_, err = tx.Exec(ctx, assignPlayerQuery, userId, storeGame.id, "OWNER")
	if err != nil {
		return fmt.Errorf("Failed to save new game: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("Failed to save new game: %w", err)
	}
	return err
}

func (s *Storage) getNextUserId(gameId string) (userId string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	res, err := s.pool.Query(ctx, "SELECT user_id FROM users_games WHERE game_id = $1 AND user_id::text != (SELECT move_player_id FROM games WHERE id = $1)", gameId)

	if err != nil {
		return "", err
	}

	userId, err = pgx.CollectOneRow(res, pgx.RowTo[string])
	if err != nil {
		return "", err
	}

	return userId, err
}

func (s *Storage) getAvailableGames(userId string) ([]gameRow, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	query := `SELECT id, board, status, type, move_player_id, owner_player_id, created_at
		FROM games
		WHERE status = 'WAITING_FOR_PLAYER'
  		AND owner_player_id::uuid != $1::uuid
		UNION
		SELECT id, board, status, type, move_player_id, owner_player_id, created_at
		FROM games
        JOIN users_games ON games.id = users_games.game_id
		WHERE status = 'PLAYER_MOVE'
  		AND user_id = $1::uuid`

	res, err := s.pool.Query(ctx, query, userId)
	games, err := pgx.CollectRows(res, pgx.RowToStructByNameLax[gameRow])
	fmt.Print(err)
	return games, err
}

func (s *Storage) joinGame(userId string, gameId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	tx, err := s.pool.Begin(ctx)

	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "UPDATE games SET status = 'PLAYER_MOVE' WHERE id = $1", gameId)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO users_games (user_id, game_id, role) VALUES ($1, $2, $3)", userId, gameId, "GUEST")
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) getUserLogin(userId string) (userLogin string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	query := `SELECT login
		FROM users
		WHERE id = $1::uuid`

	err = s.pool.QueryRow(ctx, query, userId).Scan(&userLogin)
	if err != nil {
		return "", err
	}

	return
}

func (s *Storage) updateRefreshToken(userId string, refreshToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	query := `INSERT INTO refresh_tokens (user_id, token)
	VALUES ($1, $2)
	ON CONFLICT (user_id) DO UPDATE SET token = $2`

	_, err := s.pool.Exec(ctx, query, userId, refreshToken)
	return err
}

func (s *Storage) GetUserFinishedGames(userId string) ([]gameRow, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	query := `SELECT id, board, status, type, move_player_id, owner_player_id, created_at
		FROM games
		JOIN users_games on games.id = users_games.game_id
		WHERE status IN ('PLAYER_WON', 'BOT_WON', 'DRAW') and user_id = $1`

	res, err := s.pool.Query(ctx, query, userId)
	games, err := pgx.CollectRows(res, pgx.RowToStructByNameLax[gameRow])
	fmt.Print(err)
	return games, err
}

func (s *Storage) GetScoreBoard(limit int) ([]scoreboardRow, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	query := `SELECT
 		users.id,
 		users.login,
 		(COUNT(CASE WHEN status = 'PLAYER_WON' AND move_player_id = users.id THEN 1 END) * 1.0)
 		/
  		COUNT(*) AS winrate
		FROM
		users
		JOIN users_games ON users.id = users_games.user_id
		JOIN games ON games.id = users_games.game_id
		GROUP BY
		users.id,
		users.login
		ORDER BY
		winrate DESC
		LIMIT $1;
		`

	res, err := s.pool.Query(ctx, query, limit)
	scoreboard, err := pgx.CollectRows(res, pgx.RowToStructByNameLax[scoreboardRow])
	fmt.Print(err)
	return scoreboard, err
}
