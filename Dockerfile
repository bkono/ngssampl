# syntax=docker/dockerfile:1
FROM golang:1.16
WORKDIR /go/src/github.com/bkono/ngssampl/
COPY main.go .
COPY go.mod .
COPY go.sum .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o app .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/bkono/ngssampl/app .
COPY sampler.creds .
CMD ["./app", "-creds", "sampler.creds", "-sub"] 
