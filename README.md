# cyclonedx-gomod

[![Build Status](https://github.com/CycloneDX/cyclonedx-gomod/actions/workflows/ci.yml/badge.svg)](https://github.com/CycloneDX/cyclonedx-gomod/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Apache%202.0-brightgreen.svg)](LICENSE)
[![Website](https://img.shields.io/badge/https://-cyclonedx.org-blue.svg)](https://cyclonedx.org/)
[![Slack Invite](https://img.shields.io/badge/Slack-Join-blue?logo=slack&labelColor=393939)](https://cyclonedx.org/slack/invite)
[![Group Discussion](https://img.shields.io/badge/discussion-groups.io-blue.svg)](https://groups.io/g/CycloneDX)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social&label=Follow)](https://twitter.com/CycloneDX_Spec)

*cyclonedx-gomod* creates CycloneDX Software Bill of Materials (SBOM) from Go modules

## Installation

Prebuilt binaries are available on the [releases](https://github.com/CycloneDX/cyclonedx-gomod/releases) page.

### From Source

```
go get github.com/CycloneDX/cyclonedx-gomod
```

Building from source requires Go 1.16 or newer.

## Usage

```
Usage of cyclonedx-gomod:
  -json
        Output in JSON format
  -module string
        Path to Go module (default ".")
  -noserial
        Omit serial number
  -output string
        Output path (default "-")
  -serial string
        Serial number (default [random UUID])
  -type string
        Type of the main component (default "application")
  -version
        Show version
```

In order to be able to calculate hashes, all modules have to be present in Go's module cache.  
Make sure to run `go mod download` before generating BOMs with *cyclonedx-gomod*.

### Example

```
$ go mod tidy
$ go mod download
$ cyclonedx-gomod -output bom.xml 
```

```xml
<?xml version="1.0" encoding="UTF-8"?>
<bom xmlns="http://cyclonedx.org/schema/bom/1.2" serialNumber="urn:uuid:07c19f2c-6ea7-4258-befd-19e9bc019183" version="1">
    <metadata>
        <timestamp>2021-03-08T18:49:41+01:00</timestamp>
        <tools>
            <tool>
                <vendor>CycloneDX</vendor>
                <name>cyclonedx-gomod</name>
                <version>v0.0.0-unset</version>
                <hashes>
                    <hash alg="MD5">31e8977ccf58f1dd081d5f15f248c45e</hash>
                    <hash alg="SHA-1">fcbdb1485eaa54afdac6901fde3266d9d4517505</hash>
                    <hash alg="SHA-256">940e64bb70b2bbb827f9fe3ca719324d08a1afed087ba1331311c6838eddc2d0</hash>
                    <hash alg="SHA-512">4505b70b028a0c384459a02eb5fd2fe008763c2ea8640cc97e2f75626e04c03eab4c95acfbea250703c8049590791f9feebe3cdbc954cca042fb8050e7c0c3bf</hash>
                </hashes>
            </tool>
        </tools>
        <component bom-ref="pkg:golang/github.com/CycloneDX/cyclonedx-gomod@v0.0.0-20210308115936-fe548e553e56" type="application">
            <name>github.com/CycloneDX/cyclonedx-gomod</name>
            <version>v0.0.0-20210308115936-fe548e553e56</version>
            <purl>pkg:golang/github.com/CycloneDX/cyclonedx-gomod@v0.0.0-20210308115936-fe548e553e56</purl>
        </component>
    </metadata>
    <components>
        <component bom-ref="pkg:golang/github.com/CycloneDX/cyclonedx-go@v0.1.0" type="library">
            <name>github.com/CycloneDX/cyclonedx-go</name>
            <version>v0.1.0</version>
            <scope>required</scope>
            <hashes>
                <hash alg="SHA-256">c92dc729b69e0f3c13262d3ec62a6021f7060eb8e4af75e17d7e89b28f790588</hash>
            </hashes>
            <purl>pkg:golang/github.com/CycloneDX/cyclonedx-go@v0.1.0</purl>
        </component>
        <!-- ... -->
    </components>
</bom>
```

## License

Permission to modify and redistribute is granted under the terms of the Apache 2.0 license.  
See the [LICENSE](./LICENSE) file for the full license.

## Contributing

Pull requests are welcome. But please read the
[CycloneDX contributing guidelines](https://github.com/CycloneDX/.github/blob/master/CONTRIBUTING.md) first.

It is generally expected that pull requests will include relevant tests. Tests are automatically run against all
supported Go versions for every pull request.
