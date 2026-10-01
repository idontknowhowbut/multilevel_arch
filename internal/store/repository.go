package store

import (
	"errors"
	"tictactoe/internal/domain"
)

type Repository struct {
	storage *Storage
}

var (
	ErrGameNotFound = errors.New("game not found in repo")
)

func NewRepository(s *Storage) *Repository {
	return &Repository{
		storage: s,
	}
}

func (r *Repository) Save(domainGame domain.Game) error {
	storeGame := fromDomainToStore(domainGame)
	err := r.storage.save(storeGame)
	return err
}

func (r *Repository) Load(id string) (domain.Game, error) {
	storeGame, err := r.storage.load(id)

	if err != nil {
		return domain.Game{}, err
	}
	domainGame := fromStoreToDomain(storeGame)
	return domainGame, nil
}

func (r *Repository) CheckUser(login string, pass string) (string, error) {
	userId, err := r.storage.checkUser(login, pass)
	return userId, err
}

func (r *Repository) CreateUser(login string, pass string) error {
	err := r.storage.createUser(login, pass)
	return err
}

func (r *Repository) SaveNewGame(domainGame domain.Game, userId string) error {
	storeGame := fromDomainToStore(domainGame)
	err := r.storage.saveNewGame(storeGame, userId)
	return err
}

func (r *Repository) GetNextUser(id string) (nextUserId string, err error) {
	return r.storage.getNextUserId(id)
}

func (r *Repository) GetAvailableGames(userId string) ([]domain.Game, error) {
	q, err := r.storage.getAvailableGames(userId)
	var res []domain.Game

	for _, v := range q {
		rowToAppend := game{
			id:           v.Id,
			board:        v.Board,
			gameType:     v.GameType,
			status:       v.Status,
			movePlayerId: v.MovePlayerId,
			ownerId:      v.OwnerId,
			createdAt:    v.CreatedAt,
		}

		res = append(res, fromStoreToDomain(rowToAppend))
	}

	return res, err
}

func (r *Repository) JoinGame(userId string, gameId string) error {
	err := r.storage.joinGame(userId, gameId)
	return err
}

func (r *Repository) GetUserLogin(userId string) (userLogin string, err error) {
	userLogin, err = r.storage.getUserLogin(userId)
	if err != nil {
		return "", err
	}
	return userLogin, err
}

func (r *Repository) UpdateRefreshToken(userId string, refreshToken string) (err error) {
	err = r.storage.updateRefreshToken(userId, refreshToken)
	return err
}

func (r *Repository) GetUserFinishedGames(userId string) ([]domain.Game, error) {
	q, err := r.storage.GetUserFinishedGames(userId)
	var res []domain.Game

	for _, v := range q {
		rowToAppend := game{
			id:           v.Id,
			board:        v.Board,
			gameType:     v.GameType,
			status:       v.Status,
			movePlayerId: v.MovePlayerId,
			ownerId:      v.OwnerId,
			createdAt:    v.CreatedAt,
		}

		res = append(res, fromStoreToDomain(rowToAppend))
	}

	return res, err

}

func (r *Repository) GetScoreBoard(limit int) ([]domain.UserScore, error) {
	q, err := r.storage.GetScoreBoard(limit)
	var res []domain.UserScore

	for _, v := range q {
		rowToAppend := domain.UserScore{
			UserId:    v.Id,
			UserLogin: v.Login,
			Winrate:   v.Winrate,
		}

		res = append(res, rowToAppend)
	}

	return res, err
}
