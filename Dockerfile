FROM node:20-alpine AS frontend
WORKDIR /app/console
COPY console/package.json console/package-lock.json* ./
RUN npm ci
COPY console/ .
RUN npm run build

FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/console/dist ./console/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o nexus ./cmd/server
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o migrate ./cmd/migrate

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/nexus .
COPY --from=builder /app/migrate .
COPY --from=builder /app/database/migrations ./database/migrations
COPY --from=builder /app/configs/config.example.yaml ./configs/config.yaml
COPY --from=builder /app/console/dist ./console/dist
EXPOSE 8080 9090
CMD ["./nexus"]
