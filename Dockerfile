FROM golang:1.17.7-alpine3.15@sha256:c23027af83ff27f663d7983750a9a08f442adb2e7563250787b23ab3b6750d9e as build
ARG VERSION=latest
WORKDIR /tmp/cyclonedx-gomod
RUN apk --no-cache add git make
COPY . .
RUN make install

FROM golang:1.17.7-alpine3.15@sha256:c23027af83ff27f663d7983750a9a08f442adb2e7563250787b23ab3b6750d9e
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
