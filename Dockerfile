FROM golang:alpine as builder
RUN apk add git
COPY . /src
WORKDIR /src/cmd/rait
RUN go install

FROM alpine:edge
RUN apk add --no-cache iproute2
COPY --from=builder /go/bin/rait /usr/bin/rait
ENTRYPOINT ["rait"]

