package roommgr

import (
  "github.com/ABFranco/roomz-media-server/room"
)

type RoomManager struct {
  rooms map[int64] *room.Room
}

func NewRoomManager() *RoomManager {
  return &RoomManager{
    rooms: make(map[int64]*room.Room),
  }
}

func (r *RoomManager) GetRoom(roomId int64) *room.Room {
  if _, ok := r.rooms[roomId]; !ok {
    r.rooms[roomId] = room.NewRoom(roomId)
  }
  return r.rooms[roomId]
}

func (r *RoomManager) DeleteRoom(roomId int64) {
  if _, ok := r.rooms[roomId]; ok {
    delete(r.rooms, roomId)
  }
}