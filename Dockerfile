FROM golang:1.20.2-alpine3.16@sha256:0848e987c1be480a253637c82e4e08f3c6589b644804320d14e7a2321326f98f AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.20.2-alpine3.16@sha256:0848e987c1be480a253637c82e4e08f3c6589b644804320d14e7a2321326f98f
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
