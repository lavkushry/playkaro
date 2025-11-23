# 🎰 PlayKaro - Real Money Gaming Platform

A full-stack **Real-Money Gaming (RMG)** platform built with **Go**, **React**, and **PostgreSQL**. Features live sports betting, wallet management, WebSocket-powered real-time odds, and an admin dashboard.

![Status](https://img.shields.io/badge/Status-Production%20Ready-success)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-19.0+-61DAFB?logo=react)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?logo=postgresql)

---

## 🚀 Features

### 🔐 Authentication
- JWT-based authentication
- Bcrypt password hashing
- Role-based access control (User/Admin)

### 💰 Wallet System
- Real-time balance tracking
- Deposit & Withdraw (mock payment gateway)
- Transaction history with audit trail
- Optimistic UI updates

### 🏏 Live Betting
- Real-time odds via WebSockets
- Bet slip with potential win calculator
- Support for multiple sports/matches
- Automatic bet settlement

### 👨‍💼 Admin Dashboard
- Create and manage matches
- Update odds dynamically
- One-click bet settlement
- Automated payout to winners

### 📊 History & Analytics
- Transaction history (Deposits/Withdrawals/Bets/Wins)
- Bet history with Won/Lost/Pending status
- Detailed audit logs

---

## 🛠️ Tech Stack

### Backend
- **Go** (Gin framework)
- **PostgreSQL** (ACID-compliant database)
- **DragonflyDB** (High-performance Redis alternative)
- **GraphQL** (gqlgen) + **REST API**
- **WebSockets** (Redis Pub/Sub)
- **JWT** (golang-jwt/jwt)

### Frontend
- **React 19** + **Vite**
- **Tailwind CSS** (Midnight Gold theme)
- **Zustand** (state management)
- **Axios** (HTTP client)

### Infrastructure
- **Docker** (PostgreSQL + DragonflyDB containers)
- **Git** (version control)

---

## 📦 Installation

### Prerequisites
- **Go 1.21+** ([Download](https://go.dev/dl/))
- **Node.js 18+** ([Download](https://nodejs.org/))
- **Docker Desktop** ([Download](https://www.docker.com/products/docker-desktop/))

### Quick Start

1. **Clone the repository**
```bash
git clone https://github.com/lavkushry/playkaro.git
cd playkaro
```

2. **Start the database**
```bash
cd docker
docker-compose up -d
```

3. **Run the backend**
```bash
cd ../backend
go run main.go
```
Backend runs on `http://localhost:8080`

4. **Run the frontend** (in a new terminal)
```bash
cd ../frontend
npm install
npm run dev
```
Frontend runs on `http://localhost:5173`

---

## 🎮 Usage

### Creating Your First Account
1. Go to `http://localhost:5173/register`
2. Enter username, email, mobile, and password
3. Login automatically redirects to Dashboard

### Placing Your First Bet
1. Click **"Sportsbook"** from Dashboard
2. See live matches (e.g., "India vs Australia")
3. Click on odds (e.g., India @ 1.80)
4. Enter bet amount (e.g., ₹1000)
5. Review potential win (e.g., ₹1800)
6. Click **"Place Bet"**

### Accessing Admin Dashboard
**Note:** Admin access requires manual database update
```sql
UPDATE users SET is_admin = true WHERE email = 'your@email.com';
```

Then navigate to `/admin` to:
- Create new matches
- Update odds in real-time
- Settle matches and trigger payouts

---

## 📚 API Documentation

### Authentication
```http
POST /api/v1/auth/register
POST /api/v1/auth/login
```

### Wallet (Protected)
```http
GET  /api/v1/wallet/
POST /api/v1/wallet/deposit
POST /api/v1/wallet/withdraw
```

### Betting (Protected)
```http
GET  /api/v1/matches
POST /api/v1/bet/
```

### Admin (Admin Only)
```http
POST /api/v1/admin/matches
PUT  /api/v1/admin/matches/:id/odds
POST /api/v1/admin/matches/:id/settle
```

### History (Protected)
```http
GET /api/v1/transactions
GET /api/v1/bets
```

### WebSocket
```
ws://localhost:8080/ws
```
Broadcasts odds updates via **Redis Pub/Sub** (DragonflyDB)

### GraphQL
```
POST http://localhost:8080/query
```
Playground available at `/playground`

---

## 🗄️ Database Schema

```sql
-- Users table with admin flag
CREATE TABLE users (
  id UUID PRIMARY KEY,
  username VARCHAR(50) UNIQUE,
  email VARCHAR(100) UNIQUE,
  password_hash TEXT,
  mobile VARCHAR(20),
  is_admin BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP
);

-- Wallets for balance tracking
CREATE TABLE wallets (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  balance DECIMAL(15, 2),
  currency VARCHAR(3) DEFAULT 'INR'
);

-- Matches for betting
CREATE TABLE matches (
  id UUID PRIMARY KEY,
  team_a VARCHAR(100),
  team_b VARCHAR(100),
  odds_a DECIMAL(5, 2),
  odds_b DECIMAL(5, 2),
  status VARCHAR(20) DEFAULT 'LIVE'
);

-- Bets placed by users
CREATE TABLE bets (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  match_id UUID REFERENCES matches(id),
  selection VARCHAR(10),
  amount DECIMAL(15, 2),
  odds DECIMAL(5, 2),
  potential_win DECIMAL(15, 2),
  status VARCHAR(20) DEFAULT 'PENDING'
);
```

---

## 🎨 Design System

**Color Palette (Midnight Gold)**
- Primary: `#0F172A` (Slate 900)
- Accent Gold: `#F59E0B` (Amber 500)
- Accent Blue: `#3B82F6` (Blue 500)
- Success: `#10B981` (Emerald 500)
- Error: `#EF4444` (Red 500)

**Typography**
- Font: Inter
- Dark background with high-contrast text

---

## 🧪 Testing

### Manual Testing Checklist
- [ ] Register new user
- [ ] Login and see Dashboard
- [ ] Deposit ₹10,000
- [ ] Place bet on India @ 1.80 for ₹1,000
- [ ] Check Transaction History
- [ ] Admin: Create new match
- [ ] Admin: Settle match (India wins)
- [ ] Verify wallet credited with ₹1,800

### WebSocket Testing
Open two browser tabs to `/sportsbook` and watch odds update simultaneously every 5 seconds.

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📝 License

This project is open source and available under the MIT License.

---

## 👨‍💻 Developer

**Lavkush Kumar**
- GitHub: [@lavkushry](https://github.com/lavkushry)
- Email: lavkushry@gmail.com

---

## 🙏 Acknowledgments

- Inspired by [PlayKaro365](https://playkaro365.com)
- Built following Google/Apple-level engineering standards
- PRD and Architecture documentation available in `/Memory` directory

---

**⭐ Star this repo if you find it useful!**
