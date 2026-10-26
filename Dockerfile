# Stage 1: Build Go application
FROM golang:1.27.1-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

# Stage 2: Production image
FROM alpine:3.23

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /server ./server
COPY --from=builder /src/config ./config
COPY --from=builder /src/web ./web

RUN mkdir -p /app/uploads/articles /app/uploads/users

ENV GIN_MODE=release

EXPOSE 5001

CMD ["./server"]