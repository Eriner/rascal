FROM golang:alpine AS build
WORKDIR $GOPATH/src/github.com/Eriner/rascal
COPY . .
RUN apk add --no-cache -U upx && \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/rascal && \
	upx /go/bin/rascal

FROM alpine:latest as alpine
WORKDIR /usr/share/zoneinfo
RUN apk -U --no-cache add tzdata zip ca-certificates && \
	zip -r -0 /zoneinfo.zip .

FROM scratch
COPY --from=build /go/bin/rascal .
ENV ZONEINFO /zoneinfo.zip
COPY --from=alpine /zoneinfo.zip /
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["./rascal"]
