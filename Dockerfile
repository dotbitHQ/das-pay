# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.16-buster AS build

ENV GOPROXY https://goproxy.cn,direct

WORKDIR /app

COPY . ./

RUN go build -ldflags -s -v -o das-pay cmd/main.go

##
## Deploy
##
FROM ubuntu

ARG TZ=Asia/Shanghai

RUN export DEBIAN_FRONTEND=noninteractive \
    && apt-get update \
    && apt-get install -y ca-certificates tzdata \
    && ln -fs /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone \
    && dpkg-reconfigure tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build /app/das-pay /app/das-pay
COPY --from=build /app/config/config.example.yaml /app/config/config.yaml

ENTRYPOINT ["/app/das-pay", "--config", "/app/config/config.yaml"]
