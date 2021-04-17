package roommgr

import (
  "github.com/ABFranco/roomz-media-server/room"
)

type RoomManager struct {
  rooms map[int64] room.Room
}

func (r *RoomManager) getRoom(roomId int64) room.Room {
  return r.rooms[roomId]
}

func (r *RoomManager) deleteRoom(roomId int64) {
  if _, ok := r.rooms[roomId]; ok {
    delete(r.rooms, roomId)
  }
}