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
  roomMgr   *roommgr.RoomManager
}

func (r *RoomzMediaServer) Init() *RoomzMediaServer {
  server, err := socketio.NewServer(nil)
  if err != nil {
    panic(err)
  }
  rms := &RoomzMediaServer{
    SioServer:        server,
    roomMgr:          roommgr.NewRoomManager(),
  }
  rms.routes()
  return rms
}

func (r *RoomzMediaServer) routes() {
  r.SioServer.OnConnect("/", r.connectHandler)
  r.SioServer.OnDisconnect("/", r.disconnectHandler)
  r.SioServer.OnEvent("/", "JoinMediaRoom", r.joinMediaRoomHandler)
  r.SioServer.OnEvent("/", "ReceiveMediaFrom", r.receiveMediaFromHandler)
  r.SioServer.OnEvent("/", "CompleteBroadcastOffer", r.completeBroadcastOfferHandler)
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
  userIdStr, ok := data["user_id"].(string)
  if !ok {
    log.Printf("invalid user_id")
    return
  }
  // Unfortunately, data["xxx"].(int64) cannot properly convert the input data
  // and we need to perform the cast after converting to a string.
  userId, err := strconv.ParseInt(userIdStr, 10, 64)
  if err != nil {
    log.Printf("user_id is NaN")
    return
  }
  roomIdStr, ok := data["room_id"].(string)
  if !ok {
    log.Printf("invalid room_id")
    return
  }
  roomId, err := strconv.ParseInt(roomIdStr, 10, 64)
  if err != nil {
    log.Printf("room_id is NaN")
    return
  }
  // TODO: Validate token.
  log.Printf("Getting room %v from the RoomManager", roomId)
  room := r.roomMgr.GetRoom(roomId)
  mediaMgr := room.GetMediaManager()
  log.Printf("Starting broadcast for userId: %v", userId)
  mediaMgr.StartBroastcast(userId)

  // Before the user joins the room, emit the userId's of all current
  // RoomUsers.
  log.Printf("Emitting \"ExistingMediaRoomiez\" to userId: %v", userId)
  room.SendExistingRoomiez(r.SioServer, userId, s.ID())

  room.Join(userId, s.ID())

  // Tell all RoomUsers (including the new roomy) that a new roomy
  // has arrived.
  log.Printf("Emitting \"NewMediaRoomyArrived\" to roomId: %v for userId: %v", roomId, userId)
  room.BroadcastNewRoomyArrived(r.SioServer, userId)
}

func (r *RoomzMediaServer) receiveMediaFromHandler(s socketio.Conn, data map[string]interface{}) {
  log.Printf(":ReceiveMediaFrom: received data=%v", data)
  // NOTE: peer ID's follow the format: <room_id>-<user_id>
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
  var sdpAnswer string
  // Create a RecvBroadcast Pc and send its sdpAnswer to the RFE.
  sdpAnswer, err = mediaMgr.RecvBroastcast(toPeerId, fromPeerId, sdpOffer)
  if err != nil {
    log.Printf("RecvBroadcast failed, err=%v", err)
    return
  }
  r.SioServer.BroadcastToRoom("/", s.ID(), "ReceiveMediaAnswer", map[string]interface{}{
    "from_peer_id": fromPeerId,
    "sdp_answer": sdpAnswer,
  })
  log.Printf("peerId: %v receiving broadcast from peerId: %v", toPeerId, fromPeerId)
}

func (r *RoomzMediaServer) completeBroadcastOfferHandler(s socketio.Conn, data map[string]interface{}) {
  log.Printf(":CompleteBroadcastOffer: received data=%v", data)
  // NOTE: peer ID's follow the format: <room_id>-<user_id>
  peerId, ok := data["peer_id"].(string)
  if !ok {
    log.Printf("invalid peer_id")
    return
  }
  sdpOffer, ok := data["desc"].(string)
  if !ok {
    log.Printf("no sdpOffer")
    return
  }
  roomId := strings.Split(peerId, "-")[0]
  roomId64, err := strconv.ParseInt(roomId, 10, 64)
  if err != nil {
    log.Printf("could not parse peer_id correctly")
    return
  }
  room := r.roomMgr.GetRoom(roomId64)
  mediaMgr := room.GetMediaManager()
  sdpAnswer, err := mediaMgr.CompleteBroadcast(peerId, sdpOffer)
  if err != nil {
    log.Printf("Completing broadcast failed for peerId: %v, err=%v", peerId, err)
    return
  }
  r.SioServer.BroadcastToRoom("/", s.ID(), "CompleteBroadcastAnswer", map[string]interface{}{
    "peer_id": peerId,
    "sdp_answer": sdpAnswer,
  })
}

func (r *RoomzMediaServer) leaveMediaRoomHandler(s socketio.Conn, data map[string]interface{}) {
}

func (r *RoomzMediaServer) onIceCandidateHandler(s socketio.Conn, data map[string]interface{}) {
}