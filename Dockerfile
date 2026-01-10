FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o social-network cmd/app/main.go

FROM alpine:3.18

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/social-network .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/certs ./certs

EXPOSE 8080

CMD ["./social-network"]
