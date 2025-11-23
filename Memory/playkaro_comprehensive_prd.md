# Product Requirements Document (PRD)
## Real-Money Gaming Platform - PlayKaro365 Clone

**Version:** 1.0  
**Date:** November 2025  
**Tech Stack:** Go (Backend) + React (Frontend) + PostgreSQL + Redis  
**Target Market:** India, Southeast Asia  
**Platform Type:** Real-Money Gaming (RMG) Web Application

---

## TABLE OF CONTENTS

1. Executive Summary
2. System Architecture Overview
3. User Management & Authentication
4. Wallet & Financial System
5. Game Portfolio (Complete Catalog)
6. Bonus & Promotion Engine
7. Affiliate/Agent System
8. VIP & Loyalty Program
9. Payment Gateway Integration
10. Admin Back-Office
11. Compliance & Legal Requirements
12. Technical Specifications
13. Security & Anti-Fraud
14. Performance Requirements
15. Mobile Responsiveness
16. Analytics & Reporting
17. Customer Support System
18. Notification System
19. Responsible Gaming Features
20. API Documentation Requirements

---

## 1. EXECUTIVE SUMMARY

### 1.1 Product Vision
Build a feature-complete Real-Money Gaming platform that surpasses Playkaro365 by offering:
- 3000+ games (Crash, Slots, Live Casino, Card Games, Sports Prediction)
- Seamless UPI/Bank transfer integration
- Multi-tier affiliate system with auto-commission distribution
- Advanced risk management & admin controls
- Sub-100ms game response time for 10,000 concurrent users

### 1.2 Core Value Propositions
- **For Players:** Instant deposits, fair RNG, 24/7 withdrawals, engaging gameplay
- **For Affiliates:** 3-tier commission structure, real-time earnings dashboard
- **For Admins:** Live risk exposure monitoring, fraud detection, manual override controls

### 1.3 Success Metrics (KPIs)
- Daily Active Users (DAU): 50,000+ within 6 months
- Average Revenue Per User (ARPU): ₹500/month
- Deposit Success Rate: >95%
- Withdrawal Processing Time: <2 hours (automated), <24 hours (manual review)
- System Uptime: 99.9%
- Cashout Latency: <150ms (from user click to server confirmation)

---

## 2. SYSTEM ARCHITECTURE OVERVIEW

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────┐
│                   CDN (Cloudflare)                  │
│            Static Assets, DDoS Protection            │
└────────────────────┬────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────┐
│              Load Balancer (Nginx)                  │
│         SSL Termination, Rate Limiting              │
└────────────────────┬────────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
┌───────▼────────┐      ┌────────▼────────┐
│  React SPA     │      │   Go Backend    │
│  (Frontend)    │◄─────┤   (Gin/Echo)    │
│                │ HTTP  │   API Servers   │
└────────────────┘ REST  └────────┬────────┘
                                  │
                    ┌─────────────┼─────────────┐
                    │             │             │
         ┌──────────▼──┐   ┌─────▼──────┐  ┌──▼────────┐
         │ PostgreSQL  │   │   Redis    │  │ WebSocket │
         │  (Primary)  │   │   Cache    │  │  Server   │
         │             │   │ Pub/Sub    │  │  (Games)  │
         └─────────────┘   └────────────┘  └───────────┘
                    │
         ┌──────────▼──────────┐
         │  S3 Compatible      │
         │  (Media Storage)    │
         └─────────────────────┘
```

### 2.2 Technology Stack Details

**Backend (Go)**
- Framework: Gin (HTTP routing) or Echo (if you need middleware-heavy setup)
- ORM: GORM or sqlx (prefer sqlx for performance-critical sections)
- WebSocket: gorilla/webs