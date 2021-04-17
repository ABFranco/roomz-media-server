package room

import (
  "github.com/ABFranco/roomz-media-server/mediamgr"
)

type Room struct {
  roomId   int64
  mediaMgr mediamgr.RoomMediaManager
  roomies  []RoomUser
}

type RoomUser struct {
  peerId int64
  sId    int64
}

func (r *Room) join(userId, sId int64) {
  // TODO
}

func (r *Room) leave(userId int64) {
  // TODO
}

func (r *Room) broadcastNewRoomyArrived(userId int64) {
  // TODO
}

func (r *Room) broadcastRoomyLeft(userId int64) {
  // TODO
}

func (r *Room) getRoomiez() []int64 {
  // TODO
  return []int64{}
}

func (r *Room) getMediaManager() mediamgr.RoomMediaManager {
  return r.mediaMgr
}