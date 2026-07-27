# File Uploads Go + React POC

Minimal, reusable file-upload stack for Go + React.

- **`backend/`** — Go library (`pkg/upload`) + chi demo server (local disk)
- **`frontend/`** — React 19 + Vite + TypeScript POC with progress UI
- **`frontend/upload-lib/`** — Detached TS/JS clients (no React) for reuse

## Quick start

```bash
# Terminal 1 — backend
cd backend
go run ./cmd/server

# Terminal 2 — frontend
cd frontend
npm install
npm run dev
```

Open http://localhost:5173. Vite proxies `/api` → `http://localhost:8080`.

## API

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/upload/stream` | Streaming multipart upload (single or multi file) |
| POST | `/api/upload/init` | Start chunked session |
| POST | `/api/upload/chunk?upload_id=&chunk=` | Upload one chunk |
| POST | `/api/upload/complete?upload_id=` | Assemble chunks |
| GET | `/api/upload/status?upload_id=` | Resume / missing chunks |
| GET | `/api/upload/progress?upload_id=` | SSE progress |
| GET | `/health` | Health check |

Env: `PORT` (default `8080`), `UPLOAD_DIR` (default `./uploads`), `CORS_ORIGIN` (default `http://localhost:5173`), `MAX_UPLOAD_SIZE`.

## Frontend modes

1. **Streamed** — single file multipart + XHR progress  
2. **Streamed multi** — multiple parts in one request  
3. **Chunked** — resumable 5MB chunks  
4. **Chunked multi** — parallel chunked sessions per file  

Copy-paste clients: `frontend/upload-lib/ts` and `frontend/upload-lib/js`.

## Tests

```bash
cd backend && go test ./...
cd frontend && npm test
```

Covered: chunk size math (`TotalChunks` / `ExpectedChunkSize`) and `SanitizeFilename` on both sides.

## Optional S3

S3 multipart upload is a **separate module** so the default backend has **no AWS dependencies**.

See [`backend/S3.md`](backend/S3.md) and [`backend/optional/s3storage`](backend/optional/s3storage).

To remove S3 entirely, delete `backend/optional/s3storage`.

## Layout

```
backend/
  cmd/server/          # chi app
  pkg/upload/          # reusable library
  optional/s3storage/  # optional separate module
frontend/
  src/                 # React POC UI
  upload-lib/          # detached TS + JS upload clients
```
