FROM golang:1.16-alpine as build
ARG VERSION=latest
WORKDIR /tmp/cyclonedx-gomod
COPY . .
RUN go install

FROM golang:1.16-alpine
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
