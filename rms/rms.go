package rms

import (
  "log"

  "github.com/ABFranco/roomz-media-server/roommgr"
  socketio "github.com/googollee/go-socket.io"
)

type RoomzMediaServer struct {
  SioServer *socketio.Server
  roomMgr   roommgr.RoomManager
}

func (r *RoomzMediaServer) Init() *RoomzMediaServer {
  server, err := socketio.NewServer(nil)
  if err != nil {
    panic(err)
  }
  rms := &RoomzMediaServer{
    SioServer:        server,
  }
  rms.routes()
  return rms
}

func (r *RoomzMediaServer) routes() {
  r.SioServer.OnConnect("/", r.connectHandler)
  r.SioServer.OnDisconnect("/", r.disconnectHandler)
  r.SioServer.OnEvent("/", "JoinMediaRoom", r.joinMediaRoomHandler)
  r.SioServer.OnEvent("/", "ReceiveMediaFrom", r.receiveMediaFromHandler)
  r.SioServer.OnEvent("/", "LeaveMediaRoom", r.leaveMediaRoomHandler)
  r.SioServer.OnEvent("/", "OnIceCandidate", r.onIceCandidateHandler)
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
  log.Printf(":JoinMediaRoom: received data=%v", data)
  userId, ok := data["user_id"].(int64)
  if !ok {
    log.Printf("invalid user_id")
    return
  }
  roomId, ok := data["room_id"].(int64)
  if !ok {
    log.Printf("invalid room_id")
    return
  }
  // TODO: Validate token.
  room := r.roomMgr.GetRoom(roomId)
  mediaMgr := room.GetMediaManager()
  log.Printf("Starting broadcast for userId: %v", userId)
  mediaMgr.StartBroastcast(userId)
  room.Join(userId, s.ID())

  log.Printf("Emitting \"NewMediaRoomyArrived\" to roomId: %v for userId: %v", roomId, userId)
  // TODO: Pass socketio Server here?
  room.BroadcastNewRoomyArrived(r.SioServer, userId)

  log.Printf("Emitting \"ExistingMediaRoomiez\" to userId: %v", userId)
  room.SendExistingRoomiez(r.SioServer, userId)
}

func (r *RoomzMediaServer) receiveMediaFromHandler(s socketio.Conn, data map[string]interface{}) {
}

func (r *RoomzMediaServer) leaveMediaRoomHandler(s socketio.Conn, data map[string]interface{}) {
}

func (r *RoomzMediaServer) onIceCandidateHandler(s socketio.Conn, data map[string]interface{}) {
}