# URL Shortener Service

**[View Vietnamese Version](README.vi.md)**

A URL shortening service built with Golang and React that allows users to create short, memorable links with click tracking and expiration management.
Demo https://shorty-black.vercel.app/home
---

## 📋 Table of Contents

- [Problem Description](#-problem-description)
- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [Getting Started](#-getting-started)
- [Design & Technical Decisions](#-design--technical-decisions)
- [Security Considerations](#-security-considerations)
- [Scalability](#-scalability)
- [Trade-offs](#-trade-offs)
- [Challenges & Learnings](#-challenges--learnings)

---

## 🎯 Problem Description

Building a URL shortening service that allows:
- Users input long URL → receive an easy-to-remember short URL
- Access short URL → redirect to original URL
- Track click counts
- Manage created links with expiration dates

**Summary**: Create short links, proper redirection, click tracking, prevent duplicates, and manage links per user.

---

## ✨ Features

### Core Features
- Create short URLs from long URLs
- Automatic redirection to original URLs
- Click tracking and analytics
- List all created links
- Link expiration management
- Duplicate URL prevention

### Additional Features
- User authentication with JWT
- QR code generation (via Cloudinary)
- Rate limiting (100 URLs per user per day)
- Comprehensive URL validation

---

## 🛠 Tech Stack

### Backend
- **Language**: Golang
- **Framework**: Gin
- **Database**: PostgreSQL (hosted on Neon)
- **Authentication**: JWT
- **File Storage**: Cloudinary
- **Deployment**: Fly.io

### Frontend
- **Framework**: React

---

## 🚀 Getting Started
```bash
# Clone
git clone https://github.com/nhatcn/shorty.git
cd shorty
```
### Backend Setup

```bash
# Navigate to backend folder
cd backend

# Install dependencies
go mod download

# Create .env file
New-Item -Path . -Name ".env" -ItemType "File"
```

**Configure `.env`:**
```env
DATABASE_URL=postgresql://neondb_owner:npg_xKsv3fSC5myF@ep-proud-shadow-a1gzrzfn-pooler.ap-southeast-1.aws.neon.tech/neondb?sslmode=require&channel_binding=require  (for convinient)
JWT_SECRET=your_secret_key
PORT=8080
FRONTEND_URL=http://localhost:3000
CLOUDINARY_CLOUD_NAME=your_cloudinary_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
```

**Run migrations:**


**Start backend server:**
```bash
go run cmd/server/main.go
```

Backend runs at: `http://localhost:8080`

### Frontend Setup

```bash
# Navigate to frontend folder
cd frontend

# Install dependencies
npm install

# Create .env file
New-Item -Path . -Name ".env" -ItemType "File"
```

**Configure `.env`:**
```env
BE_URL=http://localhost:8080
```

**Start development server:**
```bash
npm start
```

Frontend runs at: `http://localhost:3000`


## 🧠 Design & Technical Decisions

### Why PostgreSQL?

**Reasons:**
- Supports standard SQL, powerful for complex queries (joins, aggregation, indexes)
- Transactions handle concurrency safely (e.g., creating shortCode from auto-increment ID)
- Easy deployment on Fly.io, Neon, Railway → free tier + quick setup
- Supports constraints & indexes: UNIQUE, composite index, partial index → optimize performance

### Why RESTful API?

**Reasons:**
- Popular, easy to understand, easy to test with Postman/curl
- Clear endpoints:
  - `POST /urls` → create short URL
  - `GET /urls/:shortCode` → redirect
- Easy to extend, easy to deploy for React/Vue frontend
- JSON payload → easy to parse and validate

### Short Code Generation Algorithm

**Method**: Use PostgreSQL's auto-increment ID → encode to Base62

**Why:**
- Auto-increment ID ensures 100% uniqueness
- Encode ID to Base62 to create shortCode

### Handling Conflicts/Duplicates

**Duplicate URL + User:**
- If user submits the same URL → return existing shortCode
- PostgreSQL constraints ensure no duplicates

**ShortCode Conflicts:**
- Never happens because using auto-increment ID → Base62 → 100% unique
- If using random Base62 → must check database → retry if collision

**Concurrency:**
- Database handles transactions → ensures no duplicates, avoids race conditions

---

## 🔒 Security Considerations

### Implemented Security Measures

1. **ShortCode Predictability**
   - ShortCode can reveal URL creation order
   - Can use ID obfuscation (XOR, hash) to avoid pattern leakage

2. **URL Validation**
   - Prevent localhost, internal IPs, private IPs
   - Only allow HTTP/HTTPS schemas

3. **Click Tracking**
   - Prevent injection attacks
   - Validate IDs

4. **Rate Limiting**
   - Limit to 100 URLs per day → prevent spam

5. **Expired URLs**
   - Check ExpiresAt → prevent redirecting expired URLs

6. **File Upload (QR Code)**
   - Cloudinary ensures safety, avoid storing directly on server

7. **Auth & Authorization**
   - Users can only view/delete their own URLs
   - Prevent querying other users' URLs

---

## 📈 Scalability

### Handling 100x Traffic Increase

#### Read-Heavy Traffic
- **Cache URL mapping** (shortCode → originalURL) in Redis → reduce DB load
- **Cache click counts** or use batch updates to reduce excessive writes

#### Write-Heavy Traffic
- **Auto-increment ID**: If multiple servers, DB must handle concurrency
- **Clicks**: Can use batch insert or append to log → aggregate later

#### Database Sharding/Partitioning
- When data exceeds hundreds of millions of rows
- Can partition tables by `user_id % N` or by date range

---

## ⚖️ Trade-offs

### PostgreSQL vs NoSQL

**Why PostgreSQL:**
- Safe transactions, strong constraints & indexes, easy to deploy on Fly.io/Neon

**Drawbacks:**
- Scaling write-heavy workloads or millions of links is more complex than NoSQL

**Why it fits:**
- This is a small app with moderate traffic → PostgreSQL is simple and powerful enough

### RESTful API vs GraphQL/gRPC

**Why REST:**
- Popular, easy to understand, easy to test with Postman/curl, easy frontend integration

**Drawbacks:**
- GraphQL optimizes queries better, avoids over-fetching

**Why it fits:**
- Small app, simple data → REST is fast and easy to deploy

### Auto-increment ID + Base62 for ShortCode

**Why this approach:**
- 100% unique, no retry needed, high performance

**Drawbacks:**
- ShortCode increases sequentially → reveals link creation order
- ShortCode gets longer as link count grows

**Why it fits:**
- Small app, low IDs → short codes, easy to manage

### Duplicate URL Handling with Constraints + User Check

**Why this approach:**
- Duplicate URL from same user → return existing shortCode → prevent spam, save resources
- DB constraints ensure concurrency safety

**Drawbacks:**
- If wanting random shortCode for same URL, requires more resources and duplicate checking

**Why it fits:**
- Small app → simple, safe, efficient

### Index & Performance

**Why indexes:**
- Index on `short_code`, `user_id`, composite `(user_id, expires_at)` → fast queries

**Drawbacks:**
- Many indexes → consume more storage, slightly slower insert/update

**Why it fits:**
- Small app → this trade-off is acceptable, query performance is more important

---

## 💡 Challenges & Learnings

### What Problems Did I Encounter?

1. **New to Golang**
   - Unfamiliar with syntax and flows (struct, interface, method, package)

2. **Database Connection on Fly.io**
   - Encountered errors → couldn't deploy PostgreSQL directly

### How Did I Solve Them?

1. **Learning Golang**
   - Self-studied syntax, package structure, organizing service/repository/handler
   - Understood backend flow clearly

2. **Database Deployment**
   - Switched database to Neon (PostgreSQL cloud)
   - Fast deployment, stable, able to test API
   - Used environment variables for database connection and frontend URL

### What Did I Learn?

1. **Backend Flow Understanding**
   - Clear understanding of Golang backend flow: handler → service → repository → DB

2. **PostgreSQL Skills**
   - Handling concurrency, duplicates, transactions in PostgreSQL

3. **Performance Optimization**
   - Understanding indexes, query performance, caching

4. **System Design**
   - How to design systems that can scale

5. **Cloud Deployment**
   - Experience deploying backend + database on cloud platforms (Neon/Fly.io)

---
### Limitations & Improvements:
1. **What is currently missing?**

- ORM not used yet → Currently interacting with the database directly, making the repository code verbose.

- Project structure may not be fully optimized → Still getting familiar with Golang and common patterns (service/repository/handler).

- No custom alias for short URLs.

- No caching for clicks or links → High traffic could overload the database.

- ID not obfuscated → ShortCodes increase sequentially, revealing the order of link creation.

2. **What would be done with more time?**

- Integrate OCR to read links from images:

- Users can upload images containing QR codes or URLs → automatically extract and shorten.

- Can use Google Vision API, Tesseract, or other cloud OCR services.

- Add custom alias → Users can set their own shortCode instead of only auto-generated.

- Advanced analytics: Track clicks over time, by device, or by IP address.

- Caching: Use Redis to reduce database load on frequent redirects.

- Obfuscate ID so ShortCodes do not reveal creation order.

3. **What is needed to be production-ready?**

- HTTPS and secure header configuration.

- Auto-scaling & Load Balancer → if traffic increases 100x.

- Database backup & recovery.
