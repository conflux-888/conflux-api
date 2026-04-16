# ---- Stage 1: build admin UI ----
FROM node:22-alpine AS admin-builder
WORKDIR /app/admin
RUN corepack enable
COPY admin/package.json admin/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY admin/ ./
RUN pnpm run build

# ---- Stage 2: build Go binary with embedded admin dist ----
FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Replace the placeholder dist with the freshly built frontend
RUN rm -rf ./admin/dist
COPY --from=admin-builder /app/admin/dist ./admin/dist

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server .

# ---- Stage 3: runtime ----
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
