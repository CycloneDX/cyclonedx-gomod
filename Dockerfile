FROM golang:1.18.5-alpine3.16@sha256:8e45e2ef37d2b6d98900392029db2bc88f42c0f2a9a8035fa7da90014698e86b AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.18.5-alpine3.16@sha256:8e45e2ef37d2b6d98900392029db2bc88f42c0f2a9a8035fa7da90014698e86b
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
