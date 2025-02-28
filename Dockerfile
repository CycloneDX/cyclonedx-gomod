FROM golang:1.24.0-alpine3.20@sha256:79f7ffeff943577c82791c9e125ab806f133ae3d8af8ad8916704922ddcc9dd8 AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.24.0-alpine3.20@sha256:79f7ffeff943577c82791c9e125ab806f133ae3d8af8ad8916704922ddcc9dd8
# When running as non-root user, GOCACHE must be set to a directory
# that is writable by that user. It will otherwise default to /.cache/go-build,
# which is owned by root.
# https://github.com/golang/go/issues/26280#issuecomment-445294378
ENV GOCACHE=/tmp/go-build
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
