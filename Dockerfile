FROM golang:1.17.3-alpine3.15@sha256:a207b29286084e7342286de809756f61558b00b81f794406399027631e0dba8b as build
ARG VERSION=latest
WORKDIR /tmp/cyclonedx-gomod
RUN apk --no-cache add git make
COPY . .
RUN make install

FROM golang:1.17.3-alpine3.15@sha256:a207b29286084e7342286de809756f61558b00b81f794406399027631e0dba8b
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
