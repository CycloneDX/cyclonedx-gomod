FROM golang:1.18.2-alpine3.15@sha256:e6b729ae22a2f7b6afcc237f7b9da3a27151ecbdcd109f7ab63a42e52e750262 as build
ARG VERSION=latest
WORKDIR /tmp/cyclonedx-gomod
RUN apk --no-cache add git make
COPY . .
RUN make install

FROM golang:1.18.2-alpine3.15@sha256:e6b729ae22a2f7b6afcc237f7b9da3a27151ecbdcd109f7ab63a42e52e750262
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
