FROM golang:1.21.6-alpine3.18@sha256:869193e7c30611d635c7bc3d1ed879039b7d24710a03474437d402f06825171e AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.21.6-alpine3.18@sha256:869193e7c30611d635c7bc3d1ed879039b7d24710a03474437d402f06825171e
# When running as non-root user, GOCACHE must be set to a directory
# that is writable by that user. It will otherwise default to /.cache/go-build,
# which is owned by root.
# https://github.com/golang/go/issues/26280#issuecomment-445294378
ENV GOCACHE=/tmp/go-build
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
