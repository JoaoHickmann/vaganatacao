FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git build-base ca-certificates

WORKDIR $GOPATH/src/github.com/JoaoHickmann/vaganatacao
COPY src/* ./

RUN go get -d -v
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/vaganatacao

FROM alpine

ARG TELEGRAM_API_KEY_ARG
ARG TELEGRAM_CHANNEL_ID_ARG
ARG DB_PATH_ARG

ENV TELEGRAM_API_KEY=$TELEGRAM_API_KEY_ARG
ENV TELEGRAM_CHANNEL_ID=$TELEGRAM_CHANNEL_ID_ARG
ENV DB_PATH=$DB_PATH_ARG

VOLUME /data

COPY --from=builder /go/bin/vaganatacao /go/bin/vaganatacao
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/go/bin/vaganatacao"]