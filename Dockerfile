FROM golang:1.24.0-alpine3.20@sha256:9fed4022a220fb64327baa90cddfd98607f3b816cb4f5769187500571f73072d AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.24.0-alpine3.20@sha256:9fed4022a220fb64327baa90cddfd98607f3b816cb4f5769187500571f73072d
# When running as non-root user, GOCACHE must be set to a directory
# that is writable by that user. It will otherwise default to /.cache/go-build,
# which is owned by root.
# https://github.com/golang/go/issues/26280#issuecomment-445294378
ENV GOCACHE=/tmp/go-build
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
