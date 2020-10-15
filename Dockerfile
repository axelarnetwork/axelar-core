FROM golang:1.15 as build
WORKDIR axelar
COPY ./tssd ./tssd
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download
COPY . .

ENV CGO_ENABLED=0
RUN make build

FROM alpine:3.12
COPY --from=build /go/axelar/bin/axelar* /root/
ENV PATH="/root:${PATH}"
