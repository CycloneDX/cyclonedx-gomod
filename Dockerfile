FROM golang:1.17.3-alpine3.14@sha256:7bd9bf011a76c21b70a1e507de7f8d67656e1f293fd72af80aa46f7b054db515 as build
ARG VERSION=latest
WORKDIR /tmp/cyclonedx-gomod
RUN apk --no-cache add git make
COPY . .
RUN make install

FROM golang:1.17.3-alpine3.14@sha256:7bd9bf011a76c21b70a1e507de7f8d67656e1f293fd72af80aa46f7b054db515
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
