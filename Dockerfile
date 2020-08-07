FROM golang:alpine as build
ADD . rait
WORKDIR rait
RUN apk add make git
RUN make

FROM scratch
COPY --from=build /go/rait/bin/rait /rait

ENTRYPOINT ["/rait"]
