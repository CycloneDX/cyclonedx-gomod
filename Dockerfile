FROM golang:1.18.4-alpine3.16@sha256:d84b1ff3eeb9404e0a7dda7fdc6914cbe657102420529beec62ccb3ef3d143eb AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.18.4-alpine3.16@sha256:d84b1ff3eeb9404e0a7dda7fdc6914cbe657102420529beec62ccb3ef3d143eb
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
