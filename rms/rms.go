package rms

import (
  "log"

  socketio "github.com/googollee/go-socket.io"
)

type RoomzMediaServer struct {
  Server        *socketio.Server
}

func (r *RoomzMediaServer) Init() *RoomzMediaServer {
  server, err := socketio.NewServer(nil)
  if err != nil {
    panic(err)
  }
  rms := &RoomzMediaServer{
    Server:        server,
  }
  rms.routes()
  return rms
}

func (r *RoomzMediaServer) routes() {
  r.Server.OnConnect("/", r.connectHandler)
  r.Server.OnDisconnect("/", r.disconnectHandler)
  r.Server.OnEvent("/", "JoinMediaRoom", r.joinMediaRoomHandler)
  r.Server.OnEvent("/", "ReceiveMediaFrom", r.receiveMediaFromHandler)
  r.Server.OnEvent("/", "LeaveMediaRoom", r.leaveMediaRoomHandler)
  r.Server.OnEvent("/", "OnIceCandidate", r.onIceCandidateHandler)
}

func (r *RoomzMediaServer) connectHandler(s socketio.Conn) error {
  s.SetContext("")
  log.Printf("New user (ID=%v) connected...", s.ID())
  return nil
}

func (r *RoomzMediaServer) disconnectHandler(s socketio.Conn, msg string) {
  log.Println("User:", s.ID(), "disconnected...");
}

func (r *RoomzMediaServer) joinMediaRoomHandler(s socketio.Conn, data map[string]interface{}) {
}

func (r *RoomzMediaServer) receiveMediaFromHandler(s socketio.Conn, data map[string]interface{}) {
}

func (r *RoomzMediaServer) leaveMediaRoomHandler(s socketio.Conn, data map[string]interface{}) {
}

func (r *RoomzMediaServer) onIceCandidateHandler(s socketio.Conn, data map[string]interface{}) {
}