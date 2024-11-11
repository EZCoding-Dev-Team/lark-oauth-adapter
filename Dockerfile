FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go

FROM scratch as scratch

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 3000
ENTRYPOINT ["./main"]
