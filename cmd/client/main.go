package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"tic-tac-toe-go/internal/protocol"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:9000", "server address")
	flag.Parse()

	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close()

	log.Printf("connected to %s", *addr)

	var myPlayer string
	var currentTurn string
	var status string

	done := make(chan struct{})
	go func() {
		defer close(done)
		// Read and process server messages.
		scanner := bufio.NewScanner(conn)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 64*1024)
		for scanner.Scan() {
			var msg protocol.Message
			if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
				log.Printf("invalid server message")
				continue
			}
			switch msg.Type {
			case protocol.TypeStart:
				myPlayer = msg.Player
				playerTag := "Player 2"
				if myPlayer == "X" {
					playerTag = "Player 1"
				}
				fmt.Printf("You are %s (%s)\n", myPlayer, playerTag)
				fmt.Println("How to play: enter moves as `row col` (0-2). Example: 1 2")
				fmt.Println("Goal: get three in a row horizontally, vertically, or diagonally.")
			case protocol.TypeState:
				// Server sends authoritative state after each move.
				status = msg.Status
				currentTurn = msg.Turn
				printBoard(msg.Board)
				if msg.Status == "win" {
					fmt.Printf("Winner: %s\n", msg.Winner)
					return
				}
				if msg.Status == "draw" {
					fmt.Println("Draw.")
					return
				}
				if msg.Status == "abandoned" {
					fmt.Printf("Game ended: %s\n", msg.Error)
					return
				}
				if myPlayer != "" && myPlayer == currentTurn {
					fmt.Print("Your move (row col): ")
				} else if myPlayer != "" {
					fmt.Printf("Opponent's turn (%s). Please wait...\n", currentTurn)
				}
			case protocol.TypeError:
				fmt.Printf("Error: %s\n", msg.Error)
			}
		}
	}()

	enc := json.NewEncoder(conn)
	input := bufio.NewScanner(os.Stdin)
	// Read user moves from stdin and send to server.
	for {
		if !input.Scan() {
			break
		}
		line := strings.TrimSpace(input.Text())
		if line == "" {
			continue
		}
		if line == "quit" {
			_ = enc.Encode(protocol.Message{Type: protocol.TypeQuit})
			break
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			fmt.Println("enter move as: row col (0-2)")
			continue
		}
		row, err1 := strconv.Atoi(fields[0])
		col, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil {
			fmt.Println("enter move as: row col (0-2)")
			continue
		}
		if status != "in_progress" {
			fmt.Println("game is not in progress")
			continue
		}
		_ = enc.Encode(protocol.Message{Type: protocol.TypeMove, Row: row, Col: col})
	}

	<-done
}

func printBoard(board string) {
	if len(board) < 9 {
		fmt.Println("board unavailable")
		return
	}
	fmt.Println("Legend: X (Player 1), O (Player 2)")
	fmt.Println("Board:")
	for r := 0; r < 3; r++ {
		row := board[r*3 : r*3+3]
		fmt.Printf(" %s\n", row)
	}
}
