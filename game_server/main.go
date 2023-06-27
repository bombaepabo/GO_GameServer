package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"

	"github.com/anthdm/hollywood/actor"
	"github.com/bombaepabo/gameserver/types"
	"github.com/gorilla/websocket"
)

type PlayerSession struct {
	sessionID int
	clientID  int
	username  string
	intLobby  bool
	conn      *websocket.Conn
}

func newPlayerSession(sid int, conn *websocket.Conn) actor.Producer {
	return func() actor.Receiver {
		return &PlayerSession{
			conn:      conn,
			sessionID: sid,
		}

	}
}
func (s *PlayerSession) Receive(c *actor.Context) {
	switch c.Message().(type) {
	case actor.Started:
		s.readLoop()
	}
}
func (s *PlayerSession) readLoop() {
	var msg types.WSMessage
	for {
		if err := s.conn.ReadJSON(&msg); err != nil {
			s.handleMessage(msg)
			fmt.Println("read error", err)
			return
		}
		go s.handleMessage(msg)
	}
}
func (s *PlayerSession) handleMessage(msg types.WSMessage) {
	switch msg.Type {
	case "Login":
		var loginMsg types.Login
		if err := json.Unmarshal(msg.Data, &loginMsg); err != nil {
			panic(err)
		}
		s.clientID = loginMsg.ClientID
		s.username = loginMsg.Username
	}
}

type GameServer struct {
	ctx      *actor.Context
	sessions map[*actor.PID]struct{}
}

func newGameServer() actor.Receiver {
	return &GameServer{
		sessions: make(map[*actor.PID]struct{}),
	}
}
func (s *GameServer) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case actor.Started:
		s.startHTTP()
		s.ctx = c
		_ = msg
	}

}
func (s *GameServer) startHTTP() {
	fmt.Println("starting HTTP server on port 40000")
	http.HandleFunc("/ws", s.handleWS)
	http.ListenAndServe(":40000", nil)
}

func (s *GameServer) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		fmt.Println("ws upgrade err:", err)
		return
	}
	fmt.Printf("new client trying to connect")
	sid := rand.Intn(math.MaxInt)
	ps := newPlayerSession(sid, conn)
	pid := s.ctx.SpawnChild(ps, fmt.Sprintf("session_%d", sid))
	s.sessions[pid] = struct{}{}
}
func main() {

	e := actor.NewEngine()
	e.Spawn(newGameServer, "server")

}
