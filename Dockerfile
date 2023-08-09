FROM golang:alpine AS builder

ENV GO111MODULE="on" \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /var/app/main cmd/server/main.go

FROM alpine:3.18.3
RUN apk --no-cache add ca-certificates

COPY --from=builder /var/app/main /var/app/

EXPOSE 8080
ENTRYPOINT [ "/var/app/main" ]
