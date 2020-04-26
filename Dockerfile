FROM golang:1.14-alpine
RUN apk update && apk add sqlite-dev alpine-sdk

WORKDIR /go/src/github.com/thomas-maurice/wgnw
COPY . .

RUN make

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
RUN mkdir /data
WORKDIR /data
COPY --from=0 /go/src/github.com/thomas-maurice/wgnw/bin/wgnw /
COPY --from=0 /go/src/github.com/thomas-maurice/wgnw/bin/wgnwd /
COPY --from=0 /go/src/github.com/thomas-maurice/wgnw/bin/wgnw-server /
