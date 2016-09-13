FROM golang:1.7.1-alpine

RUN apk update && apk upgrade && apk add git

RUN mkdir -p /go/src/github.com/byuoitav
ADD . /go/src/github.com/byuoitav/raspi-deployment-microservice

WORKDIR /go/src/github.com/byuoitav/raspi-deployment-microservice
RUN go get -d -v
RUN go install -v

CMD ["/go/bin/raspi-deployment-microservice"]

EXPOSE 8008
