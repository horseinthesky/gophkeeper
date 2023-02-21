FROM golang:1.19-alpine3.17 AS builder
WORKDIR /app
COPY . .
RUN go build -o gs ./server.go

FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/gs .
COPY ./certs ./certs
COPY ./db/migrations ./db/migrations
ENTRYPOINT [ "./gs" ]
