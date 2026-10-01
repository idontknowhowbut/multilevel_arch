package app

import (
	"errors"
	"fmt"
	"tictactoe/internal/domain"
	jwtservice "tictactoe/internal/jwt"
	"tictactoe/internal/store"

	"github.com/google/uuid"
)

var (
	ErrGameIsOver     = errors.New("game is already over")
	ErrWrongPlayer    = errors.New("wrong player move")
	ErrJoiningOwnGame = errors.New("attempt to join your own game")
	ErrWrongStatus    = errors.New("game already started")
)

type Service struct {
	repository *store.Repository
	Jwtservice *jwtservice.Jwtservice
}

func New(repository *store.Repository, jwt *jwtservice.Jwtservice) *Service {
	return &Service{
		repository: repository,
		Jwtservice: jwt,
	}
}

func (s *Service) ProcessMove(id string, currBoard domain.Board, userId string) (game domain.Game, err error) {
	currGame, err := s.repository.Load(id)
	if err != nil {
		return domain.Game{}, err
	}

	userIdUuid, err := uuid.Parse(userId)
	if currGame.MovePlayerId != userIdUuid {
		fmt.Println(currGame.MovePlayerId)
		fmt.Println(userIdUuid)
		return domain.Game{}, ErrWrongPlayer
	}

	_, isOver := domain.IsGameOver(currGame.Board)
	if isOver {
		return currGame, ErrGameIsOver
	}

	err = domain.ValidateMove(currBoard, currGame.Board)
	if err != nil {
		return currGame, err
	}

	isOwnerMove := userId == currGame.OwnerId.String()

	for i := range currBoard {
		for j, l := range currBoard[i] {
			if l != currGame.Board[i][j] {
				if isOwnerMove {
					currBoard[i][j] = 1
				} else {
					currBoard[i][j] = 2
				}
			}
		}
	}

	currGame.Board = currBoard
	winner, isOver := domain.IsGameOver(currBoard)
	if isOver {
		if winner == 0 {
			currGame.Status = domain.StatusDraw
		} else {
			currGame.Status = domain.StatusPlayerWon
		}
		s.repository.Save(currGame)
		return currGame, nil
	}

	if currGame.Type == domain.TypePVE {
		currBoard = domain.GetNextMove(currBoard)
		currGame.Board = currBoard

		winner, isOver := domain.IsGameOver(currBoard)
		if isOver {
			if winner == 0 {
				currGame.Status = domain.StatusDraw
			} else {
				currGame.Status = domain.StatusPlayerWon
			}
			s.repository.Save(currGame)
			return currGame, nil
		}
	}

	if currGame.Type == domain.TypePVP {
		nextUserId, err := s.repository.GetNextUser(currGame.Id.String())
		if err != nil {
			return currGame, err
		}
		nextUserIdUuid, _ := uuid.Parse(nextUserId)
		currGame.MovePlayerId = nextUserIdUuid
	}

	s.repository.Save(currGame)

	return currGame, nil
}

func (s *Service) CreateGame(userId string, gameType domain.GameType) (game domain.Game, err error) {
	newGame := *domain.New(userId, gameType)

	switch gameType {
	case domain.TypePVE:
		newGame.Status = domain.StatusPlayerMove
		newGame.MovePlayerId, _ = uuid.Parse(userId)
	case domain.TypePVP:
		newGame.Status = domain.StatusWaitingForPlayer
		newGame.MovePlayerId, _ = uuid.Parse(userId)
	default:
		return domain.Game{}, errors.New("invalid game type")
	}

	err = s.repository.SaveNewGame(newGame, userId)
	if err != nil {
		return domain.Game{}, err
	}
	return newGame, nil
}

func (s *Service) SignUpUser(login string, pass string) error {
	err := s.repository.CreateUser(login, pass)
	return err
}

func (s *Service) GetAvailableGames(userId uuid.UUID) ([]domain.Game, error) {
	games, err := s.repository.GetAvailableGames(userId.String())
	return games, err
}

func (s *Service) JoinGame(userId uuid.UUID, gameId uuid.UUID) error {

	currGame, err := s.repository.Load(gameId.String())
	if currGame.OwnerId == userId {
		return ErrJoiningOwnGame
	}

	if currGame.Status != domain.StatusWaitingForPlayer {
		return ErrWrongStatus
	}

	err = s.repository.JoinGame(userId.String(), gameId.String())
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) GetGame(gameId uuid.UUID) (domain.Game, error) {
	currGame, err := s.repository.Load(gameId.String())
	if err != nil {
		return domain.Game{}, err
	}
	return currGame, nil
}

func (s *Service) GetUserInfo(userId string) (userLogin string, err error) {
	userLogin, err = s.repository.GetUserLogin(userId)
	return
}

func (s *Service) Authenticate(login string, pass string) (string, error) {
	userId, err := s.repository.CheckUser(login, pass)
	if err != nil {
		return "", err
	}

	return userId, nil
}

func (s *Service) CreateTokenPair(userId string) (tokenAccess string, tokenRefresh string, err error) {
	tokenAccess, err = s.Jwtservice.CreateToken(userId, jwtservice.Access)
	if err != nil {
		return "", "", err
	}

	tokenRefresh, err = s.Jwtservice.CreateToken(userId, jwtservice.Refresh)
	if err != nil {
		return "", "", err
	}

	return tokenAccess, tokenRefresh, nil
}

func (s *Service) GetUserFinishedGames(userId uuid.UUID) ([]domain.Game, error) {
	games, err := s.repository.GetUserFinishedGames(userId.String())
	return games, err
}

func (s *Service) GetScoreBoard(limit int) ([]domain.UserScore, error) {
	scoreboard, err := s.repository.GetScoreBoard(limit)
	return scoreboard, err
}
