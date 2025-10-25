FROM golang:1.25.3-alpine3.22 AS builder

WORKDIR /staffy-sso
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o staffysso ./cmd/main.go

# ---
FROM alpine:latest
COPY --from=builder staffy-sso/.env .
COPY --from=builder staffy-sso/staffysso .
COPY --from=builder staffy-sso/configs/ ./configs

EXPOSE 50051
CMD ["./staffysso"]