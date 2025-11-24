# Game Engine Service

Plugin-based game engine for PlayKaro, enabling rapid addition of new games like Winzo.

## Features

- **Plugin Architecture**: Easily add new games (Ludo, Carrom, etc.)
- **Session Management**: Multi-player game sessions
- **Real-time Updates**: WebSocket game state streaming
- **Common Logic**: Unified interface for all games

## Quick Start

### Local Development

```bash
# 1. Configure environment
cp .env.example .env

# 2. Run service
go run cmd/main.go
```

### Docker

```bash
docker build -t playkaro/game-engine:latest .
docker run -p 8083:8083 playkaro/game-engine:latest
```

## API Endpoints

### List Games
```http
GET /v1/games
```

### Create Session
```http
POST /v1/games/sessions
X-User-ID: user1

{
  "game_id": "ludo_classic"
}
```

### Join Session
```http
POST /v1/sessions/{session_id}/join
X-User-ID: user2
```

### Make Move
```http
POST /v1/sessions/{session_id}/move
X-User-ID: user1

{
  "type": "ROLL_DICE"
}
```

### WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8083/ws/sessions/{session_id}');
ws.onmessage = (event) => {
  console.log('Game state:', JSON.parse(event.data));
};
```

## Adding a New Game

1. Create a new package in `games/`
2. Implement the `IGame` interface
3. Register the game in `cmd/main.go`

```go
reg.RegisterGame(mygame.NewMyGame())
```

## License

Proprietary - PlayKaro Platform
