package mediamgr

import (
  "errors"
  "fmt"
  "io"
  "log"
  "time"

  "github.com/ABFranco/roomz-media-server/rwebrtc"

  "github.com/pion/rtcp"
  "github.com/pion/webrtc/v3"
)

const (
  rtcpPLIInterval = time.Second * 3
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
  pc, err := webrtc.NewPeerConnection(rwebrtc.Config)
  if err != nil {
    log.Printf("Failed to create a peer connection for userId: %v", userId)
    return
  }
  if _, err = pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
    log.Printf("Failed to add transceiver for userId: %v", userId)
    return
  }
  videoTrackChan := make(chan *webrtc.TrackLocalStaticRTP)
  pc.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
    go func() {
      ticker := time.NewTicker(rtcpPLIInterval)
      for range ticker.C {
        if rtcpSendErr := pc.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())}}); rtcpSendErr != nil {
          log.Printf("RTCP Send Error: %v for userId: %v", rtcpSendErr, userId)
        }
      }
    }()

    localVideoTrack, newTrackErr := webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, "video", "pion")
    if newTrackErr != nil {
      log.Printf("Error creating new local track for userId: %v", userId)
    }
    videoTrackChan <- localVideoTrack

    rtpBuf := make([]byte, 1400)
    for {
      i, _, readErr := remoteTrack.Read(rtpBuf)
      if readErr != nil {
        log.Printf("read error for userId: %v", userId)
      }
      // TODO: SpatialAudio Filters.
      if _, err = localVideoTrack.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
        log.Printf("write error for userId: %v", userId)
      }
    }
  })
  r.broadcastVideoChannels[userId] = videoTrackChan
  peerId := fmt.Sprintf("%v-%v", userId, userId)
  r.pcs[peerId] = pc
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