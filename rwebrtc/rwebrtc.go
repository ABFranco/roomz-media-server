package rwebrtc

import (
  "context"
  "log"

  "github.com/ABFranco/roomz-media-server/rwebrtc/signal"

  "github.com/pion/interceptor"
  "github.com/pion/webrtc/v3"
)

var (
  Config = webrtc.Configuration{
    ICEServers: []webrtc.ICEServer{
      {
        URLs: []string{"stun:stun.l.google.com:19302"},
      },
    },
  }
)

type MediaConfig struct {
  WantVideo bool
  WantAudio bool
}

// NewRoomzPeerConnection cretes a fresh WebRTC connection.
func NewRoomzPeerConnection() (*webrtc.PeerConnection, error) {
  m := &webrtc.MediaEngine{}
  if err := registerCodecs(m); err != nil {
    panic(err)
  }
  i := &interceptor.Registry{}
  if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
    panic(err)
  }
  api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i))
  return api.NewPeerConnection(Config)
}

// RegisterTracks takes a config object, and determines which media tracks to
// register onto a peer connection.
func RegisterTracks(pc *webrtc.PeerConnection, config MediaConfig) ([]*webrtc.TrackLocalStaticRTP, error) {
  var mediaSenders []*webrtc.RTPSender
  var tracks []*webrtc.TrackLocalStaticRTP
  if config.WantAudio {
    audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion")
    if err != nil {
      return nil, err
    }
    audioSender, err := pc.AddTrack(audioTrack)
    if err != nil {
      return nil, err
    }
    tracks = append(tracks, audioTrack)
    mediaSenders = append(mediaSenders, audioSender)
  }
  if config.WantVideo {
    videoTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion")
    if err != nil {
      return nil, err
    }
    videoSender, err := pc.AddTrack(videoTrack)
    if err != nil {
      return nil, err
    }
    tracks = append(tracks, videoTrack)
    mediaSenders = append(mediaSenders, videoSender)
  }
  for _, sender := range mediaSenders {
    // Read incoming RTCP packets.
    // Before these packets are retuned they are processed by interceptors. For things
    // like NACK this needs to be called.
    go func(sender *webrtc.RTPSender) {
      rtcpBuf := make([]byte, 1500)
      for {
        if _, _, rtcpErr := sender.Read(rtcpBuf); rtcpErr != nil {
          return
        }
      }
    }(sender)
  }
  return tracks, nil
}

// AwaitIceChanges handles any change in the ICEConnection on a peer
// connection.
// TODO: If connection breaks, delete roomTracks corresponding to
// peer.
func AwaitIceChanges(pc *webrtc.PeerConnection) {
  _, cancel := context.WithCancel(context.Background())
  pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
    log.Printf("Connection state has changed: %s\n", connectionState.String())
    if connectionState == webrtc.ICEConnectionStateConnected {
      log.Println("Ctrl+C the remote client to stop the demo")
    } else if connectionState == webrtc.ICEConnectionStateFailed ||
            connectionState == webrtc.ICEConnectionStateDisconnected {
      log.Println("Done forwarding")
      cancel()
    }
  })
}

func SendAnswer(pc *webrtc.PeerConnection, sdp string) (string, error) {
  offer := webrtc.SessionDescription{}
  signal.Decode(sdp, &offer)
  if err := pc.SetRemoteDescription(offer); err != nil {
    return "", err
  }
  answer, err := pc.CreateAnswer(nil)
  if err != nil {
    return "", err
  }
  gatherComplete := webrtc.GatheringCompletePromise(pc)
  if err = pc.SetLocalDescription(answer); err != nil {
    return "", err
  }
  <-gatherComplete
  return signal.Encode(*pc.LocalDescription()), nil
}

func registerCodecs(m *webrtc.MediaEngine) error {
  if err := m.RegisterCodec(webrtc.RTPCodecParameters{
    RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/VP8", ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
  }, webrtc.RTPCodecTypeVideo); err != nil {
    return err
  }
  if err := m.RegisterCodec(webrtc.RTPCodecParameters{
    RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
  }, webrtc.RTPCodecTypeAudio); err != nil {
    return err
  }
  return nil
}