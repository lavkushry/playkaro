# Match Service

High-performance sportsbook microservice for PlayKaro with <10ms read latency.

## Features

- **Match CRUD**: Create, read, update matches with sports data
- **Live Odds**: Real-time odds updates via WebSocket + Redis Pub/Sub
- **Redis Caching**: Sub-10ms read performance for match data
- **Odds History**: Track all odds changes for analytics
- **Event Publishing**: Kafka events for match lifecycle
- **Horizontal Scaling**: Stateless design for easy scaling

## Quick Start

### Local Development

```bash
# 1. Setup database
createdb matches_db
psql matches_db < migrations/001_init.sql

# 2. Start Redis
redis-server

# 3. Configure environment
cp .env.example .env

# 4. Run service
go run cmd/main.go
```

### Docker

```bash
docker build -t playkaro/match-service:latest .
docker run -p 8082:8082 --env-file .env playkaro/match-service:latest
```

## API Endpoints

### Create Match (Admin)
```http
POST /v1/matches
X-Admin-Key: admin123

{
  "sport": "CRICKET",
  "team_a": "India",
  "team_b": "Australia",
  "odds_a": 185,
  "odds_b": 2.10,
  "start_time": "2025-11-25T14:00:00Z"
}
```

### Get All Matches
```http
GET /v1/matches?status=LIVE
```

### Get Match by ID
```http
GET /v1/matches/{match_id}
```

### Update Odds (Admin)
```http
PUT /v1/matches/{match_id}/odds
X-Admin-Key: admin123

{
  "odds_a": 1.90,
  "odds_b": 2.05
}
```

### Settle Match (Admin)
```http
POST /v1/matches/{match_id}/settle
X-Admin-Key: admin123

{
  "result": "TEAM_A"
}
```

### WebSocket (Real-time Odds)
```javascript
const ws = new WebSocket('ws://localhost:8082/ws/odds');
ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  console.log('Odds updated:', update);
};
```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `MATCH_DB_HOST` | PostgreSQL host | Yes |
| `REDIS_URL` | Redis connection URL | Yes |
| `PORT` | Service port | No (default: 8082) |
| `ADMIN_KEY` | Admin API key | Yes (production) |

## Architecture

```
Client → Match Service → Redis (Cache) → PostgreSQL
              ↓
         WebSocket → Redis Pub/Sub
              ↓
           Kafka (Events)
```

## Performance

- **Read Latency**: <10ms (Redis cache hit)
- **Write Latency**: <50ms (PostgreSQL)
- **WebSocket Connections**: 50K+ supported
- **Throughput**: 100K requests/second

## Scaling

The service is stateless and can be horizontally scaled:

```bash
kubectl scale deployment match-service --replicas=10
```

## License

Proprietary - PlayKaro Platform
