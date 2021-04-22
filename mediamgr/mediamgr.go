package mediamgr

import (
  "context"
  "errors"
  "fmt"
  "io"
  "log"
  "time"

  "github.com/ABFranco/roomz-media-server/rwebrtc"
  "github.com/ABFranco/roomz-media-server/rwebrtc/signal"

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
  log.Printf("Creating new broadcaster peer connection for userId: %v", userId)
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
    log.Printf("Creating new local video track for userId: %v", userId)
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
      // TODO: Spatial-Audio Filters.
      if _, err = localVideoTrack.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
        log.Printf("write error for userId: %v", userId)
      }
    }
  })
  log.Printf("Created broadcastVideoChannel for userId: %v", userId)
  r.broadcastVideoChannels[userId] = videoTrackChan
  peerId := fmt.Sprintf("%v-%v", userId, userId)
  r.pcs[peerId] = pc
}

func (r *RoomMediaManager) CompleteBroadcast(peerId, sdpOffer string) (string, error) {
  pc, ok := r.pcs[fmt.Sprintf(peerId)]
  if !ok {
    return "", errors.New("cannot find user peer connection to complete broadcast for")
  }
  offer := webrtc.SessionDescription{}
  signal.Decode(sdpOffer, &offer)
  log.Printf("Setting remote description for peerId: %v", peerId)
  if err := pc.SetRemoteDescription(offer); err != nil {
    return "", err
  }
  log.Printf("Creating answer for peerId: %v", peerId)
  answer, err := pc.CreateAnswer(nil)
  if err != nil {
    return "", err
  }
  gatherComplete := webrtc.GatheringCompletePromise(pc)
  if err = pc.SetLocalDescription(answer); err != nil {
    return "", err
  }
  <-gatherComplete
  localDesc := signal.Encode(*pc.LocalDescription())

  // Spawn a goroutine to monitor ICE Connection changes.
  go func() {
    _, cancel := context.WithCancel(context.Background())
    pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
      log.Printf("peerId: %s connection state has changed: %s\n", peerId, connectionState.String())
      if connectionState == webrtc.ICEConnectionStateConnected {
        log.Printf("peerId: %s is connected.", peerId)
      } else if connectionState == webrtc.ICEConnectionStateFailed ||
              connectionState == webrtc.ICEConnectionStateDisconnected {
        log.Printf("peerId: %s is disconnected.", peerId)
        cancel()
      }
    })
  }()
  return localDesc, nil
}

func (r *RoomMediaManager) RecvBroastcast(toPeerId, fromPeerId, sdpOffer string) {
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