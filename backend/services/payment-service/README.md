# Payment Service

Production-grade payment processing microservice for PlayKaro.

## Features

- **Multi-Gateway Support**: Razorpay, Cashfree
- **Fraud Detection**: Velocity checks, amount anomaly detection
- **Webhook Processing**: Secure signature verification
- **Event Publishing**: Kafka event streaming
- **High Availability**: Production-ready with Docker/Kubernetes

## Quick Start

### Local Development

```bash
# 1. Setup database
createdb payments_db
psql payments_db < migrations/001_init.sql

# 2. Configure environment
cp .env.example .env
# Edit .env with your Razorpay credentials

# 3. Run service
go run cmd/main.go
```

### Docker

```bash
docker build -t playkaro/payment-service:latest .
docker run -p 8081:8081 --env-file .env playkaro/payment-service:latest
```

## API Endpoints

### Initiate Deposit
```http
POST /v1/payments/deposit
Authorization: Bearer <JWT>

{
  "amount": 1000,
  "currency": "INR",
  "gateway": "razorpay"
}
```

### Webhook Handler
```http
POST /v1/payments/webhook/razorpay
X-Razorpay-Signature: <signature>

<Razorpay webhook payload>
```

### Get Order Status
```http
GET /v1/payments/order/{order_id}
Authorization: Bearer <JWT>
```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `PAYMENT_DB_HOST` | PostgreSQL host | Yes |
| `RAZORPAY_KEY_ID` | Razorpay API key | Yes |
| `RAZORPAY_KEY_SECRET` | Razorpay secret | Yes |
| `PORT` | Service port | No (default: 8081) |

## Architecture

```
Client → API Gateway → Payment Service → Razorpay/Cashfree
                             ↓
                        PostgreSQL
                             ↓
                          Kafka (Events)
```

## Fraud Detection

- **Velocity Checks**: Max 5 deposits/hour
- **Daily Limits**: ₹50,000/day
- **Anomaly Detection**: Flags unusual amounts

## Deployment

Production deployment uses Kubernetes:

```bash
kubectl apply -f k8s/payment-service.yaml
```

## License

Proprietary - PlayKaro Platform
