FROM golang:1.23.4-alpine3.20@sha256:d8ec50621234b920502039e2a6937e7e6897a282cf654c76bb4cc9a79b6df4cb AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.23.4-alpine3.20@sha256:d8ec50621234b920502039e2a6937e7e6897a282cf654c76bb4cc9a79b6df4cb
# When running as non-root user, GOCACHE must be set to a directory
# that is writable by that user. It will otherwise default to /.cache/go-build,
# which is owned by root.
# https://github.com/golang/go/issues/26280#issuecomment-445294378
ENV GOCACHE=/tmp/go-build
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
