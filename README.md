## Movie App (Go + React)

A full-stack movie explorer with Go/Gin backend, MongoDB, and a React (Vite) client styled in a Netflix-like dark theme. Features JWT auth, protected admin reviews, AI-powered recommendations, and YouTube trailer streaming.

### Stack
- Backend: Go (Gin), MongoDB, JWT auth, Swagger docs
- Frontend: React + Vite, React Router, Bootstrap, custom Netflix-like styling
- Docker: Services for API, MongoDB, seed importer, and client

### Prerequisites
- Docker and Docker Compose
- Optional: Go 1.21+, Node 20+, Yarn (if running locally without Docker)

### Environment
Backend env file: `Backend/movie-app-go/.env` (already provided) defines Mongo creds, JWT secrets, and OpenRouter keys.
Client env file: `Client/movie-app-react/.env` with `VITE_API_URL=http://localhost:5000/api/v1` for local/dev.

### Run with Docker (recommended)
```sh
docker-compose --env-file Backend/movie-app-go/.env up --build
```
Services:
- `db`: MongoDB with data seeded from `./seed/*.json`
- `mongo-seed`: Imports genres, rankings, movies, users (admin user: admin@movieapp.local / AdminPass123!)
- `movie-api`: Go backend on http://localhost:5000
- `client`: React build served via nginx on http://localhost:5173

### Local development (frontend only)
```sh
cd Client/movie-app-react
yarn install
yarn dev --host
```
Ensure the backend is running at `VITE_API_URL` (default http://localhost:5000/api/v1).

### Local development (backend)
```sh
cd Backend/movie-app-go
go run main.go
```
Swagger docs available at http://localhost:5000/swagger/index.html.

### Seeding data
Seed JSON files live in `/seed` (genres, rankings, movies, users). With Docker, seeding happens on `docker-compose up`. To re-seed manually:
```sh
docker-compose --env-file Backend/movie-app-go/.env run --rm mongo-seed
```

### Key routes (API)
- `POST /api/v1/register`, `POST /api/v1/login`, `POST /api/v1/logout`
- `GET /api/v1/movies`, `GET /api/v1/movie/:imdbId`
- `GET /api/v1/genres`, `GET /api/v1/recommendatedmovies`, `GET /api/v1/recommendations-ai`
- Admin/protected: `PATCH /api/v1/movie/review/:imdbId`, user CRUD

### Frontend highlights
- Hero banner with featured movie, dark glassy navbar, responsive cards, hover play overlay
- Trailer playback via ReactPlayer on the stream route using the movie’s `youtube_id`
- Protected routes gate admin review/editor

### Credentials
- Admin: `admin@movieapp.local` / `AdminPass123!`

### Notes
- If the client in Docker cannot reach the API, ensure `VITE_API_URL` is host-accessible (http://localhost:5000/api/v1) for browser traffic.
- AI recommendations may take time; the UI falls back to cached recommendations first, then AI with a timeout.