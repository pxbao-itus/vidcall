# Group Video Call Implementation with Pion WebRTC

This project implements a group video call system using Pion WebRTC in Go with WebSocket signaling.

## Architecture Overview

### Backend Components

1. **RTCManager** (`internal/module/rtc/rtc_manager.go`)
   - Manages WebRTC peer connections for all users in all rooms
   - Handles SDP offers, answers, and ICE candidates
   - Maintains connection lifecycle (create, get, remove)
   - Structure: `map[roomID]map[userID]map[peerID]*webrtc.PeerConnection`

2. **WebSocket Handler** (`internal/module/rtc/ws_handler.go`)
   - Handles WebSocket connections for real-time signaling
   - Routes WebRTC messages between peers
   - Manages room join/leave events
   - Cleans up connections on disconnect

3. **Data Transfer Objects** (`internal/module/rtc/rtc_dto.go`)
   - `WebSocketMessage`: Base message structure with event routing
   - `RTCSDP`: Contains SDP offer/answer data
   - `RTCIceCandidate`: ICE candidate information

### Frontend (`internal/template/call.html`)

A complete browser-based video call interface with:
- Local video preview
- Dynamic peer video grid
- Audio/video mute controls
- WebSocket signaling integration
- Automatic peer connection management

## How It Works

### 1. Connection Flow

```
User A joins room → WebSocket connected → Gets local media stream
User B joins room → Server notifies User A of new peer
User A creates offer → Sends via WebSocket → User B receives
User B creates answer → Sends back to User A
ICE candidates exchanged → Direct P2P connection established
Video/audio streams flow directly between peers
```

### 2. Signaling Messages

**Offer Message** (Initiator creates connection):
```json
{
  "event": "offer",
  "to": "peer-user-id",
  "data": {
    "type": "offer",
    "sdp": "v=0\r\no=- ... [SDP data]"
  }
}
```

**Answer Message** (Receiver responds):
```json
{
  "event": "answer",
  "from": "peer-user-id",
  "data": {
    "type": "answer",
    "sdp": "v=0\r\no=- ... [SDP data]"
  }
}
```

**ICE Candidate Message** (Network path discovery):
```json
{
  "event": "ice-candidate",
  "to": "peer-user-id",
  "data": {
    "candidate": "candidate:... [ICE data]",
    "sdpMid": "0",
    "sdpMLineIndex": 0
  }
}
```

### 3. Mesh Topology

The implementation uses a **mesh topology** where:
- Each user maintains direct P2P connections to all other users
- For N users, each user has (N-1) connections
- Scales well for small groups (2-8 people)
- Media streams flow directly between peers (no server relay)

```
User A ←→ User B
  ↑  ╲    ╱  ↑
  |    ╳    |
  ↓  ╱    ╲  ↓
User C ←→ User D
```

## Key Features

### Backend Features
- ✅ Multiple concurrent rooms
- ✅ Dynamic peer connection management
- ✅ Automatic cleanup on disconnect
- ✅ ICE candidate trickling
- ✅ Connection state monitoring
- ✅ STUN server configuration

### Frontend Features
- ✅ Automatic media device access
- ✅ Grid layout for multiple participants
- ✅ Audio/video mute toggles
- ✅ Connection status display
- ✅ Graceful error handling
- ✅ Responsive design

## Configuration

### STUN Servers

Current configuration uses Google's public STUN servers:
```go
config := webrtc.Configuration{
    ICEServers: []webrtc.ICEServer{
        {
            URLs: []string{
                "stun:stun.l.google.com:19302",
                "stun:stun1.l.google.com:19302",
            },
        },
    },
}
```

For production, consider adding TURN servers for NAT traversal:
```go
{
    URLs: []string{"turn:your-turn-server.com:3478"},
    Username: "username",
    Credential: "password",
}
```

## Usage

### Starting the Server

```bash
# Build the application
go build -o vidcall

# Run the server
./vidcall
```

### Accessing a Room

1. Navigate to `http://localhost:8080/` (or your configured port)
2. Create or join a room
3. Allow camera and microphone access
4. Share the room URL with others
5. Video call starts automatically when peers join

### Browser Requirements

- Chrome 56+, Firefox 52+, Safari 11+, Edge 79+
- HTTPS required for production (camera/mic access)
- WebRTC and WebSocket support

## API Endpoints

### WebSocket Connection
```
GET /rooms/{roomID}/ws
```
Upgrades to WebSocket for signaling

## Scalability Considerations

### Current Implementation (Mesh)
- **Best for:** 2-8 participants
- **Pros:** Low latency, no server bandwidth
- **Cons:** Client bandwidth scales with O(N)

### For Larger Groups
Consider implementing:
1. **SFU (Selective Forwarding Unit)**: Server forwards streams
2. **MCU (Multipoint Control Unit)**: Server mixes streams
3. **Simulcast**: Multiple quality streams per user

## Troubleshooting

### No video/audio
- Check browser permissions
- Verify HTTPS in production
- Check firewall/NAT settings
- Test STUN/TURN server connectivity

### Connection fails
- Verify WebSocket connection
- Check ICE candidate generation
- Review browser console for errors
- Test with symmetric NAT

### Performance issues
- Reduce video quality in getUserMedia
- Limit number of participants
- Consider implementing simulcast
- Use TURN server for relay

## Dependencies

- `github.com/pion/webrtc/v4` - WebRTC implementation
- `github.com/gorilla/websocket` - WebSocket support
- `go.uber.org/zap` - Logging
- `go.uber.org/fx` - Dependency injection

## Testing

To test locally with multiple users:
1. Open multiple browser windows/tabs
2. Use different browsers (Chrome, Firefox)
3. Use incognito/private windows
4. Use different devices on same network

## Security Considerations

- Implement authentication before WebSocket upgrade
- Validate room permissions
- Rate limit connection attempts
- Use secure WebSocket (WSS) in production
- Implement room passwords or tokens
- Monitor resource usage per user/room

## Future Enhancements

- [ ] Screen sharing support
- [ ] Recording functionality
- [ ] Chat messaging
- [ ] Virtual backgrounds
- [ ] Bandwidth adaptation
- [ ] Dominant speaker detection
- [ ] Layout customization
- [ ] Mobile app support

## License

See project LICENSE file.

