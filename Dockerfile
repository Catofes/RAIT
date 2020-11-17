FROM golang:alpine as build
ADD . rait
WORKDIR rait
RUN apk add make git
RUN make

FROM alpine:edge
COPY --from=build /go/rait/bin/rait /usr/local/bin/rait
COPY --from=build /go/rait/bin/info /usr/local/bin/info

ENTRYPOINT ["/usr/local/bin/rait"]
