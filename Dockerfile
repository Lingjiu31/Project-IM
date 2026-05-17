FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o im-server ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/im-server .
EXPOSE 8080
CMD ["./im-server"]
