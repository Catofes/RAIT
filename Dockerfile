FROM golang:alpine as builder
RUN apk add git
COPY . /src
RUN GO11MODULE=on go build -o /go/bin/rait /src/cmd/rait/rait.go

FROM alpine:edge
RUN apk add --no-cache iproute2
COPY --from=builder /go/bin/rait /usr/bin/rait
ENTRYPOINT ["rait"]

