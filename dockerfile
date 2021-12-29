## Build
##
FROM golang:1.17-buster AS build
WORKDIR /usr/local/go/src/IntraProxy/

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY cmd ./cmd

RUN go build -o /IntraProxy ./cmd/

##
## Deploy
##
FROM gcr.io/distroless/base-debian10

WORKDIR /srv/IntraProxy/

EXPOSE 443

COPY --from=build /IntraProxy /srv/IntraProxy/IntraProxy

ENTRYPOINT ["/srv/IntraProxy/IntraProxy"]