FROM gitpod/workspace-go:2022-02-04-10-54-10

USER root

RUN \
    wget -O "/usr/local/bin/cyclonedx" https://github.com/CycloneDX/cyclonedx-cli/releases/download/v0.22.0/cyclonedx-linux-x64 && \
    echo "ae39404a9dc8b2e7be0a9559781ee9fe3492201d2629de139d702fd4535ffdd6 /usr/local/bin/cyclonedx" | sha256sum -c && \
    chmod +x "/usr/local/bin/cyclonedx"

USER gitpod