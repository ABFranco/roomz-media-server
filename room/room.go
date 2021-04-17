package room

import (
  "github.com/ABFranco/roomz-media-server/mediamgr"
)

type Room struct {
  roomId   int64
  mediaMgr *mediamgr.RoomMediaManager
  roomies  []RoomUser
}

type RoomUser struct {
  peerId int64
  sId    int64
}

func NewRoom(roomId int64) *Room {
  r := &Room{
    roomId: roomId,
    mediaMgr: mediamgr.NewRoomMediaManager(roomId),
    roomies: []RoomUser{},
  }
  return r
}

func (r *Room) Join(userId, sId int64) {
  // TODO
}

func (r *Room) Leave(userId int64) {
  // TODO
}

func (r *Room) BroadcastNewRoomyArrived(userId int64) {
  // TODO
}

func (r *Room) BroadcastRoomyLeft(userId int64) {
  // TODO
}

func (r *Room) GetRoomiez() []int64 {
  // TODO
  return []int64{}
}

func (r *Room) GetMediaManager() *mediamgr.RoomMediaManager {
  return r.mediaMgr
}