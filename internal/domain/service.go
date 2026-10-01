package domain

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidMoveCellValue     = errors.New("Invalid Move: Incorrect cell value")
	ErrInvalidMoveCellNotEmpty  = errors.New("Invalid Move: Cell is already taken")
	ErrInvalidMoveMultipleMoves = errors.New("Invalid Move: Only one move was expected")
	ErrWrongPlayerMove          = errors.New("Invalid Move: Wrong player")
)

func GetNextMove(currBoard Board) (bestMove Board) {
	var bestRow, bestCol int
	bestMove = currBoard
	bestScore := int(math.Inf(-1))

	for i, row := range currBoard {
		for j, cell := range row {
			if cell == 0 {
				currBoard[i][j] = 2
				moveScore := MiniMax(currBoard, 0, false)
				currBoard[i][j] = 0
				if moveScore > bestScore {
					bestRow = i
					bestCol = j
					bestScore = moveScore
				}
			}
		}
	}
	bestMove[bestRow][bestCol] = 2
	return bestMove
}

func MiniMax(currBoard Board, depth int, isMax bool) (score int) {
	if winner, isOver := IsGameOver(currBoard); isOver {
		switch winner {
		case 0:
			score = 0
		case 1:
			score = depth - 10
		case 2:
			score = 10 - depth
		}
		return score
	}
	if isMax {
		bestScore := -1000
		for i, row := range currBoard {
			for j, cell := range row {
				if cell == 0 {
					currBoard[i][j] = 2
					miniMaxScore := MiniMax(currBoard, depth+1, !isMax)
					bestScore = max(bestScore, miniMaxScore)
					currBoard[i][j] = 0
				}
			}
		}
		return bestScore

	} else {
		bestScore := 1000
		for i, row := range currBoard {
			for j, cell := range row {
				if cell == 0 {
					currBoard[i][j] = 1
					miniMaxScore := MiniMax(currBoard, depth+1, !isMax)
					bestScore = min(bestScore, miniMaxScore)
					currBoard[i][j] = 0
				}
			}
		}
		return bestScore
	}
}

func ValidateMove(currBoard Board, prevBoard Board) error {
	changeCount := 0
	for i, row := range currBoard {
		for j, cell := range row {
			prevCell := prevBoard[i][j]
			if cell != prevCell {
				if prevCell == 0 {
					if cell == 1 || cell == 2 {
						changeCount++
					} else {
						return ErrInvalidMoveCellValue
					}
				} else {
					return ErrInvalidMoveCellNotEmpty
				}
			}
		}
	}
	if changeCount != 1 {
		return ErrInvalidMoveMultipleMoves
	} else {
		return nil
	}
}

func IsGameOver(currBoard Board) (winner int, isGameOver bool) {
	type position struct {
		row int
		col int
	}

	type winLine [3]position

	winLines := [...]winLine{
		{{row: 0, col: 0}, {row: 0, col: 1}, {row: 0, col: 2}},
		{{row: 1, col: 0}, {row: 1, col: 1}, {row: 1, col: 2}},
		{{row: 2, col: 0}, {row: 2, col: 1}, {row: 2, col: 2}},

		{{row: 0, col: 0}, {row: 1, col: 0}, {row: 2, col: 0}},
		{{row: 0, col: 1}, {row: 1, col: 1}, {row: 2, col: 1}},
		{{row: 0, col: 2}, {row: 1, col: 2}, {row: 2, col: 2}},

		{{row: 0, col: 0}, {row: 1, col: 1}, {row: 2, col: 2}},
		{{row: 2, col: 0}, {row: 1, col: 1}, {row: 0, col: 2}},
	}

	isGameOver = false
	winner = 0

	for _, line := range winLines {
		cell := currBoard[line[0].row][line[0].col]
		if cell == 0 {
			continue
		} else if cell == currBoard[line[1].row][line[1].col] &&
			cell == currBoard[line[2].row][line[2].col] {
			winner = cell
			isGameOver = true
			return
		}
	}

	if isGameOver == false {
		for _, row := range currBoard {
			for _, cell := range row {
				if cell == 0 {
					return 0, false
				}
			}
		}
	}
	return 0, true
}

func New(userId string, gameType GameType) *Game {
	userIdUuid, err := uuid.Parse(userId)
	if err != nil {
		return nil
	}

	return &Game{
		Id:        uuid.New(),
		Board:     Board{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
		OwnerId:   userIdUuid,
		Type:      gameType,
		Status:    StatusCreated,
		CreatedAt: time.Now(),
	}

}
