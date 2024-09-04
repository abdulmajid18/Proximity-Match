FROM golang:1.23.0-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

COPY .env .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -a -installsuffix cgo -o main ./cmd/auth_service/main.go

FROM alpine:latest  

RUN apk --no-cache add netcat-openbsd

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/.env .

COPY entrypoint.sh /root/entrypoint.sh  

RUN chmod +x /root/entrypoint.sh  

EXPOSE 8080

ENTRYPOINT ["/root/entrypoint.sh"]

CMD ["./main"]