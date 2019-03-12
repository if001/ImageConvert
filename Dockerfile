FROM golang:1.11.5 as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /go/src/ImageConvert
COPY . .

RUN go get github.com/nfnt/resize
RUN go get github.com/muesli/smartcrop &&\
cd $GOPATH/src/github.com/muesli/smartcrop &&\
go get -u -v &&\
go build && go test -v
Run make


FROM alpine:3.9
RUN apk add --no-cache --virtual build-dependencies gcc make autoconf libc-dev libtool &&\
    apk add --no-cache imagemagick
COPY --from=builder /go/src/ImageConvert/app /go_app/app
COPY --from=builder /go/src/ImageConvert/sphericalpano2rect /go_app/sphericalpano2rect
RUN mkdir -p /go_app/img
EXPOSE 8080
ENTRYPOINT ["/go_app/app"]