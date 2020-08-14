# Build container
FROM golang:1.14 AS build
ADD . /code
WORKDIR /code
RUN go mod download
RUN CGO_ENABLED=0 go build -o /awsses_exporter .

# Deploy/Run Container
FROM alpine
LABEL maintainer="Joe Adams - @sysadmind"

COPY --from=build /awsses_exporter /bin/awsses_exporter

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

ENTRYPOINT ["/bin/awsses_exporter"]
EXPOSE     9199
