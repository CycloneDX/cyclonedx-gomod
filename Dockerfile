FROM golang:1.22.2-alpine3.18@sha256:d995eb689a0c123590a3d34c65f57f3a118bda3fa26f92da5e089ae7d8fd81a0 AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.22.2-alpine3.18@sha256:d995eb689a0c123590a3d34c65f57f3a118bda3fa26f92da5e089ae7d8fd81a0
# When running as non-root user, GOCACHE must be set to a directory
# that is writable by that user. It will otherwise default to /.cache/go-build,
# which is owned by root.
# https://github.com/golang/go/issues/26280#issuecomment-445294378
ENV GOCACHE=/tmp/go-build
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
