FROM golang:alpine as builder
RUN apk add git
COPY . /src
RUN cd /src/cmd/rait && go install

FROM alpine:edge
COPY --from=builder /go/bin/rait /usr/local/bin/rait
ENTRYPOINT rait
