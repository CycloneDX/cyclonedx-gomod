FROM golang:1.18.3-alpine3.15@sha256:f9181168749690bddb6751b004e976bf5d427425e0cfb50522e92c06f761def7 as build
ARG VERSION=latest
WORKDIR /tmp/cyclonedx-gomod
RUN apk --no-cache add git make
COPY . .
RUN make install

FROM golang:1.18.3-alpine3.15@sha256:f9181168749690bddb6751b004e976bf5d427425e0cfb50522e92c06f761def7
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
