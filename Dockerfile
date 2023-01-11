FROM golang:1.19.5-alpine3.16@sha256:65060885c8882f0119d77edcb42b414a0b96f6a8d466e9bbd782311ae51bf9f1 AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.19.5-alpine3.16@sha256:65060885c8882f0119d77edcb42b414a0b96f6a8d466e9bbd782311ae51bf9f1
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
