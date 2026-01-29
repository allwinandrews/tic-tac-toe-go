package game

import "testing"

func TestWinRow(t *testing.T) {
	g := New()
	moves := [][3]int{
		{0, 0, int(PlayerX)},
		{1, 0, int(PlayerO)},
		{0, 1, int(PlayerX)},
		{1, 1, int(PlayerO)},
		{0, 2, int(PlayerX)},
	}
	for _, m := range moves {
		if err := g.ApplyMove(m[0], m[1], rune(m[2])); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if g.Status != StatusWin || g.Winner != PlayerX {
		t.Fatalf("expected X to win, got status=%s winner=%c", g.Status, g.Winner)
	}
}

func TestInvalidTurn(t *testing.T) {
	g := New()
	if err := g.ApplyMove(0, 0, PlayerO); err == nil {
		t.Fatalf("expected turn validation error")
	}
}

func TestDraw(t *testing.T) {
	g := New()
	moves := [][3]int{
		{0, 0, int(PlayerX)},
		{0, 1, int(PlayerO)},
		{0, 2, int(PlayerX)},
		{1, 1, int(PlayerO)},
		{1, 0, int(PlayerX)},
		{1, 2, int(PlayerO)},
		{2, 1, int(PlayerX)},
		{2, 0, int(PlayerO)},
		{2, 2, int(PlayerX)},
	}
	for _, m := range moves {
		if err := g.ApplyMove(m[0], m[1], rune(m[2])); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if g.Status != StatusDraw {
		t.Fatalf("expected draw, got status=%s", g.Status)
	}
}
