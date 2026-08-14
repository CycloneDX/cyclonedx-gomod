FROM golang:1.27rc2-alpine3.23@sha256:f12c2dc8d14504742f545658e8e49e09e545f2e396788b49797c8052f53434ba AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.27rc2-alpine3.23@sha256:f12c2dc8d14504742f545658e8e49e09e545f2e396788b49797c8052f53434ba
# When running as non-root user, GOCACHE must be set to a directory
# that is writable by that user. It will otherwise default to /.cache/go-build,
# which is owned by root.
# https://github.com/golang/go/issues/26280#issuecomment-445294378
ENV GOCACHE=/tmp/go-build
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
