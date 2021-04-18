import React, { useEffect, useRef } from 'react';
import io from 'socket.io-client';
import './App.css';



function App(props) {
  const room_id = useRef();
  const user_id = useRef();
  const outgoing_video = useRef();
  const incoming_video = useRef();
  const incoming_video_0 = useRef();
  const incoming_video_1 = useRef();
  const incoming_video_2 = useRef();
  let vidRefs = [
    incoming_video_0,
    incoming_video_1,
    incoming_video_2
  ]
  let nConnections = 0;

  var ICE_SERVERS = [
    {urls:"stun:stun.l.google.com:19302"}
  ];
  var peers = {};
  let videos = [
    { peer_id: "0-0" },
  ];
  let videoRefs = [];
  var myUserId = "";
  var myRoomId = "";
  var myPeerId = "";
  let outgoingMediaStream = null;

  socket.on('connect', () => {
    console.log(":connect: CONNECTED to server.")
  })
  socket.on('disconnect', () => {
    socket.removeAllListeners()
    console.log(':disconnect: DISCONNECTED from server.');
  });
  useEffect(() => {
    askToConnect()
  })
  
  // request to access audio/video
  function setupLocalMedia(cb, eb) {
    if (outgoingMediaStream != null) {
      if (cb) cb();
      return
    }
    console.log('asking for local audio/video inputs')
    navigator.getUserMedia = (navigator.getUserMedia ||
      navigator.webkitGetUserMedia ||
      navigator.mozGetUserMedia ||
      navigator.msGetUserMedia);

    navigator.getUserMedia({"audio": true, "video": true},
      function(stream) {
        console.log('granted access to audio/video')
        outgoingMediaStream = stream
        outgoing_video.current.srcObject = outgoingMediaStream
        if (cb) cb();
      },
      function() {
        console.log('access denied for audio/video')
        alert('you must succomb')
        if (eb) eb();
      });
  }

  // send room_id and user_id info upon joining, hold onto peerId
  function joinRoom() {
    let roomId = room_id.current.value;
    let userId = user_id.current.value;
    setupLocalMedia(function() {
      console.log('sending mediaJoin')
      socket.emit('mediaJoin', {
        'user_id': userId,
        'room_id': roomId,
      })
      myUserId = userId
      myRoomId = roomId
      myPeerId = roomId + "-" + userId
    })
  }
  
  // respond to incoming peers
  socket.on('addPeer', (data) => {
    console.log('addPeer: ', data)
    let incomingPeerId = data["peer_id"];
    // establish webtrc connection for peer
    let pc = new RTCPeerConnection(
      {"iceServers": ICE_SERVERS},
      {"optional": [{"DtlsSrtpKeyAgreement": true}]}
    )

    // register audio/video tracks if needed
    if (data["is_og"]) {
      let tracks = outgoingMediaStream.getTracks()
      console.log('setting outgoing tracks: ', tracks)
      for (var i = 0; i < tracks.length; i++) {
        pc.addTrack(tracks[i], outgoingMediaStream)
      }
    }

    // save peer connection for later reference
    peers[incomingPeerId] = pc;

    pc.ontrack = function(event) {
      let trackUserId = incomingPeerId.split("-")[1];
      if (event.streams.length > 0 && trackUserId >= 0) {
        console.log('setting track for track user id: ', trackUserId)
        vidRefs[trackUserId].current.srcObject = event.streams[0]
      }
    }
     // if null ice candidate, create offer of sdp
    pc.onicecandidate = function(event) {
      console.log('ice candidate')
      if (event.candidate === null) {
        console.log('sending odpOffer here?')
        // socket.emit('sdpOffer', {
        //   'peer_id': myPeerId,
        //   'desc': btoa(JSON.stringify(pc.localDescription))
        // })
      }
    }
    // build offer to connect
    pc.createOffer(
      function(localDescription) {
        console.log("local description is: ", localDescription)
        pc.setLocalDescription(localDescription,
          function() {
            console.log('offer setLocalDescription succeeded')
            console.log('sending first odpOffer')
            socket.emit('sdpOffer', {
              'peer_id': myPeerId,
              'desc': btoa(JSON.stringify(pc.localDescription))
            })
          },
          function() {
            alert("offer setLocalDescription failed!")
          })
      },
      function (e) {
        console.log("error sending offer: ", e);
      })
  })

  socket.on('sdpAnswer', (data) => {
    console.log('sdpAnswer: ', data)
    var remoteDescription = JSON.parse(atob(data.desc))
    var pc = peers[data.peer_id]
    if (remoteDescription !== '') {
      var desc = new RTCSessionDescription(remoteDescription)
      var stuff = pc.setRemoteDescription(desc,
        function() {
          nConnections += 1
          console.log("setRemoteDescription succeeded");
        },
        function(e) {
          console.log("setRemoteDescription error: ", e)
        }
      );
      console.log("description object: ", desc);
    }
  })

  return (
    <div className="App">
      <div className="header">
        <h1>Roomz Media Server Playground</h1>
      </div>
      <div className="videos">
        {/* <div className="video" id="outgoing">
          <label htmlFor="outgoing-vid">Outgoing Video</label><br/>
          <video ref={outgoing_video} id="outgoing-vid" autoPlay controls/>
        </div>
        <div className="video" id="incoming">
          <label htmlFor="incoming-vid">Incoming Video</label><br/>
          <video ref={incoming_video} id="incoming-vid" autoPlay controls/>
        </div> */}
        {/* {videos.map((v, i) => (
          <div className="video" id="incoming">
            <label htmlFor="incoming-vid">Video {v.peer_id}</label><br/>
            <video key={i} value={v.peer_id} ref={ref => videoRefs[i] = ref} id="incoming-vid" autoPlay controls/>
          </div>
        ))} */}
      </div>
      <div className="user-settings">
        <form className="user-settings-form">
          <div className="user-input-form">
            <label htmlFor="room-id">Room Id: </label>
            <input id="room-id" ref={room_id} autoFocus/>
          </div>
          <div className="user-input-form">
            <label htmlFor="user-id">User Id: </label>
            <input id="user-id" ref={user_id} autoFocus/>
          </div>
        </form>
        <div className="user-actions">
          <button className="roomz-btn button-primary" onClick={joinRoom}>JoinMediaRoom</button>
        </div>
      </div>
    </div>
  );
}

export default App;
