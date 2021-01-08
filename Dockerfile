# syntax=docker/dockerfile:experimental

FROM golang:1.15-alpine3.12 as build

RUN apk add --no-cache --update \
  openssh-client \
  git \
  ca-certificates \
  make

WORKDIR axelar

RUN git config --global url."git@github.com:axelarnetwork".insteadOf https://github.com/axelarnetwork
RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts

COPY ./go.mod .
COPY ./go.sum .
RUN --mount=type=ssh go mod download

COPY . .
ENV CGO_ENABLED=0
RUN make build

FROM alpine:3.12

COPY --from=build /go/axelar/bin/axelar* /root/
ENV PATH="/root:${PATH}"
