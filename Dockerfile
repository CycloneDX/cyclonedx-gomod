FROM golang:1.17.4-alpine3.15@sha256:d8bd35607c405fcef71d749aa367f86954706c3a57a602fb0bcaae3581043f8f as build
ARG VERSION=latest
WORKDIR /tmp/cyclonedx-gomod
RUN apk --no-cache add git make
COPY . .
RUN make install

FROM golang:1.17.4-alpine3.15@sha256:d8bd35607c405fcef71d749aa367f86954706c3a57a602fb0bcaae3581043f8f
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
