FROM golang:1.20.0-alpine3.16@sha256:86eb22dea4141189e5656abd0616880387f0515a8d17ee1ef4b2ac79eb6e4ded AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.20.0-alpine3.16@sha256:86eb22dea4141189e5656abd0616880387f0515a8d17ee1ef4b2ac79eb6e4ded
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
