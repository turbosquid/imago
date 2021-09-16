FROM golang:1.17-alpine as builder
LABEL description="farmhand build"
MAINTAINER mowings@turbosquid.com
ENV GOPATH=/go:/app:/app/vendor
RUN apk add git
COPY  . /app/
WORKDIR /app/
RUN go build -v

FROM alpine
RUN apk update && apk add bash imagemagick
RUN rm -rf /var/cache/apk/*
RUN mkdir -p /app
COPY --from=builder /app/imago  /app/imago
COPY settings.example.yml /app/settings.yml
WORKDIR /app
ENTRYPOINT ["/app/imago"]
