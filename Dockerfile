# Builder
FROM golang:1.18 AS builder
WORKDIR /go/src/user-service
COPY . .
ARG GOOS=linux
ARG GOARCH=amd64
RUN CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -o /bin/user-service ./cmd/server

# Image
FROM alpine:3.14
COPY --from=builder /bin/user-service /bin/user-service
ENTRYPOINT /bin/user-service server
EXPOSE 9000 9001
