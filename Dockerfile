FROM golang:1.26rc3-alpine3.23@sha256:343c20fd6876bfb5ba9f46b0a452008b7dced3804e424ff7ada0ceadafad5c55 AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.26rc3-alpine3.23@sha256:343c20fd6876bfb5ba9f46b0a452008b7dced3804e424ff7ada0ceadafad5c55
# When running as non-root user, GOCACHE must be set to a directory
# that is writable by that user. It will otherwise default to /.cache/go-build,
# which is owned by root.
# https://github.com/golang/go/issues/26280#issuecomment-445294378
ENV GOCACHE=/tmp/go-build
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
