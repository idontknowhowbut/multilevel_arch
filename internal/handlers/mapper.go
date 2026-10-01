package handlers

import (
	"tictactoe/internal/domain"
)

func fromDomainToHandler(domainBoard domain.Board) (handlerBoard Board) {
	handlerBoard = Board(domainBoard)
	return handlerBoard
}

func fromHandlerToDomain(handlerBoard Board) (domainBoard domain.Board) {
	domainBoard = domain.Board(handlerBoard)
	return domainBoard
}
