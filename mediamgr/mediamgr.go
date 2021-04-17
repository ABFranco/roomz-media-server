package mediamgr

import (
  "github.com/pion/webrtc/v3"
)

type RoomMediaManager struct {
  roomId int64
  broadcastAudioChannels map[int64]chan(*webrtc.TrackLocalStaticRTP)
  broadcastVideoChannels map[int64]chan(*webrtc.TrackLocalStaticRTP)
  pcs map[string]*webrtc.PeerConnection
}

func NewRoomMediaManager(roomId int64) *RoomMediaManager {
  return &RoomMediaManager{
    roomId: roomId,
    broadcastAudioChannels: make(map[int64]chan(*webrtc.TrackLocalStaticRTP)),
    broadcastVideoChannels: make(map[int64]chan(*webrtc.TrackLocalStaticRTP)),
    pcs: make(map[string]*webrtc.PeerConnection),
  }
}

func (r *RoomMediaManager) StartBroastcast(userId int64) {
  // TODO
}

func (r *RoomMediaManager) CompleteBroadcast(userId int64, sdpOffer string) {
  // TODO
}

func (r *RoomMediaManager) RecvBroastcast(toUserId, fromUserId int64, sdpOffer string) {
  // TODO
}

func (r *RoomMediaManager) MuteAudio(userId int64) {
  // TODO
}

func (r *RoomMediaManager) MuteVideo(userId int64) {
  // TODO
}

func (r *RoomMediaManager) RemoteAudioBroastcast(userId int64) {
  // TODO
}

func (r *RoomMediaManager) RemoteVideoBroastcast(userId int64) {
  // TODO
}

func (r *RoomMediaManager) UnrecvAudioBroastcast(toUserId, fromUserId int64) {
  // TODO
}

func (r *RoomMediaManager) UnrecvVideoBroastcast(toUserId, fromUserId int64) {
  // TODO
}