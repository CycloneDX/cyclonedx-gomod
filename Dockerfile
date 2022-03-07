FROM golang:1.17.8-alpine3.15@sha256:b35984144ec2c2dfd6200e112a9b8ecec4a8fd9eff0babaff330f1f82f14cb2a as build
ARG VERSION=latest
WORKDIR /tmp/cyclonedx-gomod
RUN apk --no-cache add git make
COPY . .
RUN make install

FROM golang:1.17.8-alpine3.15@sha256:b35984144ec2c2dfd6200e112a9b8ecec4a8fd9eff0babaff330f1f82f14cb2a
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
