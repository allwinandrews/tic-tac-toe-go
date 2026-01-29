package game

import (
	"errors"
)

const (
	Empty   = '.'
	PlayerX = 'X'
	PlayerO = 'O'
)

const (
	StatusInProgress = "in_progress"
	StatusWin        = "win"
	StatusDraw       = "draw"
	StatusAbandoned  = "abandoned"
)

type Game struct {
	board  [9]rune
	Moves  int
	Next   rune
	Status string
	Winner rune
}

func New() *Game {
	// Start with an empty board and X to move.
	g := &Game{
		Next:   PlayerX,
		Status: StatusInProgress,
	}
	for i := range g.board {
		g.board[i] = Empty
	}
	return g
}

func (g *Game) BoardString() string {
	// Row-major encoding of the board for transmission to clients.
	out := make([]rune, len(g.board))
	copy(out, g.board[:])
	return string(out)
}

func (g *Game) ApplyMove(row, col int, player rune) error {
	// Enforce game rules and update status/turn.
	if g.Status != StatusInProgress {
		return errors.New("game is already finished")
	}
	if player != g.Next {
		return errors.New("not your turn")
	}
	if row < 0 || row > 2 || col < 0 || col > 2 {
		return errors.New("move out of bounds")
	}
	idx := row*3 + col
	if g.board[idx] != Empty {
		return errors.New("cell already occupied")
	}
	g.board[idx] = player
	g.Moves++
	if winner := g.checkWinner(); winner != Empty {
		g.Status = StatusWin
		g.Winner = winner
		return nil
	}
	if g.Moves == 9 {
		g.Status = StatusDraw
		return nil
	}
	g.Next = Opponent(player)
	return nil
}

func Opponent(player rune) rune {
	// Toggle player token.
	if player == PlayerX {
		return PlayerO
	}
	return PlayerX
}

func (g *Game) checkWinner() rune {
	// Check all 8 winning lines.
	lines := [8][3]int{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{0, 3, 6},
		{1, 4, 7},
		{2, 5, 8},
		{0, 4, 8},
		{2, 4, 6},
	}
	for _, line := range lines {
		a, b, c := line[0], line[1], line[2]
		if g.board[a] != Empty && g.board[a] == g.board[b] && g.board[a] == g.board[c] {
			return g.board[a]
		}
	}
	return Empty
}
