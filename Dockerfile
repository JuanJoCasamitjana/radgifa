# Build stage for frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# Build stage for backend
FROM golang:1.25.5-alpine AS backend-builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@v1.8.12 \
	&& swag init -g cmd/api/main.go -o docs

COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Final stage
FROM alpine:latest

WORKDIR /root/

COPY --from=backend-builder /app/main .
COPY --from=backend-builder /app/frontend/dist ./frontend/dist

EXPOSE 8080

CMD ["./main"]
