FROM golang:1.14

ENV GOPATH /go
RUN apt-get update && \
    apt-get install librdkafka-dev

WORKDIR /go/src/github.com/im-so-sorry/streaming-vk
ADD . /go/src/github.com/im-so-sorry/streaming-vk
RUN go install github.com/im-so-sorry/streaming-vk

ENTRYPOINT /go/bin/streaming-vk

#RUN go get -d -v ./...
#RUN go install -v ./...

#RUN go build

CMD ["streaming-vk"]