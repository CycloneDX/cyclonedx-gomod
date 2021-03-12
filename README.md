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
        <timestamp>2021-03-13T00:31:25+01:00</timestamp>
        <tools>
            <tool>
                <vendor>CycloneDX</vendor>
                <name>cyclonedx-gomod</name>
                <version>v0.0.0-unset</version>
                <hashes>
                    <hash alg="MD5">d6856603b707ca75498c3bdd652aaebe</hash>
                    <hash alg="SHA-1">d2cf76a375a0fb7e4692ed3ee1a22de3a53aba63</hash>
                    <hash alg="SHA-256">dc73154f76ae14d1ba2150713de59d156d733de57763924cd6fd71386ba3b8ca</hash>
                    <hash alg="SHA-512">fbba3204590549cca545709e671f7191824acce57a6ad25432208c4dd662ce3475a62012602c716868ed0490ca96059e14f4c13e8e3a3474ad6dedada0a1a280</hash>
                </hashes>
            </tool>
        </tools>
        <component bom-ref="pkg:golang/github.com/CycloneDX/cyclonedx-gomod@v0.0.0-20210312235402-7b06d181cac7" type="application">
            <name>github.com/CycloneDX/cyclonedx-gomod</name>
            <version>v0.0.0-20210312235402-7b06d181cac7</version>
            <purl>pkg:golang/github.com/CycloneDX/cyclonedx-gomod@v0.0.0-20210312235402-7b06d181cac7</purl>
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
    <dependencies>
        <dependency ref="pkg:golang/github.com/CycloneDX/cyclonedx-gomod@v0.0.0-20210312235402-7b06d181cac7">
            <dependency ref="pkg:golang/github.com/CycloneDX/cyclonedx-go@v0.1.0"></dependency>
            <dependency ref="pkg:golang/github.com/google/uuid@v1.2.0"></dependency>
            <dependency ref="pkg:golang/github.com/stretchr/testify@v1.7.0"></dependency>
            <dependency ref="pkg:golang/golang.org/x/mod@v0.4.2"></dependency>
        </dependency>
        <dependency ref="pkg:golang/github.com/CycloneDX/cyclonedx-go@v0.1.0">
            <dependency ref="pkg:golang/github.com/bradleyjkemp/cupaloy/v2@v2.6.0"></dependency>
            <dependency ref="pkg:golang/github.com/stretchr/testify@v1.7.0"></dependency>
        </dependency>
        <!-- ... -->
    </dependencies>
</bom>
```

Checkout the [`examples`](./examples) directory for complete BOM examples.

### Replacements

By using the [`replace` directive](https://golang.org/ref/mod#go-mod-file-replace), users of Go modules can replace the 
content of a given module, e.g.:

```
require github.com/jameskeane/bcrypt v0.0.0-20170924085257-7509ea014998
replace github.com/jameskeane/bcrypt => github.com/ProtonMail/bcrypt v0.0.0-20170924085257-7509ea014998
```

We consider the replaced module (`github.com/jameskeane/bcrypt`) to be the ancestor of the replacement 
(`github.com/ProtonMail/bcrypt`) and include it in the replacement's [pedigree](https://cyclonedx.org/use-cases/#pedigree):

```xml
<component bom-ref="pkg:golang/github.com/ProtonMail/bcrypt@v0.0.0-20170924085257-7509ea014998" type="library">
  <name>github.com/ProtonMail/bcrypt</name>
  <version>v0.0.0-20170924085257-7509ea014998</version>
  <scope>required</scope>
  <hashes>
    <hash alg="SHA-256">613dae57042245067109a69a8707dc813ab68f78faeb0d349ffdbb81bff3b9bb</hash>
  </hashes>
  <purl>pkg:golang/github.com/ProtonMail/bcrypt@v0.0.0-20170924085257-7509ea014998</purl>
  <pedigree>
    <ancestors>
      <component bom-ref="pkg:golang/github.com/jameskeane/bcrypt@v0.0.0-20170924085257-7509ea014998" type="library">
        <name>github.com/jameskeane/bcrypt</name>
        <version>v0.0.0-20170924085257-7509ea014998</version>
        <hashes>
          <hash alg="SHA-256">c510a93977f0fe9cf70bc2b8ec586828f64b985128d88a1f5d2e355b7e895f9f</hash>
        </hashes>
        <purl>pkg:golang/github.com/jameskeane/bcrypt@v0.0.0-20170924085257-7509ea014998</purl>
      </component>
    </ancestors>
  </pedigree>
</component>
```

The [dependency graph](https://cyclonedx.org/use-cases/#dependency-graph) will also reference the replacement, 
not the replaced module:

```xml
<dependencies>
    <dependency ref="pkg:golang/github.com/ProtonMail/proton-bridge@v0.0.0-20210210160947-565c0b6ddf0f">
        <dependency ref="pkg:golang/github.com/ProtonMail/bcrypt@v0.0.0-20170924085257-7509ea014998"></dependency>
        <!-- ... -->
    </dependency>
    <dependency ref="pkg:golang/github.com/ProtonMail/bcrypt@v0.0.0-20170924085257-7509ea014998"></dependency>
</dependencies>
```

### Hashes

*cyclonedx-gomod* uses the same hashing algorithm Go uses for its module integrity checks.  
[`vikyd/go-checksum`](https://github.com/vikyd/go-checksum#calc-checksum-of-module-directory) does a great job of
explaining what exactly that entails. In essence, the hash you see in a BOM should be the same as in your `go.sum` file,
just in a different format. This is because the CycloneDX specification enforces hashes to be provided in hex encoding,
while Go uses base64 encoded values.

To verify a hash found in a BOM, do the following:

1. Hex decode the value
2. Base64 encode the value
3. Prefix the value with `h1:`

Given the hex encoded hash `a8962d5e72515a6a5eee6ff75e5ca1aec2eb11446a1d1336931ce8c57ab2503b`, we'd end up with a
module checksum of `h1:qJYtXnJRWmpe7m/3XlyhrsLrEURqHRM2kxzoxXqyUDs=`. 
Now, query your [checksum database](https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md#checksum-database) 
for the expected checksum and compare the values.

## License

Permission to modify and redistribute is granted under the terms of the Apache 2.0 license.  
See the [LICENSE](./LICENSE) file for the full license.

## Contributing

Pull requests are welcome. But please read the
[CycloneDX contributing guidelines](https://github.com/CycloneDX/.github/blob/master/CONTRIBUTING.md) first.

It is generally expected that pull requests will include relevant tests. Tests are automatically run against all
supported Go versions for every pull request.
