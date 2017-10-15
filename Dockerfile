FROM golang:1.9

WORKDIR /go/src/app
COPY src/ .

EXPOSE 80
CMD go run *.go --port=80 -cpuMax=4 -root=/var/www/html
