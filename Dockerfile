FROM golang:1.19-alpine AS builder

WORKDIR /build

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN pwd
RUN go build -ldflags="-s -w" -o ./server ./cmd/server/server.go

FROM scratch

COPY --from=builder ["/build/server", "/build/config.toml", "/"]

EXPOSE 8080

ENTRYPOINT ["/server", "-config", "config.toml"]