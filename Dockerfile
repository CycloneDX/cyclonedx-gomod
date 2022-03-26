FROM golang:1.18.0-alpine3.15@sha256:bb6ae029f163091e27c15094dba9b63429e301a7a6856cf1427439efe94e95f1 as build
ARG VERSION=latest
WORKDIR /tmp/cyclonedx-gomod
RUN apk --no-cache add git make
COPY . .
RUN make install

FROM golang:1.18.0-alpine3.15@sha256:bb6ae029f163091e27c15094dba9b63429e301a7a6856cf1427439efe94e95f1
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
