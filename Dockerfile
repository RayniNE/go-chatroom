FROM golang:1.25.1 AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o ./main cmd/chatroom/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /build/main ./main
RUN apk --no-cache add ca-certificates
ENV ENVIRONMENT=prod
EXPOSE 8080
CMD ["./main"]