package server

import (
	"bufio"
	"encoding/json"
	"net"
	"testing"
	"time"

	"tic-tac-toe-go/internal/game"
	"tic-tac-toe-go/internal/protocol"
)

func TestSessionPlayToWin(t *testing.T) {
	c1, s1 := net.Pipe()
	c2, s2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	p1 := newClientConn(s1)
	p2 := newClientConn(s2)

	done := make(chan struct{})
	go func() {
		runSession(p1, p2)
		close(done)
	}()

	dec1 := json.NewDecoder(bufio.NewReader(c1))
	dec2 := json.NewDecoder(bufio.NewReader(c2))
	enc1 := json.NewEncoder(c1)
	enc2 := json.NewEncoder(c2)

	// Start messages
	msg1 := readMsg(t, c1, dec1)
	msg2 := readMsg(t, c2, dec2)
	if msg1.Type != protocol.TypeStart || msg2.Type != protocol.TypeStart {
		t.Fatalf("expected start messages, got %v and %v", msg1.Type, msg2.Type)
	}
	if msg1.Player != string(game.PlayerX) || msg2.Player != string(game.PlayerO) {
		t.Fatalf("unexpected players: %s and %s", msg1.Player, msg2.Player)
	}

	// Initial state
	state1 := readMsg(t, c1, dec1)
	state2 := readMsg(t, c2, dec2)
	if state1.Type != protocol.TypeState || state2.Type != protocol.TypeState {
		t.Fatalf("expected state messages, got %v and %v", state1.Type, state2.Type)
	}

	// X wins across top row.
	moves := []struct {
		enc *json.Encoder
		row int
		col int
	}{
		{enc1, 0, 0},
		{enc2, 1, 1},
		{enc1, 0, 1},
		{enc2, 2, 2},
		{enc1, 0, 2},
	}

	var last1, last2 protocol.Message
	for _, mv := range moves {
		if err := mv.enc.Encode(protocol.Message{Type: protocol.TypeMove, Row: mv.row, Col: mv.col}); err != nil {
			t.Fatalf("encode move: %v", err)
		}
		last1 = readMsg(t, c1, dec1)
		last2 = readMsg(t, c2, dec2)
	}

	if last1.Status != game.StatusWin || last2.Status != game.StatusWin {
		t.Fatalf("expected win status, got %s and %s", last1.Status, last2.Status)
	}
	if last1.Winner != string(game.PlayerX) || last2.Winner != string(game.PlayerX) {
		t.Fatalf("expected X to win, got %s and %s", last1.Winner, last2.Winner)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("session did not finish")
	}
}

func readMsg(t *testing.T, conn net.Conn, dec *json.Decoder) protocol.Message {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var msg protocol.Message
	if err := dec.Decode(&msg); err != nil {
		t.Fatalf("decode message: %v", err)
	}
	return msg
}
