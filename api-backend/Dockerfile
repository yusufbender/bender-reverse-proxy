FROM golang:1.21 AS builder
WORKDIR /app
COPY backend.go .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o server backend.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 5678
CMD ["./server"]