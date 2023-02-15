FROM golang:1.20.1-alpine3.16@sha256:9266e89c290fe79635bda268c5edf3334ff76950db2416e8d57fc9ecc869f859 AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.20.1-alpine3.16@sha256:9266e89c290fe79635bda268c5edf3334ff76950db2416e8d57fc9ecc869f859
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
