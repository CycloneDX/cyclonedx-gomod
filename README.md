# cyclonedx-gomod

[![Build Status](https://github.com/CycloneDX/cyclonedx-gomod/actions/workflows/ci.yml/badge.svg)](https://github.com/CycloneDX/cyclonedx-gomod/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/CycloneDX/cyclonedx-gomod)](https://goreportcard.com/report/github.com/CycloneDX/cyclonedx-gomod)
[![Latest GitHub release](https://img.shields.io/github/v/release/CycloneDX/cyclonedx-gomod?sort=semver)](https://github.com/CycloneDX/cyclonedx-gomod/releases/latest)
[![License](https://img.shields.io/badge/license-Apache%202.0-brightgreen.svg)](LICENSE)
[![Website](https://img.shields.io/badge/https://-cyclonedx.org-blue.svg)](https://cyclonedx.org/)
[![Slack Invite](https://img.shields.io/badge/Slack-Join-blue?logo=slack&labelColor=393939)](https://cyclonedx.org/slack/invite)
[![Group Discussion](https://img.shields.io/badge/discussion-groups.io-blue.svg)](https://groups.io/g/CycloneDX)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social&label=Follow)](https://twitter.com/CycloneDX_Spec)

*cyclonedx-gomod* creates CycloneDX Software Bill of Materials (SBOM) from Go modules

## Installation

Prebuilt binaries are available on the [releases](https://github.com/CycloneDX/cyclonedx-gomod/releases) page.

### From Source

```shell
go install github.com/CycloneDX/cyclonedx-gomod@v0.10.0
```

Building from source requires Go 1.16 or newer.

## Compatibility

*cyclonedx-gomod* will produce BOMs for the latest version of the CycloneDX specification 
[supported by cyclonedx-go](https://github.com/CycloneDX/cyclonedx-go#compatibility), which currently is [1.3](https://cyclonedx.org/docs/1.3/). 
You can use the [CycloneDX CLI](https://github.com/CycloneDX/cyclonedx-cli#convert-command) to convert between multiple 
BOM formats or specification versions. 

## Usage

```
USAGE
  cyclonedx-gomod <SUBCOMMAND> [FLAGS...] [<ARG>...]

SUBCOMMANDS
  app      Generate SBOM for an application
  bin      Generate SBOM for a binary
  mod      Generate SBOM for a module
  version  Show version information
```

### Subcommands

#### `app`

```
USAGE
  cyclonedx-gomod app [FLAGS...] MODPATH

Generate SBOM for an application.

In order to produce accurate results, build constraints must be configured
via environment variables. These build constraints should mimic the ones passed
to the "go build" command for the application.

Noteworthy environment variables that act as build constraints are:
  - GOARCH       The target architecture (386, amd64, etc.)
  - GOOS         The target operating system (linux, windows, etc.)
  - CGO_ENABLED  Whether or not CGO is enabled
  - GOFLAGS      Pass build tags

A complete overview of all environment variables can be found here:
  https://pkg.go.dev/cmd/go#hdr-Environment_variables

Unless the -reproducible flag is provided, build constraints will be 
included as properties of the main component.

The -main flag should be used to specify the path to the application's main file.
-main must point to a go file within MODPATH. If -main is not specified, "main.go" is assumed.

By passing -files, all files that would be compiled into the binary will be included
as subcomponents of their respective module. Files versions follow the v0.0.0-SHORTHASH pattern, 
where SHORTHASH is the first 12 characters of the file's SHA1 hash.

Examples:
  $ GOARCH=arm64 GOOS=linux GOFLAGS="-tags=foo,bar" cyclonedx-gomod app -output linux-arm64.bom.xml
  $ cyclonedx-gomod app -json -output acme-app.bom.json -files -licenses -main cmd/acme-app/main.go /usr/src/acme-module

FLAGS
  -files=false         Include files
  -json=false          Output in JSON
  -licenses=false      Resolve module licenses
  -main main.go        Path to the application's main file, relative to MODPATH
  -noserial=false      Omit serial number
  -novprefix=false     Omit "v" prefix from versions
  -output -            Output file path (or - for STDOUT)
  -reproducible=false  Make the SBOM reproducible by omitting dynamic content
  -serial ...          Serial number
  -std=false           Include Go standard library as component and dependency of the module
  -verbose=false       Enable verbose output
```

#### `bin`

```
USAGE
  cyclonedx-gomod bin [FLAGS...] PATH

Generate SBOM for a binary.

When license resolution is enabled, all modules (including the main module) 
will be downloaded to the module cache using "go mod download".

Please note that data embedded in binaries shouldn't be trusted,
unless there's solid evidence that the binaries haven't been modified
since they've been built.

Example:
  $ cyclonedx-gomod bin -json -output minikube-v1.22.0.bom.json -version v1.22.0 ./minikube

FLAGS
  -json=false          Output in JSON
  -licenses=false      Resolve module licenses
  -noserial=false      Omit serial number
  -novprefix=false     Omit "v" prefix from versions
  -output -            Output file path (or - for STDOUT)
  -reproducible=false  Make the SBOM reproducible by omitting dynamic content
  -serial ...          Serial number
  -std=false           Include Go standard library as component and dependency of the module
  -verbose=false       Enable verbose output
  -version ...         Version of the main component
```

#### `mod`

```
USAGE
  cyclonedx-gomod mod [FLAGS...] [PATH]

Generate SBOM for a module.

Examples:
  $ cyclonedx-gomod mod -licenses -type library -json -output bom.json ./cyclonedx-go
  $ cyclonedx-gomod mod -reproducible -test -output bom.xml ./cyclonedx-go

FLAGS
  -json=false          Output in JSON
  -licenses=false      Resolve module licenses
  -noserial=false      Omit serial number
  -novprefix=false     Omit "v" prefix from versions
  -output -            Output file path (or - for STDOUT)
  -reproducible=false  Make the SBOM reproducible by omitting dynamic content
  -serial ...          Serial number
  -std=false           Include Go standard library as component and dependency of the module
  -test=false          Include test dependencies
  -type application    Type of the main component
  -verbose=false       Enable verbose output
```

### Examples

Checkout the [`examples`](./examples) directory for examples of SBOMs generated with *cyclonedx-gomod*.

### GitHub Actions 🤖

We made a GitHub Action to help integrate *cyclonedx-gomod* into existing CI/CD workflows!  
You can find it on the GitHub marketplace: [*gh-gomod-generate-sbom*](https://github.com/marketplace/actions/cyclonedx-gomod-generate-sbom)

### Docker 🐳

```shell
$ docker run -it --rm \
    -v "/path/to/mymodule:/usr/src/mymodule" \
    -v "$(pwd):/out" \
    cyclonedx/cyclonedx-gomod:v1 mod -json -output /out/bom.json /usr/src/mymodule
```

## Important Notes

### Vendoring

Modules that use [vendoring](https://golang.org/ref/mod#go-mod-vendor) are, although in a limited manner, supported.  
Limitations are as follows:

* **No hashes.** Go doesn't copy all module files to `vendor`, only those that are required to build
  and test the main module. Because [module checksums](#hashes) consider almost all files in a module's directory though, 
  calculating accurate hashes from the `vendor` directory is not possible. As a consequence, BOMs for modules that use
  vendoring do not include component hashes.
* **License detection may fail.** Go doesn't always copy license files when vendoring modules, which may cause license detection to fail.

### Licenses

There is currently no standard way for developers to declare their module's license.  
Detecting licenses based on files in a repository is a non-trivial task, which is why *cyclonedx-gomod*  
uses [`go-license-detector`](https://github.com/go-enry/go-license-detector) to resolve module licenses.

While `go-license-detector`'s license matching *may* be accurate most of the time, BOMs should state facts.  
This is why license resolution is an opt-in feature (using the `-licenses` flag). If you are a vendor and legally
required to provide 100% accurate BOMs, **do not** use this feature.

### Hashes

*cyclonedx-gomod* uses the same hashing algorithm Go uses for its [module authentication](https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md#module-authentication-with).  
[`vikyd/go-checksum`](https://github.com/vikyd/go-checksum#calc-checksum-of-module-directory) does a great job of
explaining what exactly that entails. In essence, the hash you see in a BOM should be the same as in your `go.sum` file,
just in a different format. This is because the CycloneDX specification enforces hashes to be provided in hex encoding,
while Go uses base64 encoded values.

To verify a hash found in a BOM, do the following:

1. Hex decode the value
2. Base64 encode the value
3. Prefix the value with `h1:`
4. Compare with the expected module checksum

#### Example

Given the following `component` element in a BOM:

```xml
<component bom-ref="pkg:golang/github.com/google/uuid@v1.2.0" type="library">
  <name>github.com/google/uuid</name>
  <version>v1.2.0</version>
  <scope>required</scope>
  <hashes>
    <hash alg="SHA-256">
      a8962d5e72515a6a5eee6ff75e5ca1aec2eb11446a1d1336931ce8c57ab2503b
    </hash>
  </hashes>
  <licenses>
    <license>
      <id>BSD-3-Clause</id>
      <url>https://spdx.org/licenses/BSD-3-Clause.html</url>
    </license>
  </licenses>
  <purl>pkg:golang/github.com/google/uuid@v1.2.0</purl>
  <externalReferences>
    <reference type="vcs">
      <url>https://github.com/google/uuid</url>
    </reference>
  </externalReferences>
</component>
```

We take the hash, hex decode it, base64 encode the resulting bytes and prefix that with `h1:` (demonstrated [here](https://gchq.github.io/CyberChef/#recipe=From_Hex('Auto')To_Base64('A-Za-z0-9%2B/%3D')Pad_lines('Start',3,'h1:')&input=YTg5NjJkNWU3MjUxNWE2YTVlZWU2ZmY3NWU1Y2ExYWVjMmViMTE0NDZhMWQxMzM2OTMxY2U4YzU3YWIyNTAzYg) in a CyberChef recipe).

In this case, we end up with `h1:qJYtXnJRWmpe7m/3XlyhrsLrEURqHRM2kxzoxXqyUDs=`.  
In order to verify that this matches what we expect, we can query Go's [checksum database](https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md#checksum-database) for the component we're inspecting:

```
$ curl https://sum.golang.org/lookup/github.com/google/uuid@v1.2.0
2580307
github.com/google/uuid v1.2.0 h1:qJYtXnJRWmpe7m/3XlyhrsLrEURqHRM2kxzoxXqyUDs=
github.com/google/uuid v1.2.0/go.mod h1:TIyPZe4MgqvfeYDBFedMoGGpEw/LqOeaOT+nhxU+yHo=

go.sum database tree
3935567
SapHtgdNCeF00Cx8kqztePV24kgzNg++Xovae42HAMw=

— sum.golang.org Az3grsm7Wm4CVNR1RHq9BFnu9jzcRlU2uw7lr0gfUWgO6+rqPNjT+fUTl9gH0NRTgdwW9nItuQSMbhSaLCsk8YeYSAs=
```

Line 2 of the response tells us that the checksum in our BOM matches that known to the checksum database.

### Version Detection

For the main module and local [replacement modules](https://golang.org/ref/mod#go-mod-file-replace), *cyclonedx-gomod* will perform version detection using Git:

* If the `HEAD` commit is tagged and the tag is a valid [semantic version](https://golang.org/ref/mod#versions), that tag is used.
* If `HEAD` is not tagged, a [pseudo version](https://golang.org/ref/mod#pseudo-versions) is generated.

> Please note that pseudo versions take the previous version into consideration.
> If your repository has been cloned with limited depth, *cyclonedx-gomod* may not be able to see any previous versions.
> For example, [actions/checkout@v2](https://github.com/actions/checkout/tree/v2.3.4#checkout-v2) clones repositories with `fetch-depth: 1` per default.

At the moment, no VCS other than Git is supported. If you need support for another VCS, please open an issue or submit a PR.

## Copyright & License

CycloneDX GoMod is Copyright (c) OWASP Foundation. All Rights Reserved.

Permission to modify and redistribute is granted under the terms of the Apache 2.0 license.  
See the [LICENSE](./LICENSE) file for the full license.

## Contributing

Pull requests are welcome. But please read the
[CycloneDX contributing guidelines](https://github.com/CycloneDX/.github/blob/master/CONTRIBUTING.md) first.

It is generally expected that pull requests will include relevant tests. Tests are automatically run against all
supported Go versions for every pull request.

### Running Tests

Some tests make use of the [CycloneDX CLI](https://github.com/CycloneDX/cyclonedx-cli), e.g. to validate BOMs.  
Make sure to download the CLI binary and make it available as `cyclonedx` in your `$PATH`.  
See also *Setup CycloneDX CLI* in the [workflow](https://github.com/CycloneDX/cyclonedx-gomod/blob/master/.github/workflows/ci.yml).
