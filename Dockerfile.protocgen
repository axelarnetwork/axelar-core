# syntax=docker/dockerfile:experimental

FROM tendermintdev/sdk-proto-gen:v0.2 as build

# Remove the outdated Go installation
RUN rm -rf /usr/local/go

COPY --from=golang:1.23-alpine /usr/local/go/ /usr/local/go/

RUN apk add --no-cache --update \
  git \
  ca-certificates \
  nodejs

WORKDIR /workspace

COPY ./go.mod .
COPY ./go.sum .

RUN go mod download
RUN go install github.com/regen-network/cosmos-proto/protoc-gen-gocosmos
RUN npm install -g yarn
