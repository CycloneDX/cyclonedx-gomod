FROM golang:1.17.5-alpine3.15@sha256:4918412049183afe42f1ecaf8f5c2a88917c2eab153ce5ecf4bf2d55c1507b74 as build
ARG VERSION=latest
WORKDIR /tmp/cyclonedx-gomod
RUN apk --no-cache add git make
COPY . .
RUN make install

FROM golang:1.17.5-alpine3.15@sha256:4918412049183afe42f1ecaf8f5c2a88917c2eab153ce5ecf4bf2d55c1507b74
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
