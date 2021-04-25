package mediamgr

import (
  "context"
  "errors"
  "fmt"
  "io"
  "log"
  "strings"
  "strconv"
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

// StartBroadcast creates a broadcaster peer connection which registers audio/
// video transceivers and sets an event handler to ensure when a track is
// officially registered on the pc, that media is routed to the correct
// audio/video channel. This peer connection still requires the offer/answer
// process, so we save the peer connection in a map.
func (r *RoomMediaManager) StartBroastcast(userId int64) {
  log.Printf("Creating new broadcaster peer connection for userId: %v", userId)
  pc, err := rwebrtc.NewRoomzPeerConnection()
  if err != nil {
    log.Printf("Failed to create a peer connection for userId: %v", userId)
    return
  }
  if _, err = pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
    log.Printf("Failed to add transceiver for userId: %v", userId)
    return
  }
  if _, err = pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
    log.Printf("Failed to add transceiver for userId: %v", userId)
    return
  }
  audioTrackChan := make(chan *webrtc.TrackLocalStaticRTP)
  videoTrackChan := make(chan *webrtc.TrackLocalStaticRTP)
  pc.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
    // I need to refresh my webrtc knowledge, but this is a requirement to
    // write RTCP packets on the peer connection. Please do not focus on
    // this for now.
    go func() {
      ticker := time.NewTicker(rtcpPLIInterval)
      for range ticker.C {
        if rtcpSendErr := pc.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())}}); rtcpSendErr != nil {
          log.Printf("RTCP Send Error: %v for userId: %v", rtcpSendErr, userId)
        }
      }
    }()
    // Depending on the remote track's codec MimeType, we can register a
    // TrackLocalStaticRTP instance, and send its data down the correct
    // audio/video track channel.
    var localVideoTrack *webrtc.TrackLocalStaticRTP
    var localAudioTrack *webrtc.TrackLocalStaticRTP
    var newTrackErr error
    if remoteTrack.Codec().MimeType == "video/VP8" {
      log.Printf("Creating new local video track for userId: %v", userId)
      localVideoTrack, newTrackErr = webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, "video", "pion")
      if newTrackErr != nil {
        log.Printf("Error creating new local track for userId: %v", userId)
      }
      videoTrackChan <- localVideoTrack
    } else {
      log.Printf("Creating new local audio track for userId: %v", userId)
      localAudioTrack, newTrackErr = webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, "audio", "pion")
      if newTrackErr != nil {
        log.Printf("Error creating new local track for userId: %v", userId)
      }
      audioTrackChan <- localAudioTrack
    }

    // Spawn a goroutine to read incoming data from this track and ensure that
    // data gets written to the correct audio/video track. Potentially, I do
    // not need a goroutine for this section.
    rtpBuf := make([]byte, 1400)
    go func() {
      log.Printf("Waiting for RTP data for userId: %v", userId)
      for {
        i, _, readErr := remoteTrack.Read(rtpBuf)
        if readErr != nil {
          log.Printf("read error for userId: %v", userId)
        }
  
        // TODO(hridayesh): Spatial-Audio Filters.
        if remoteTrack.Codec().MimeType == "video/VP8" {
          if localVideoTrack != nil {
            if _, err = localVideoTrack.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
              log.Printf("write error for userId: %v", userId)
            }
          } else {
            log.Printf("localVideoTrack is nil!")
          }
        } else {
          if localAudioTrack != nil {
            if _, err = localAudioTrack.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
              log.Printf("write error for userId: %v", userId)
            }
          } else {
            log.Printf("localAudioTrack is nil!")
          }
        }
      }
    }()
  })
  log.Printf("Created broadcastVideoChannel for userId: %v", userId)
  r.broadcastAudioChannels[userId] = audioTrackChan
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

// RecvBroadcast is how a peer asks to recv another RoomUser's audio/video
// data. It creates a fresh recvBroadcast pc and emits '
func (r *RoomMediaManager) RecvBroastcast(toPeerId, fromPeerId, sdpOffer string) (string, error) {
  fromUserIdStr := strings.Split(fromPeerId, "-")[1]
  fromUserId, _ := strconv.ParseInt(fromUserIdStr, 10, 64)
  fromAudioChannel := r.broadcastAudioChannels[fromUserId]
  fromVideoChannel := r.broadcastVideoChannels[fromUserId]
  // I'm interested to see how this would work when I do the mute experiment.
  log.Printf("peerId:%v receiving audio/video tracks from peerId:%v", toPeerId, fromPeerId)
  fromAudioTrack := <- fromAudioChannel
  fromVideoTrack := <- fromVideoChannel
  
  log.Printf("Creating RecvBroadcast pc for peerId:%v from peerId:%v", toPeerId, fromPeerId)
  recvBroadcastPc, err := rwebrtc.NewRoomzPeerConnection()
  if err != nil {
    return "", err
  }
  // Add audio and video tracks onto peer connection so that data can be sent
  // the RFE peer connection.
  for _, track := range []webrtc.TrackLocal{fromAudioTrack, fromVideoTrack} {
    rtpSender, err := recvBroadcastPc.AddTrack(track)
    if err != nil {
      return "", nil
    }

    // Again, ignore for now.
    go func() {
      rtcpBuf := make([]byte, 1500)
      for {
        if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
          return
        }
      }
    }()
  }

  offer := webrtc.SessionDescription{}
  signal.Decode(sdpOffer, &offer)
  err = recvBroadcastPc.SetRemoteDescription(offer)
  log.Printf("Set remote description - recvbroadcast")
  if err != nil {
    log.Printf("Failed to set remote description. err=%v", err)
    return "", nil
  }
  sdpAnswer, err := recvBroadcastPc.CreateAnswer(nil)
  if err != nil {
    log.Printf("Failed to create answer on recvBroadcast peer connection.")
    return "", nil
  }
  log.Printf("sdpAnswer:%v recvbroadcast", sdpAnswer)
  gatherComplete := webrtc.GatheringCompletePromise(recvBroadcastPc)
  err = recvBroadcastPc.SetLocalDescription(sdpAnswer)
  if err != nil {
    return "", nil
  }
  <-gatherComplete
  localDesc := signal.Encode(*recvBroadcastPc.LocalDescription())
  // Spawn a goroutine to monitor ICE Connection changes.
  go func() {
    _, cancel := context.WithCancel(context.Background())
    recvBroadcastPc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
      // Not really a prefix? Consider changing the name.
      prefix := fmt.Sprintf("recvBroadcastPc peerId: %v to receive from peerId:%s", toPeerId, fromPeerId)
      log.Printf("%s connection state has changed: %s\n", prefix, connectionState.String())
      if connectionState == webrtc.ICEConnectionStateConnected {
        log.Printf("%s is connected.", prefix)
      } else if connectionState == webrtc.ICEConnectionStateFailed ||
              connectionState == webrtc.ICEConnectionStateDisconnected {
        log.Printf("%s: is disconnected.", prefix)
        cancel()
      }
    })
  }()
  log.Printf("local desc: %v recvBroadcast", localDesc)
  // TODO: save RecvBroadcast pc in MediaMgr map.
  return localDesc, nil
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