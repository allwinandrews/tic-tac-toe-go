# Tic-Tac-Toe over TCP (Go)

This project provides a minimal TCP server and CLI client for playing tic-tac-toe against another connected player.

## Architecture
- Server is a single instance that listens for TCP connections, pairs two clients, and runs one game session per pair.
- Client is a simple CLI that sends moves and prints game state.
- Communication is newline-delimited JSON messages (one JSON object per line).

## Walkthrough (end-to-end flow)
- Start the server; it accepts TCP connections and places each client into a matchmaking queue.
- When two clients are available, the server starts a session and assigns X to Player 1 and O to Player 2.
- The server sends a `start` message to each client, followed by an initial `state` with an empty board.
- Clients send `move` messages; the server validates, updates the game, and broadcasts an authoritative `state`.
- The session ends on win/draw, or becomes `abandoned` if a client disconnects.

## Protocol (NDJSON)
All messages have a `type` field. Strings use `X` or `O` for players.

Server -> Client
- `start`: assigns player symbol.
  - `{ "type": "start", "player": "X" }`
- `state`: authoritative board and turn.
  - `{ "type": "state", "board": "X..O..X..", "turn": "O", "status": "in_progress" }`
  - `status` is one of `in_progress`, `win`, `draw`, `abandoned`
  - `winner` is set when `status == win`
- `error`: human-readable message.
  - `{ "type": "error", "error": "not your turn" }`

Client -> Server
- `move`: requests a move (0-based row/col).
  - `{ "type": "move", "row": 1, "col": 2 }`
- `quit`: disconnect intentionally.
  - `{ "type": "quit" }`

Board encoding: 9 characters, row-major, `X`, `O`, or `.` for empty.

## Build
Requires Go 1.22+.

```
go build ./cmd/server
go build ./cmd/client
```

## Run
Start server:
```
./server -addr :9000
```

Start two clients (in separate terminals):
```
./client -addr 127.0.0.1:9000
```

## Manual test
- Connect two clients.
- Enter moves as `row col` (0-2).
- Server validates turn order and occupancy, and broadcasts game state.
