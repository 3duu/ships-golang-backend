# ❤️ GeoMatch Backend - Go Dating App API

This is the backend for **GeoMatch**, a modern location-based dating app inspired by Happn and Tinder. Built in **Go**, it supports real-time chat, location-based discovery, photo uploads, and a swipe-to-match system — all backed by **MongoDB** with geospatial indexing.

---

## 🚀 Features

- 🔐 JWT Authentication with email verification
- 👤 Profile creation & editing
- 📍 Nearby discovery using MongoDB geospatial queries
- 💘 Swipe system (like, superlike, dislike)
- 🤝 Match system with WebSocket notifications
- ✉️ Real-time chat with WebSocket
- 🖊️ Typing indicators in chat
- 🧭 Crossing paths engine (Happn-style)
- 📸 Profile photo upload, ordering & deletion
- 👀 "You got liked" queue (Tinder Gold-style)

---

## 🧱 Tech Stack

- **Go** (Golang)
- **MongoDB** (with geospatial indexing)
- **WebSockets** (via Gorilla WebSocket)
- **JWT** authentication
- **Docker** for containerized development (optional)

---

## 📦 Folder Structure

📁 internal/ ├── handlers/ # HTTP & WebSocket handlers ├── models/ # All data models (User, Swipe, Match, etc.) ├── ws/ # WebSocket manager ├── middlewares/ # Auth middleware ├── database/ # Mongo connection + index setup


---

## ⚙️ Getting Started

### 1. Clone the Repository

bash
git clone https://github.com/your-username/geomatch-backend.git
cd geomatch-backend

2. Create .env File
   PORT=8080
   MONGO_URI=mongodb://localhost:27017
   JWT_SECRET=your-secret
   
3. Start MongoDB with Docker
   docker-compose up -d
   Ensure your docker-compose.yml has MongoDB with ports and volume configured

4. Run the Server
   go run main.go

🛠 API Overview
Auth
POST /api/register

POST /api/login

GET /api/verify-email?token=...

Profile
GET /api/me

PUT /api/profile

POST /api/upload-photo

GET /api/photo/{userId}

DELETE /api/photo/{photoId}

PUT /api/photo-order

Discovery
GET /api/nearby-users

GET /api/queue

POST /api/swipe/{userId}

GET /api/got-liked

Match & Chat
GET /api/matches

GET /api/messages/{matchId}

POST /api/messages/{matchId}

Real-time (WebSocket)
/ws – Match & typing notifications

/ws/chat – Real-time chat messages

Location
POST /api/ping-location

GET /api/crossed-paths?since=24h&limit=10

🔧 Dev Tips
All requests require a valid Authorization: Bearer <token> header after login

MongoDB geospatial queries require a 2dsphere index on location

WebSocket connections must be authenticated via JWT middleware


