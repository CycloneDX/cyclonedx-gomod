# cyclonedx-gomod

[![Build Status](https://github.com/CycloneDX/cyclonedx-gomod/actions/workflows/ci.yml/badge.svg)](https://github.com/CycloneDX/cyclonedx-gomod/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Apache%202.0-brightgreen.svg)](LICENSE)
[![Website](https://img.shields.io/badge/https://-cyclonedx.org-blue.svg)](https://cyclonedx.org/)
[![Slack Invite](https://img.shields.io/badge/Slack-Join-blue?logo=slack&labelColor=393939)](https://cyclonedx.org/slack/invite)
[![Group Discussion](https://img.shields.io/badge/discussion-groups.io-blue.svg)](https://groups.io/g/CycloneDX)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social&label=Follow)](https://twitter.com/CycloneDX_Spec)

*cyclonedx-gomod* creates CycloneDX Software Bill of Materials (SBOM) from Go modules

## Installation

```
go get github.com/CycloneDX/cyclonedx-gomod
```

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
  -version
        Show version
```

In order to be able to calculate hashes, all modules have to be present in Go's module cache.  
Make sure you run `go mod download` before generating BOMs with *cyclonedx-gomod*.

### Example

```
$ go mod tidy
$ go mod download
$ cyclonedx-gomod -json -output bom.json 
```

```json
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.2",
  "serialNumber": "urn:uuid:60dc90e1-807b-4297-b919-13fc0bd01d40",
  "version": 1,
  "metadata": {
    "timestamp": "2021-03-07T22:34:13+01:00",
    "tools": [
      {
        "vendor": "CycloneDX",
        "name": "cyclonedx-gomod",
        "version": "v0.0.0-unset"
      }
    ],
    "component": {
      "bom-ref": "pkg:golang/github.com/CycloneDX/cyclonedx-gomod@v0.0.0-20210307202123-2cd971b532b3",
      "type": "application",
      "name": "github.com/CycloneDX/cyclonedx-gomod",
      "version": "v0.0.0-20210307202123-2cd971b532b3",
      "purl": "pkg:golang/github.com/CycloneDX/cyclonedx-gomod@v0.0.0-20210307202123-2cd971b532b3"
    }
  },
  "components": [
    {
      "bom-ref": "pkg:golang/github.com/CycloneDX/cyclonedx-go@v0.1.0",
      "type": "library",
      "name": "github.com/CycloneDX/cyclonedx-go",
      "version": "v0.1.0",
      "scope": "required",
      "hashes": [
        {
          "alg": "SHA-256",
          "content": "c92dc729b69e0f3c13262d3ec62a6021f7060eb8e4af75e17d7e89b28f790588"
        }
      ],
      "purl": "pkg:golang/github.com/CycloneDX/cyclonedx-go@v0.1.0"
    },
    ...
  ]
}
```

## License

Permission to modify and redistribute is granted under the terms of the Apache 2.0 license.  
See the [LICENSE](./LICENSE) file for the full license.

## Contributing

Pull requests are welcome. But please read the
[CycloneDX contributing guidelines](https://github.com/CycloneDX/.github/blob/master/CONTRIBUTING.md) first.

It is generally expected that pull requests will include relevant tests. Tests are automatically run against all
supported Go versions for every pull request.
