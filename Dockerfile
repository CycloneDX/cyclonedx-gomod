FROM golang:1.24.1-alpine3.20@sha256:3d9132b88a6317b846b55aa8e821821301906fe799932ecbc4f814468c6977a5 AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.24.1-alpine3.20@sha256:3d9132b88a6317b846b55aa8e821821301906fe799932ecbc4f814468c6977a5
# When running as non-root user, GOCACHE must be set to a directory
# that is writable by that user. It will otherwise default to /.cache/go-build,
# which is owned by root.
# https://github.com/golang/go/issues/26280#issuecomment-445294378
ENV GOCACHE=/tmp/go-build
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
