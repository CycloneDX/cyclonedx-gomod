FROM gitpod/workspace-go:latest

USER root

RUN \
    wget -O "/usr/local/bin/cyclonedx" https://github.com/CycloneDX/cyclonedx-cli/releases/download/v0.24.0/cyclonedx-linux-x64 && \
    echo "691cf7ed82ecce1f85e6d21bccd1ed2d7968e40eb6be7504b392c8b3a0943891 /usr/local/bin/cyclonedx" | sha256sum -c && \
    chmod +x "/usr/local/bin/cyclonedx"

USER gitpod