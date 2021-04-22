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
  room.Join(userId, s.ID())

  log.Printf("Emitting \"NewMediaRoomyArrived\" to roomId: %v for userId: %v", roomId, userId)
  room.BroadcastNewRoomyArrived(r.SioServer, userId)

  log.Printf("Emitting \"ExistingMediaRoomiez\" to userId: %v", userId)
  room.SendExistingRoomiez(r.SioServer, userId)
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

  // If requesting to receive media from themselves, they must first complete
  // the offer/answer process on the broadcast pc.
  if fromPeerId == toPeerId {
    log.Printf("Completing broadcast for peerId: %v", fromPeerId)
    sdpAnswer, err = mediaMgr.CompleteBroadcast(fromPeerId, sdpOffer)
    if err != nil {
      log.Printf("Completing broadcast failed, err=%v", err)
      return
    }
  }
  // Create a RecvBroadcast Pc and send its sdpAnswer to the RFE.
  sdpAnswer, err = mediaMgr.RecvBroastcast(toPeerId, fromPeerId, sdpOffer)
  if err != nil {
    log.Printf("RecvBroadcast failed, err=%v", err)
    return
  }
  r.SioServer.BroadcastToRoom("/", s.ID(), "ReceiveMediaAnswer", map[string]interface{}{
    "fromPeerId": fromPeerId,
    "sdp_answer": sdpAnswer,
  })
  log.Printf("peerId: %v receiving broadcast from peerId: %v", toPeerId, fromPeerId)
}

func (r *RoomzMediaServer) leaveMediaRoomHandler(s socketio.Conn, data map[string]interface{}) {
}

func (r *RoomzMediaServer) onIceCandidateHandler(s socketio.Conn, data map[string]interface{}) {
}