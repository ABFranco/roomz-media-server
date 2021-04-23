package room

import (
  "fmt"
  "log"

  "github.com/ABFranco/roomz-media-server/mediamgr"
  socketio "github.com/googollee/go-socket.io"
)

type Room struct {
  roomId   int64
  mediaMgr *mediamgr.RoomMediaManager
  roomies  []*RoomUser
}

type RoomUser struct {
  userId int64
  sId    string
}

func NewRoom(roomId int64) *Room {
  r := &Room{
    roomId: roomId,
    mediaMgr: mediamgr.NewRoomMediaManager(roomId),
    roomies: []*RoomUser{},
  }
  return r
}

func (r *Room) Join(userId int64, sId string) {
  newRoomUser := &RoomUser{
    userId: userId,
    sId:    sId,
  }
  r.roomies = append(r.roomies, newRoomUser)
}

func (r *Room) Leave(userId int64) {
  for i, roomUser := range r.roomies {
    if roomUser.userId == userId {
      r.roomies = append(r.roomies[:i], r.roomies[i+1:]...)
      break
    }
  }
}

func (r *Room) BroadcastNewRoomyArrived(server *socketio.Server, userId int64) {
  for _, roomUser := range r.roomies {
    log.Printf("Sending \"NewMediaRoomyArrived\" to userId: %v", roomUser.userId)
    server.BroadcastToRoom("/", roomUser.sId, "NewMediaRoomyArrived", map[string]interface{}{
      "peer_id": fmt.Sprintf("%v-%v", r.roomId, userId),
    })
  }
}

func (r *Room) BroadcastRoomyLeft(server *socketio.Server, userId int64) {
  for _, roomUser := range r.roomies {
    if roomUser.userId != userId {
      log.Printf("Sending \"MediaRoomyLeft\" to userId: %v", roomUser.userId)
      server.BroadcastToRoom("/", roomUser.sId, "MediaRoomyLeft", map[string]interface{}{
        "peer_id": fmt.Sprintf("%v-%v", r.roomId, userId),
      })
    }
  }
}

func (r *Room) SendExistingRoomiez(server *socketio.Server, userId int64, sId string) {
  var peerIds []string
  for _, roomUser := range r.roomies {
    peerIds = append(peerIds, fmt.Sprintf("%v-%v", r.roomId, roomUser.userId))
  }
  peerId := fmt.Sprintf("%v-%v", r.roomId, userId)
  log.Printf("Sending \"ExistingMediaRoomiez\" to peerId: %v", peerId)
  server.BroadcastToRoom("/", sId, "ExistingMediaRoomiez", map[string]interface{}{
    "peer_ids": peerIds,
  })
}

func (r *Room) GetRoomiez() []int64 {
  var userIds []int64
  for _, roomUser := range r.roomies {
    userIds = append(userIds, roomUser.userId)
  }
  return userIds
}

func (r *Room) GetMediaManager() *mediamgr.RoomMediaManager {
  return r.mediaMgr
}