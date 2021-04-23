import React, { useEffect, useRef } from 'react';

import * as rmsClient from '../api/RoomzMediaServerClient.js'
import './App.css';



function App(props) {
  const roomIdRef = useRef();
  const userIdRef = useRef();
  const egressMediaRef = useRef();
  const ingressMediaRef = useRef();

  var ICE_SERVERS = [
    {urls:"stun:stun.l.google.com:19302"}
  ];
  var roomyPcs = {};
  let videos = [
    { peer_id: "0-0" },
  ];
  let videoRefs = [];
  var myPeerId = "";
  let egressMediaStream = null;

  useEffect(() => {
    rmsClient.askToConnect()
  })
  
  // setupLocalMedia requests access to the user's microphone and webcam and
  // properly sets up the egress media stream.
  // NOTE: This will likely be called on load within the vestibule component.
  function setupLocalMediaUtil(cb, eb) {
    if (egressMediaStream != null) {
      if (cb) cb();
      return
    }
    console.log('asking for local audio/video inputs')
    navigator.getUserMedia = (navigator.getUserMedia ||
      navigator.webkitGetUserMedia ||
      navigator.mozGetUserMedia ||
      navigator.msGetUserMedia);
    
    // TODO: Pass config to mute audio/video.
    navigator.getUserMedia({"audio": true, "video": true},
      function(stream) {
        console.log('granted access to audio/video')
        egressMediaStream = stream
        egressMediaRef.current.srcObject = egressMediaStream
        if (cb) cb();
      },
      function() {
        console.log('access denied for audio/video')
        alert('have fun being lame on zoom')
        if (eb) eb();
      });
  }

  function setupLocalMedia() {
    setupLocalMediaUtil(() => {
      console.log('successfully setup local media')
    })
  }

  function newPeerConnection() {
    return new RTCPeerConnection(
      {"iceServers": ICE_SERVERS},
      // This is needed for chrome/firefox/edge support.
      {"optional": [{"DtlsSrtpKeyAgreement": true}]}
    )
  }

  // completeBroadcast creates a broadcaster pc, adds the local media tracks
  // onto the broadcaster pc, and performs the offer/answer process with the
  // RMS. This should be called before emitting a "RecvMediaFrom" event
  // asking to receive their own media.
  function completeBroadcast(peerId) {
    console.log('completing broadcast for peerId=%o', peerId)
    let broadcastPc = newPeerConnection();
    let tracks = egressMediaStream.getTracks()
    for (var i = 0; i < tracks.length; i++) {
      broadcastPc.addTrack(tracks[i], egressMediaStream)
    }
    
    broadcastPc.onicecandidate = function(event) {
      console.log('received ICE Candidate for broadcast pc for peerId=%o', peerId)
      console.log(event)
    }

    // Store peer connection in map.
    roomyPcs[peerId] = broadcastPc;

    broadcastPc.createOffer(
      function(localDescription) {
        console.log('set local description for peerId=%o\'s broadcaster pc.', peerId)
        broadcastPc.setLocalDescription(localDescription,
        function() {
          let data = {
            'peer_id': peerId,
            'desc': btoa(JSON.stringify(broadcastPc.localDescription))
          }
          rmsClient.completeBroadcastOffer(data, () => {
            console.log('peerId=%o send CompleteBroadcastOffer event', peerId)
            rmsClient.awaitCompleteBroadcastAnswer((resp) => {
              let sdpAnswer = JSON.parse(atob(resp["sdp_answer"]));
              if (sdpAnswer !== '') {
                var remoteDescription = new RTCSessionDescription(sdpAnswer)
                var tmp = broadcastPc.setRemoteDescription(remoteDescription,
                function() {
                  console.log('set remote description on peerId=%o\'s broadcast pc', peerId);
                }, function (e) {
                  console.log('error=%o setting remote description for peerId\'s broadcast pc', e, peerId)
                })
              }
            })
          })
        })
      },
      function(e) {
        console.log('error=%o setting peerId=%o\'s local description', e, peerId)
      }
    )

  }

  // joinMediaRoom emits the 'JoinMediaRoom' event to the RMS and registers
  // event handlers for possible response events from the RMS.
  function joinMediaRoom() {
    // TODO: validation.
    let roomId = roomIdRef.current.value;
    let userId = userIdRef.current.value;
    myPeerId = roomId + "-" + userId;
    let data = {
      'user_id': userId,
      'room_id': roomId,
    }
    rmsClient.joinMediaRoom(data, () => {
      console.log('userId: %o joined media room: %o', userId, roomId);
      rmsClient.awaitExistingMediaRoomiez((resp) => {
        console.log('received existing media roomiez, resp=%o', resp)
        // TODO: iterate through each roomy and request to receive media from
        // them.
      })
      rmsClient.awaitNewMediaRoomyArrived((resp) => {
        console.log('new roomy arrived, resp=%o', resp)
        let newPeerId = resp["peer_id"];
        let isBroadcaster = false;

        // Create a fresh peer connection.
        let recvBroadcastPc = newPeerConnection()

        if (newPeerId === myPeerId) {
          completeBroadcast(myPeerId)
        }

        // Store this peer connection in RoomyPeerConnection map
        let recvKey = newPeerId + "to" + myPeerId
        roomyPcs[recvKey] = recvBroadcastPc;

        // Setup handlers for when we receive data back on this peer connection.
        recvBroadcastPc.ontrack = function(event) {
          let fromUserId = newPeerId.split("-")[1];
          if (event.streams.length > 0 && fromUserId >= 0) {
            console.log('setting up media for userId=%o', fromUserId)
            ingressMediaRef.current.srcObject = event.streams[0]
            // TODO: Add to grid component somehow.
            // Keep video refs for now?
          }
        }

        // Setup handler to monitor ICE candidates we can use on the peer
        // connection.
        recvBroadcastPc.onicecandidate = function(event) {
          console.log('received ICE candidate from newPeerId=%o', newPeerId)
          console.log(event)
        }

        // Add an offer on the peer connection and after setting the local
        // description of the peer connection, emit the 'ReceiveMediaFrom'
        // event using the SDP.
        console.log('creating offer..')
        recvBroadcastPc.createOffer(
          function(localDescription) {
            console.log('set local description for newPeerId=%o', newPeerId)
            recvBroadcastPc.setLocalDescription(localDescription,
            function() {
              // With the local description, we can send the event. We will
              // await a 'ReceiveMediaAnswer' event and set the remote
              // description on this peer connection.
              let data = {
                'from_peer_id': newPeerId,
                'to_peer_id':   myPeerId,
                'desc':         btoa(JSON.stringify(recvBroadcastPc.localDescription))
              }
              rmsClient.receiveMediaFrom(data, () => {
                console.log('peerId=%o requested to receive media from peerId=%o', myPeerId, newPeerId);
                rmsClient.awaitReceiveMediaAnswer((resp) => {
                  console.log('received media answer resp=%o for peerId=%o', resp, newPeerId);
                  // TODO: validate.
                  let sdpAnswer = JSON.parse(atob(resp["sdp_answer"]));
                  if (sdpAnswer !== '') {
                    var remoteDescription = new RTCSessionDescription(sdpAnswer);
                    var tmp = recvBroadcastPc.setRemoteDescription(remoteDescription,
                    function() {
                      console.log('set remote description for peerId=%o', newPeerId);
                    }, function (e) {
                      console.log('error=%o setting remote description for peerId=%o', e, newPeerId);
                    });
                    console.log('remote description=%o for peerId=%o', newPeerId)
                  }
                });
              });
            });
          },
          function(e) {
            console.log('error setting local description=%o', e)
          }
        );
      });
    });
  }

  return (
    <div className="App">
      <div className="header">
        <h1>Roomz Media Server Playground</h1>
      </div>
      <div className="vestibule-video-preview-ctnr">
        <div className="video" id="vestibule-video-preview">
          <label htmlFor="outgoing-vid">Vestibule Video</label><br/>
          <video ref={egressMediaRef} id="vestibule-vid" autoPlay controls/>
        </div>
        <div className="user-actions">
          <button className="roomz-btn button-primary" onClick={setupLocalMedia}>Setup Media</button>
        </div>
      </div>
      <div className="user-settings">
        <form className="user-settings-form">
          <div className="user-input-form">
            <label htmlFor="room-id">Room Id: </label>
            <input id="room-id" ref={roomIdRef} autoFocus/>
          </div>
          <div className="user-input-form">
            <label htmlFor="user-id">User Id: </label>
            <input id="user-id" ref={userIdRef} autoFocus/>
          </div>
        </form>
        <div className="user-actions">
          <button className="roomz-btn button-primary" onClick={joinMediaRoom}>JoinMediaRoom</button>
        </div>
      </div>
      <div className="grid-cntr">
        <div className="video" id="ingress-media-ref">
          <label htmlFor="incoming-vid">Ingress Video</label><br/>
          <video ref={ingressMediaRef} id="ingress-vid" autoPlay controls/>
        </div>
      </div>
    </div>
  );
}

export default App;
