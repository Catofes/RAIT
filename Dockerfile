FROM golang:alpine as build
ADD . rait
WORKDIR rait
RUN apk add make git
RUN make

FROM gcr.io/distroless/static
COPY --from=build /go/rait/bin/rait /usr/local/bin/rait

ENTRYPOINT ["/usr/local/bin/rait"]
