package protocol

// Message types used over the TCP connection (newline-delimited JSON).
const (
	TypeStart = "start"
	TypeState = "state"
	TypeMove  = "move"
	TypeError = "error"
	TypeQuit  = "quit"
)

// Message is the shared wire format for both client->server and server->client.
// Fields are optional depending on the message type.
type Message struct {
	Type   string `json:"type"`
	Player string `json:"player,omitempty"`
	Row    int    `json:"row,omitempty"`
	Col    int    `json:"col,omitempty"`
	Board  string `json:"board,omitempty"`
	Turn   string `json:"turn,omitempty"`
	Status string `json:"status,omitempty"`
	Winner string `json:"winner,omitempty"`
	Error  string `json:"error,omitempty"`
}
