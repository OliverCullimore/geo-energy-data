## Define BUILD_VERSION
ARG BUILD_VERSION
## Import base golang image
FROM golang:1.14-alpine AS builder
## Install git for fetching the dependencies
RUN apk update && apk add --no-cache git
## Create an app directory to contain all of the code
RUN mkdir /app
## Create a config directory
RUN mkdir /config
## Copy the current directory into our /app directory
ADD . /app
## Set workdir
WORKDIR /app
## Build go application and specify the name of the executable
RUN  go install github.com/olivercullimore/geo-energy-data/server && go build -o main -ldflags="-X main.BuildVersion=$BUILD_VERSION" .

## Import fresh base alpine image
FROM alpine:latest
## Create an app directory to contain all of the code
RUN mkdir /app
## Create a config directory
RUN mkdir /config
## Copy the built executable into /app
COPY --from=builder /app/main /app/
## Set workdir
WORKDIR /app
## Trigger our newly built Go program
CMD ["/app/main"]
## expose necessary ports
#EXPOSE 80