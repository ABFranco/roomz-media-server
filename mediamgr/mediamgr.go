package mediamgr

import (
  "github.com/pion/webrtc/v3"
)

type RoomMediaManager struct {
  roomId int64
  broadcastAudioChannels map[int64]chan(*webrtc.TrackLocalStaticRTP)
  broadcastVideoChannels map[int64]chan(*webrtc.TrackLocalStaticRTP)
  pcs map[int64]*webrtc.PeerConnection
}

func (r *RoomMediaManager) startBroastcast(userId int64) {
  // TODO
}

func (r *RoomMediaManager) completeBroadcast(userId int64, sdpOffer string) {
  // TODO
}

func (r *RoomMediaManager) recvBroastcast(toUserId, fromUserId int64, sdpOffer string) {
  // TODO
}

func (r *RoomMediaManager) muteAudio(userId int64) {
  // TODO
}

func (r *RoomMediaManager) muteVideo(userId int64) {
  // TODO
}

func (r *RoomMediaManager) remoteAudioBroastcast(userId int64) {
  // TODO
}

func (r *RoomMediaManager) remoteVideoBroastcast(userId int64) {
  // TODO
}

func (r *RoomMediaManager) unrecvAudioBroastcast(toUserId, fromUserId int64) {
  // TODO
}

func (r *RoomMediaManager) unrecvVideoBroastcast(toUserId, fromUserId int64) {
  // TODO
}