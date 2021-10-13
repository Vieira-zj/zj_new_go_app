package pkg

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	isDebug = false
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Maximum message size allowed from peer.
	maxMessageSize = 8192
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Time to wait before force close on connection.
	closeGracePeriod = 10 * time.Second
	// EndOfTransmission end
	EndOfTransmission = "\u0004"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		if r.Method == "GET" {
			log.Println("websocket not support GET")
			return false
		}
		return true
	},
}

// TerminalMessage is the messaging protocol between ShellController and TerminalSession.
type TerminalMessage struct {
	// stdin: term to pod; resize: term to pod; stdout: pod to term;
	Operation string `json:"operation"`
	Data      string `json:"data"`
	Rows      uint16 `json:"rows"`
	Cols      uint16 `json:"cols"`
}

// TerminalSession implements PtyHandler, and handles pod executor stdin and stdout.
type TerminalSession struct {
	wsConn   *websocket.Conn
	sizeChan chan remotecommand.TerminalSize
	doneChan chan struct{}
}

// NewTerminalSessionWS creates TerminalSession.
func NewTerminalSessionWS(conn *websocket.Conn) *TerminalSession {
	return &TerminalSession{
		wsConn:   conn,
		sizeChan: make(chan remotecommand.TerminalSize),
		doneChan: make(chan struct{}),
	}
}

// NewTerminalSession creates TerminalSession.
func NewTerminalSession(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*TerminalSession, error) {
	conn, err := upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	return NewTerminalSessionWS(conn), nil
}

// Done .
func (term *TerminalSession) Done() chan struct{} {
	return term.doneChan
}

// Close closes terminal session.
func (term *TerminalSession) Close() error {
	return term.wsConn.Close()
}

// Next is called in a loop from remotecommand as long as the process is running.
func (term *TerminalSession) Next() *remotecommand.TerminalSize {
	myLogPrintln("[Next] terminal session")
	select {
	case size := <-term.sizeChan:
		myLogPrintln("[NextEnd] terminal is resized")
		return &size
	case <-term.doneChan:
		return nil
	}
}

// Read is called in a loop from remotecommand as long as the process is running.
// Read implements io.Reader (stdin), workflow:
// exec cmd in pod <= pod executor stdin <= TerminalSession read() <= via websocket ReadMessage() <= webshell term
func (term *TerminalSession) Read(p []byte) (int, error) {
	myLogPrintln("[Read] terminal session")
	_, message, err := term.wsConn.ReadMessage()
	myLogPrintln("[ReadEnd] read message from ws, and copy to stdin")
	if err != nil {
		log.Printf("ws read message err: %v\n", err)
		return copy(p, EndOfTransmission), err
	}

	var msg TerminalMessage
	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		log.Printf("json unmarshal ws message err: %v\n", err)
		return copy(p, EndOfTransmission), err
	}

	switch msg.Operation {
	case "stdin":
		return copy(p, msg.Data), nil
	case "resize":
		term.sizeChan <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
		return 0, nil
	default:
		log.Printf("unknown message type '%s'\n", msg.Operation)
		return copy(p, EndOfTransmission), fmt.Errorf("unknown message type '%s'", msg.Operation)
	}
}

// Write is called from remotecommand whenever there is any output from stdout.
// Write implements io.Writer (stdout), workflow:
// pod cmd output => pod executor stdout => TerminalSession write() => via websocket WriteMessage() => webshell term
func (term *TerminalSession) Write(p []byte) (int, error) {
	myLogPrintln("[Write] terminal session")
	msg, err := json.Marshal(TerminalMessage{
		Operation: "stdout",
		Data:      string(p),
	})
	if err != nil {
		log.Printf("json marshal stdin bytes error: %v\n", err)
		return 0, err
	}

	if err := term.wsConn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Printf("ws write message error: %v\n", err)
		return 0, err
	}
	myLogPrintln("[WriteEnd] copy message from stdout, and write to ws")
	return len(p), nil
}

func myLogPrintln(msg string) {
	if isDebug {
		log.Println(msg)
	}
}
