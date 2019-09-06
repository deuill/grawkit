FROM golang:1.13 AS builder
WORKDIR /mnt

COPY play/go.mod play/go.sum play/play.go /mnt/
RUN go build -o play play.go

FROM debian:stable-slim
WORKDIR /play

COPY --from=builder /mnt/play .
COPY grawkit .
COPY play/static static

USER nobody
EXPOSE 8080

CMD ["/play/play", "-script-path", "grawkit", "-listen-address", ":8080"]
