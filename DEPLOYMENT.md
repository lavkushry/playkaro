# 🚀 Deployment Guide

This guide covers deploying PlayKaro to production using popular platforms.

---

## Option 1: Railway (Recommended)

**Railway** is perfect for full-stack apps with databases. Free tier includes PostgreSQL!

### Backend Deployment

1. **Install Railway CLI**
```bash
npm i -g @railway/cli
```

2. **Login to Railway**
```bash
railway login
```

3. **Initialize Project**
```bash
cd backend
railway init
```

4. **Add PostgreSQL**
```bash
railway add
# Select: PostgreSQL
```

5. **Set Environment Variables**
```bash
railway variables set PORT=8080
railway variables set JWT_SECRET=$(openssl rand -hex 32)
```

Railway will automatically detect `DATABASE_URL` from PostgreSQL.

6. **Deploy**
```bash
railway up
```

Your backend will be live at: `https://your-app.railway.app`

### Frontend Deployment (Vercel)

1. **Install Vercel CLI**
```bash
npm i -g vercel
```

2. **Deploy Frontend**
```bash
cd frontend
vercel --prod
```

3. **Update API URL**
In `frontend/src/store/*.js`, update:
```javascript
const API_URL = 'https://your-backend.railway.app/api/v1';
```

---

## Option 2: Docker + AWS EC2

### Prerequisites
- AWS account
- EC2 instance (t2.micro for testing)
- Docker installed on EC2

### Steps

1. **SSH into EC2**
```bash
ssh -i your-key.pem ubuntu@your-ec2-ip
```

2. **Install Docker**
```bash
sudo apt update
sudo apt install docker.io docker-compose -y
sudo usermod -aG docker $USER
```

3. **Clone Repository**
```bash
git clone https://github.com/lavkushry/playkaro.git
cd playkaro
```

4. **Set Environment Variables**
```bash
cd backend
cp .env.example .env
nano .env  # Edit with production values
```

5. **Start Services**
```bash
cd docker
docker-compose up -d

cd ../backend
docker build -t playkaro-backend .
# Run DragonflyDB (Redis compatible)
docker run -d --name dragonfly -p 6379:6379 docker.dragonflydb.io/dragonflydb/dragonfly

# Run Backend linked to Dragonfly
docker run -d -p 8080:8080 --env-file .env --link dragonfly:dragonfly playkaro-backend
```

6. **Configure Nginx (Optional)**
```bash
sudo apt install nginx
```

Create `/etc/nginx/sites-available/playkaro`:
```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:5173;
    }

    location /api {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
    }
}
```

```bash
sudo ln -s /etc/nginx/sites-available/playkaro /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

---

## Option 3: Render

### Backend

1. Create new **Web Service** on Render
2. Connect GitHub repo
3. Set:
   - **Build Command**: `cd backend && go build -o main`
   - **Start Command**: `cd backend && ./main`
4. Add environment variables
5. Add PostgreSQL database (Render provides free tier)
6. **Redis/DragonflyDB**: Render offers Redis. Use that connection string.

### Frontend

1. Create new **Static Site**
2. Set:
   - **Build Command**: `cd frontend && npm install && npm run build`
   - **Publish Directory**: `frontend/dist`

---

## Environment Variables Checklist

### Backend
- ✅ `PORT` (8080)
- ✅ `DB_HOST`
- ✅ `DB_PORT` (5432)
- ✅ `DB_USER`
- ✅ `DB_PASSWORD`
- ✅ `DB_NAME`
- ✅ `JWT_SECRET` (use `openssl rand -hex 32`)

### Frontend
- ✅ `VITE_API_URL` (your backend URL)

---

## Post-Deployment

### 1. Test Endpoints
```bash
curl https://your-backend.railway.app/health
```

Expected: `{"status":"ok"}`

### 2. Create Admin User
Connect to production database:
```sql
UPDATE users SET is_admin = true WHERE email = 'admin@playkaro.com';
```

### 3. Monitor Logs
```bash
# Railway
railway logs

# Docker
docker logs <container-id>
```

### 4. Set Up SSL (if using EC2)
```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

---

## Troubleshooting

### Backend won't start
- Check `DATABASE_URL` is set correctly
- Verify PostgreSQL is accessible
- Check logs: `railway logs` or `docker logs`

### CORS errors
- Add frontend URL to backend CORS config
- Update `main.go`:
```go
r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"https://your-frontend.vercel.app"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
    AllowCredentials: true,
}))
```

### WebSocket not connecting
- Ensure WebSocket endpoint is accessible
- Check firewall rules (EC2 Security Groups)
- Use `wss://` for HTTPS sites

---

## Cost Estimates

| Platform | Backend | Database | Frontend | Total/month |
|----------|---------|----------|----------|-------------|
| Railway  | $5      | Free     | -        | $5          |
| Vercel   | -       | -        | Free     | $0          |
| AWS EC2  | $5      | $10      | $5       | $20         |
| Render   | Free    | Free     | Free     | $0          |

**Recommended**: Railway (Backend + DB) + Vercel (Frontend) = **$5/month**

---

## Production Checklist

- [ ] Environment variables set
- [ ] Database migrations run
- [ ] Admin user created
- [ ] SSL certificate installed
- [ ] CORS configured
- [ ] Logs monitoring set up
- [ ] Backups enabled
- [ ] Health check endpoint tested
- [ ] WebSocket connection verified
- [ ] Payment gateway configured (if applicable)

---

**Need help?** Open an issue on [GitHub](https://github.com/lavkushry/playkaro/issues)
