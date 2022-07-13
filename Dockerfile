FROM golang:1.18.4-alpine3.16@sha256:46f1fa18ca1ec228f7ea4978ad717f0a8c5e51436e7b8efaf64011f7729886df AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.18.4-alpine3.16@sha256:46f1fa18ca1ec228f7ea4978ad717f0a8c5e51436e7b8efaf64011f7729886df
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
