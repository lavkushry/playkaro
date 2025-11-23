# PlayKaro - Real Money Gaming Platform

## Prerequisites
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose

## Getting Started

### 1. Start Infrastructure
```bash
cd docker
docker-compose up -d
```

### 2. Run Backend
```bash
cd backend
go run main.go
```
Server runs on `http://localhost:8080`

### 3. Run Frontend
```bash
cd frontend
npm install
npm run dev
```
App runs on `http://localhost:5173`

## Project Structure
- `/backend`: Go + Gin + Postgres
- `/frontend`: React + Vite + Tailwind
- `/docker`: Database infrastructure
