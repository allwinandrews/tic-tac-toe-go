package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"tic-tac-toe-go/internal/game"
	"tic-tac-toe-go/internal/protocol"
)

type Server struct {
	Addr string
	mu   sync.Mutex
	ln   net.Listener
	// closing indicates a graceful shutdown to suppress accept errors.
	closing bool
}

func (s *Server) ListenAndServe() error {
	if s.Addr == "" {
		return errors.New("addr is required")
	}
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	s.mu.Lock()
	s.ln = ln
	s.mu.Unlock()
	defer ln.Close()

	// waiting is a simple matchmaking queue.
	waiting := make(chan *clientConn)
	go matchmaking(waiting)
	defer close(waiting)

	for {
		conn, err := ln.Accept()
		if err != nil {
			if s.isClosing(err) {
				return nil
			}
			return fmt.Errorf("accept: %w", err)
		}
		cc := newClientConn(conn)
		waiting <- cc
	}
}

// Shutdown stops accepting new connections and lets ListenAndServe return.
func (s *Server) Shutdown() error {
	s.mu.Lock()
	s.closing = true
	ln := s.ln
	s.mu.Unlock()
	if ln == nil {
		return nil
	}
	return ln.Close()
}

func (s *Server) isClosing(err error) bool {
	s.mu.Lock()
	closing := s.closing
	s.mu.Unlock()
	return closing && errors.Is(err, net.ErrClosed)
}

type clientConn struct {
	conn   net.Conn
	send   chan protocol.Message
	player rune
	once   sync.Once
}

func newClientConn(conn net.Conn) *clientConn {
	cc := &clientConn{
		conn: conn,
		send: make(chan protocol.Message, 16),
	}
	// Dedicated writer goroutine for each client.
	go cc.writeLoop()
	return cc
}

func (c *clientConn) close() {
	c.once.Do(func() {
		close(c.send)
	})
}

func (c *clientConn) sendMsg(msg protocol.Message) {
	// Non-blocking send; slow clients are disconnected.
	select {
	case c.send <- msg:
	default:
		c.close()
	}
}

func (c *clientConn) writeLoop() {
	writer := bufio.NewWriter(c.conn)
	enc := json.NewEncoder(writer)
	for msg := range c.send {
		// Keep the writer from hanging forever on a dead connection.
		_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := enc.Encode(msg); err != nil {
			break
		}
		if err := writer.Flush(); err != nil {
			break
		}
	}
	_ = c.conn.Close()
}

func matchmaking(waiting <-chan *clientConn) {
	var queue []*clientConn
	for cc := range waiting {
		queue = append(queue, cc)
		// Pair clients in FIFO order.
		for len(queue) >= 2 {
			p1 := queue[0]
			p2 := queue[1]
			queue = queue[2:]
			go runSession(p1, p2)
		}
	}
}

type move struct {
	player *clientConn
	row    int
	col    int
}

func runSession(p1, p2 *clientConn) {
	// One game per paired clients.
	g := game.New()
	p1.player = game.PlayerX
	p2.player = game.PlayerO

	p1.sendMsg(protocol.Message{Type: protocol.TypeStart, Player: string(p1.player)})
	p2.sendMsg(protocol.Message{Type: protocol.TypeStart, Player: string(p2.player)})
	broadcastState(g, p1, p2)

	moves := make(chan move)
	done := make(chan *clientConn, 2)

	// Reader goroutines forward moves and notify on disconnect.
	go readLoop(p1, moves, done)
	go readLoop(p2, moves, done)

	for {
		select {
		case mv := <-moves:
			if mv.player == nil {
				continue
			}
			// Apply the move and broadcast authoritative state.
			if err := g.ApplyMove(mv.row, mv.col, mv.player.player); err != nil {
				mv.player.sendMsg(protocol.Message{Type: protocol.TypeError, Error: err.Error()})
				continue
			}
			broadcastState(g, p1, p2)
			if g.Status != game.StatusInProgress {
				p1.close()
				p2.close()
				return
			}
		case quitter := <-done:
			// Notify the remaining player and end the session.
			other := p1
			if quitter == p1 {
				other = p2
			}
			other.sendMsg(protocol.Message{
				Type:   protocol.TypeState,
				Board:  g.BoardString(),
				Turn:   string(g.Next),
				Status: game.StatusAbandoned,
				Error:  "opponent disconnected",
			})
			time.Sleep(100 * time.Millisecond)
			p1.close()
			p2.close()
			return
		}
	}
}

func readLoop(c *clientConn, moves chan<- move, done chan<- *clientConn) {
	defer func() {
		done <- c
	}()
	scanner := bufio.NewScanner(c.conn)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 64*1024)
	for scanner.Scan() {
		var msg protocol.Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			c.sendMsg(protocol.Message{Type: protocol.TypeError, Error: "invalid message"})
			continue
		}
		switch msg.Type {
		case protocol.TypeMove:
			// Forward the requested move to the session loop.
			moves <- move{player: c, row: msg.Row, col: msg.Col}
		case protocol.TypeQuit:
			return
		default:
			c.sendMsg(protocol.Message{Type: protocol.TypeError, Error: "unknown message type"})
		}
	}
}

func broadcastState(g *game.Game, p1, p2 *clientConn) {
	msg := protocol.Message{
		Type:   protocol.TypeState,
		Board:  g.BoardString(),
		Turn:   string(g.Next),
		Status: g.Status,
	}
	if g.Status == game.StatusWin {
		msg.Winner = string(g.Winner)
	}
	p1.sendMsg(msg)
	p2.sendMsg(msg)
}
