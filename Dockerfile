FROM golang:1.20.2-alpine3.16@sha256:50e46c122d9b4e9502be0678b52a807be98f54d3fcab417df986bbc4d210dbf7 AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.20.2-alpine3.16@sha256:50e46c122d9b4e9502be0678b52a807be98f54d3fcab417df986bbc4d210dbf7
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
