# Build container
FROM golang:1.9 AS build
ADD . /go/src/github.com/sysadmind/awsses_exporter
RUN cd /go/src/github.com/sysadmind/awsses_exporter && go get
RUN CGO_ENABLED=0 go install github.com/sysadmind/awsses_exporter

# Deploy/Run Container
FROM alpine
LABEL maintainer="Joe Adams - @sysadmind"

COPY --from=build /go/bin/awsses_exporter /bin/awsses_exporter

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

ENTRYPOINT ["/bin/awsses_exporter"]
EXPOSE     9199
