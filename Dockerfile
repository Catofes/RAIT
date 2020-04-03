FROM golang:alpine as builder
RUN apk add git
COPY . /src
WORKDIR /src/cmd/rait
RUN go install

FROM nickcao/router:latest
COPY --from=builder /go/bin/rait /usr/bin/rait
COPY misc/docker-entry /usr/bin/entry
ENTRYPOINT ["entry"]
