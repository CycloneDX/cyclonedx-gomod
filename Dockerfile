FROM golang:1.18rc1-alpine3.15@sha256:a8c33744dfc0b56098eaaea862c35a944b6df0eb92cead5bb43acf620fb14080 as build
ARG VERSION=latest
WORKDIR /tmp/cyclonedx-gomod
RUN apk --no-cache add git make
COPY . .
RUN make install

FROM golang:1.18rc1-alpine3.15@sha256:a8c33744dfc0b56098eaaea862c35a944b6df0eb92cead5bb43acf620fb14080
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
