
import io from 'socket.io-client';
// Connect to RMS
const rmsClientSocket = io("http://localhost:5000", {
  autoConnect: false,
  reconnection: true,
  reconnectionAttempts: 2,
  timeout: 10000 // timeout for each reconnection attempt
});

async function askToConnect() {
  try {
    await rmsClientSocket.connect()
    return true
  } catch (e) {
    console.error("Failed to create socket connection: %s", e)
    return false
  }
}

const events = {
  JOIN_MEDIA_ROOM: 'JoinMediaRoom',
  RECEIVE_MEDIA_FROM: 'ReceiveMediaFrom',
  EXISTING_MEDIA_ROOMIEZ: 'ExistingMediaRoomiez',
  NEW_MEDIA_ROOMY_ARRIVED: 'NewMediaRoomyArrived',
  RECEIVE_MEDIA_ANSWER: 'ReceiveMediaAnswer',
}

rmsClientSocket.on('disconnect', () => {
  console.log(':rms: DISCONNECTED from RMS');
})

rmsClientSocket.on('reconnect_failed', () => {
  console.log(':rms: Failed to reconnect. Closing socket.');
})

function joinMediaRoom(data, cb) {
  console.log(':rms.joinMediaRoom: Sending request to join media room, data=%o', data)
  rmsClientSocket.emit(events.JOIN_MEDIA_ROOM, data);
  cb()
}

function receiveMediaFrom(data, cb) {
  console.log(':rms.receiveMediaFrom: Sending request to join media room, data=%o', data)
  rmsClientSocket.emit(events.RECEIVE_MEDIA_FROM, data);
  cb()
}

function awaitExistingMediaRoomiez(cb) {
  rmsClientSocket.on(events.EXISTING_MEDIA_ROOMIEZ, (resp) => {
    console.log(':sio.awaitExistingMediaRoomiez: Received response=%o', resp)
    cb(resp)
    rmsClientSocket.off(events.EXISTING_MEDIA_ROOMIEZ)
  })
}

function awaitNewMediaRoomyArrived(cb) {
  rmsClientSocket.on(events.NEW_MEDIA_ROOMY_ARRIVED, (resp) => {
    console.log(':sio.awaitNewMediaRoomyArrived: Received response=%o', resp)
    cb(resp)
    rmsClientSocket.off(events.NEW_MEDIA_ROOMY_ARRIVED)
  })
}

function awaitReceiveMediaAnswer(cb) {
  rmsClientSocket.on(events.RECEIVE_MEDIA_ANSWER, (resp) => {
    console.log(':sio.awaitReceiveMediaAnswer: Received response=%o', resp)
    cb(resp)
    rmsClientSocket.off(events.RECEIVE_MEDIA_ANSWER)
  })
}

export {
  rmsClientSocket,
  askToConnect,
  joinMediaRoom,
  receiveMediaFrom,
  awaitExistingMediaRoomiez,
  awaitNewMediaRoomyArrived,
  awaitReceiveMediaAnswer,
};