FROM golang:alpine as builder

WORKDIR /go/src/github.com/gpsamson/segment-redis-proxy
ADD . .

RUN CGO_ENABLED=0 go build -ldflags '-s -w -extldflags "-static"' -o /usr/local/bin/segment-redis-proxy ./main.go

FROM alpine

COPY --from=builder /usr/local/bin/segment-redis-proxy /usr/local/bin/segment-redis-proxy

CMD ["/usr/local/bin/segment-redis-proxy"]
