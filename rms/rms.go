package rms

import (
  "log"
  "strconv"
  "strings"

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
  room.BroadcastNewRoomyArrived(r.SioServer, userId)

  log.Printf("Emitting \"ExistingMediaRoomiez\" to userId: %v", userId)
  room.SendExistingRoomiez(r.SioServer, userId)
}

func (r *RoomzMediaServer) receiveMediaFromHandler(s socketio.Conn, data map[string]interface{}) {
  log.Printf(":ReceiveMediaFrom: received data=%v", data)
  fromPeerId, ok := data["from_peer_id"].(string)
  if !ok {
    log.Printf("invalid from_peer_id")
    return
  }
  toPeerId, ok := data["to_peer_id"].(string)
  if !ok {
    log.Printf("invalid to_peer_id")
    return
  }
  sdpOffer, ok := data["desc"].(string)
  if !ok {
    log.Printf("no sdpOffer")
    return
  }
  roomId := strings.Split(fromPeerId, "-")[0]
  roomId64, err := strconv.ParseInt(roomId, 10, 64)
  if err != nil {
    log.Printf("could not parse peer_id correctly")
    return
  }
  room := r.roomMgr.GetRoom(roomId64)
  mediaMgr := room.GetMediaManager()
  if fromPeerId == toPeerId {
    log.Printf("Completing broadcast for peerId: %v", fromPeerId)
    sdpAnswer, err := mediaMgr.CompleteBroadcast(fromPeerId, sdpOffer)
    if err != nil {
      log.Printf("Completing broadcast failed, err=%v", err)
    }
    // NOTE: I'll move this chunk after the if-else after playback success.
    r.SioServer.BroadcastToRoom("/", s.ID(), "ReceiveMediaAnswer", map[string]interface{}{
      "fromPeerId": fromPeerId,
      "sdp_answer": sdpAnswer,
    })
  } else {
    log.Printf("peerId: %v receiving broadcast from peerId: %v", toPeerId, fromPeerId)
    mediaMgr.RecvBroastcast(toPeerId, fromPeerId, sdpOffer)
  }
}

func (r *RoomzMediaServer) leaveMediaRoomHandler(s socketio.Conn, data map[string]interface{}) {
}

func (r *RoomzMediaServer) onIceCandidateHandler(s socketio.Conn, data map[string]interface{}) {
}