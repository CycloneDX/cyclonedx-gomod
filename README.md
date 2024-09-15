# cyclonedx-gomod

[![Build Status](https://github.com/CycloneDX/cyclonedx-gomod/actions/workflows/ci.yml/badge.svg)](https://github.com/CycloneDX/cyclonedx-gomod/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/CycloneDX/cyclonedx-gomod)](https://goreportcard.com/report/github.com/CycloneDX/cyclonedx-gomod)
[![Go Reference](https://pkg.go.dev/badge/github.com/CycloneDX/cyclonedx-gomod.svg)](https://pkg.go.dev/github.com/CycloneDX/cyclonedx-gomod)
[![Latest GitHub release](https://img.shields.io/github/v/release/CycloneDX/cyclonedx-gomod?sort=semver)](https://github.com/CycloneDX/cyclonedx-gomod/releases/latest)
[![License](https://img.shields.io/badge/license-Apache%202.0-brightgreen.svg)](LICENSE)  
[![Website](https://img.shields.io/badge/https://-cyclonedx.org-blue.svg)](https://cyclonedx.org/)
[![Slack Invite](https://img.shields.io/badge/Slack-Join-blue?logo=slack&labelColor=393939)](https://cyclonedx.org/slack/invite)
[![Group Discussion](https://img.shields.io/badge/discussion-groups.io-blue.svg)](https://groups.io/g/CycloneDX)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social&label=Follow)](https://twitter.com/CycloneDX_Spec)

*cyclonedx-gomod* creates CycloneDX Software Bill of Materials (SBOM) from Go modules

## Installation

Prebuilt binaries are available on the [releases](https://github.com/CycloneDX/cyclonedx-gomod/releases) page.

### Homebrew

```shell
brew install cyclonedx/cyclonedx/cyclonedx-gomod
```

### From Source

```shell
go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest
```

Building from source requires Go 1.23.1 or newer.

## Compatibility

*cyclonedx-gomod* aims to produce SBOMs according to the latest CycloneDX specification, and currently supports up to [1.6](https://cyclonedx.org/docs/1.6/). 
You can use the [CycloneDX CLI](https://github.com/CycloneDX/cyclonedx-cli#convert-command) to convert between multiple BOM formats or specification versions. 

## Usage

```
USAGE
  cyclonedx-gomod <SUBCOMMAND> [FLAGS...] [<ARG>...]

cyclonedx-gomod creates CycloneDX Software Bill of Materials (SBOM) from Go modules.

Multiple subcommands are offered, each targeting different use cases:

- SBOMs generated with "app" include only those modules that the target application
  actually depends on. Modules required by tests or packages that are not imported
  by the application are not included. Build constraints are evaluated, which enables
  a very detailed view of what's really compiled into an application's binary.
  
- SBOMs generated with "mod" include the aggregate of modules required by all 
  packages in the target module. This optionally includes modules required by
  tests and test packages. Build constraints are NOT evaluated, allowing for 
  a "whole picture" view on the target module's dependencies.

- "bin" offers support for generating rudimentary SBOMs from binaries built with Go modules.

Distributors of applications will typically use "app" and provide the resulting SBOMs
alongside their application's binaries. This enables users to only consume SBOMs for
artifacts that they actually use. For example, a Go module may include "server" and
"client" applications, of which only the "client" is distributed to users. 
Additionally, modules included in "client" may differ, depending on which platform 
it was compiled for.

Vendors or maintainers may choose to use "mod" for internal use, where it's too
cumbersome to deal with many SBOMs for the same product. Possible use cases are: 
- Tracking of component inventory
- Tracking of third party component licenses
- Continuous monitoring for vulnerabilities
"mod" may also be used to generate SBOMs for libraries.

SUBCOMMANDS
  app      Generate SBOMs for applications
  bin      Generate SBOMs for binaries
  mod      Generate SBOMs for modules
  version  Show version information
```

### Subcommands

#### `app`

```
USAGE
  cyclonedx-gomod app [FLAGS...] [MODULE_PATH]

Generate SBOMs for applications.

In order to produce accurate SBOMs, build constraints must be configured
via environment variables. These build constraints should mimic the ones passed
to the "go build" command for the application.

Environment variables that act as build constraints are:
  - GOARCH       The target architecture (386, amd64, etc.)
  - GOOS         The target operating system (linux, windows, etc.)
  - CGO_ENABLED  Whether or not CGO is enabled
  - GOFLAGS      Flags that are passed to the Go command (e.g. build tags)

A complete overview of all environment variables can be found here:
  https://pkg.go.dev/cmd/go#hdr-Environment_variables

Applicable build constraints are included as properties of the main component.

Because build constraints influence Go's module selection, an SBOM should be generated
for each target in the build matrix.

The -main flag should be used to specify the path to the application's main package.
It must point to a directory within MODULE_PATH. If not set, MODULE_PATH is assumed.

In order to not only include modules, but also the packages within them,
the -packages flag can be used. Packages are represented as subcomponents of modules.

By passing -files, all files that would be included in a binary will be attached
as subcomponents of their respective package. File versions follow the v0.0.0-SHORTHASH pattern,
where SHORTHASH is the first 12 characters of the file's SHA1 hash.
Because files are subcomponents of packages, -files can only be used in conjunction with -packages.
When -paths option is additionally enabled, each file would have a property with
a file path relative to its module root.

Licenses detected via -licenses flag will, per default, be reported as evidence.
This is because it can not be guaranteed that the detected licenses are in fact correct.
In case analysis software ingesting the BOM generated by this tool can not yet handle
evidences, detected licenses may be asserted using the -assert-licenses flag.
For documentation on the respective fields of the CycloneDX specification, refer to:
  * https://cyclonedx.org/docs/1.4/json/#components_items_licenses
  * https://cyclonedx.org/docs/1.4/json/#components_items_evidence_licenses

Examples:
  $ GOARCH=arm64 GOOS=linux GOFLAGS="-tags=foo,bar" cyclonedx-gomod app -output linux-arm64.bom.xml
  $ cyclonedx-gomod app -json -output acme-app.bom.json -packages -files -licenses -main cmd/acme-app /usr/src/acme-module

FLAGS
  -assert-licenses=false  Assert detected licenses
  -files=false            Include files
  -json=false             Output in JSON
  -licenses=false         Perform license detection
  -main string            Path to the application's main package, relative to MODULE_PATH
  -noserial=false         Omit serial number
  -output -               Output file path (or - for STDOUT)
  -output-version 1.6     Output spec verson (1.6, 1.5, 1.4, 1.3, 1.2, 1.1, 1.0)
  -packages=false         Include packages
  -paths=false            Include file paths relative to their module root
  -serial string          Serial number
  -std=false              Include Go standard library as component and dependency of the module
  -verbose=false          Enable verbose output
```

#### `bin`

```
USAGE
  cyclonedx-gomod bin [FLAGS...] BINARY_PATH

Generate SBOMs for binaries.

Although the binary is never executed by cyclonedx-gomod, it must be executable.
This is a requirement by the "go version -m" command that is used to provide this functionality.

When license detection is enabled, all modules (including the main module)
will be downloaded to the module cache using "go mod download".
For the download of the main module to work, its version has to be provided
via the -version flag.

Licenses detected via -licenses flag will, per default, be reported as evidence.
This is because it can not be guaranteed that the detected licenses are in fact correct.
In case analysis software ingesting the BOM generated by this tool can not yet handle
evidences, detected licenses may be asserted using the -assert-licenses flag.
For documentation on the respective fields of the CycloneDX specification, refer to:
  * https://cyclonedx.org/docs/1.4/json/#components_items_licenses
  * https://cyclonedx.org/docs/1.4/json/#components_items_evidence_licenses

Please note that data embedded in binaries shouldn't be trusted,
unless there's solid evidence that the binaries haven't been modified
since they've been built.

Example:
  $ cyclonedx-gomod bin -json -output acme-app-v1.0.0.bom.json -version v1.0.0 ./acme-app

FLAGS
  -assert-licenses=false  Assert detected licenses
  -json=false             Output in JSON
  -licenses=false         Perform license detection
  -noserial=false         Omit serial number
  -output -               Output file path (or - for STDOUT)
  -output-version 1.6     Output spec verson (1.6, 1.5, 1.4, 1.3, 1.2, 1.1, 1.0)
  -serial string          Serial number
  -std=false              Include Go standard library as component and dependency of the module
  -verbose=false          Enable verbose output
  -version string         Version of the main component
```

#### `mod`

```
USAGE
  cyclonedx-gomod mod [FLAGS...] [MODULE_PATH]

Generate SBOMs for modules.

Licenses detected via -licenses flag will, per default, be reported as evidence.
This is because it can not be guaranteed that the detected licenses are in fact correct.
In case analysis software ingesting the BOM generated by this tool can not yet handle
evidences, detected licenses may be asserted using the -assert-licenses flag.
For documentation on the respective fields of the CycloneDX specification, refer to:
  * https://cyclonedx.org/docs/1.4/json/#components_items_licenses
  * https://cyclonedx.org/docs/1.4/json/#components_items_evidence_licenses

Examples:
  $ cyclonedx-gomod mod -licenses -type library -json -output bom.json ./cyclonedx-go
  $ cyclonedx-gomod mod -test -output bom.xml ./cyclonedx-go

FLAGS
  -assert-licenses=false  Assert detected licenses
  -json=false             Output in JSON
  -licenses=false         Perform license detection
  -noserial=false         Omit serial number
  -output -               Output file path (or - for STDOUT)
  -output-version 1.6     Output spec verson (1.6, 1.5, 1.4, 1.3, 1.2, 1.1, 1.0)
  -serial string          Serial number
  -std=false              Include Go standard library as component and dependency of the module
  -test=false             Include test dependencies
  -type application       Type of the main component
  -verbose=false          Enable verbose output
```

### Examples 📃

In order to demonstrate what SBOMs generated with *cyclonedx-gomod* look like, 
as well as to give you an idea about the differences between the commands `app`, 
`mod` and `bin`, we provide example SBOMs for each command in the [`examples`](./examples) directory.

The whole process of generating these examples is encapsulated in [`Dockerfile.examples`](./Dockerfile.examples).  
To generate them yourself, simply execute the following command:

```shell
$ make examples
```

### GitHub Actions 🤖

We made a GitHub Action to help integrate *cyclonedx-gomod* into existing CI/CD workflows!  
You can find it on the GitHub marketplace: [*gh-gomod-generate-sbom*](https://github.com/marketplace/actions/cyclonedx-gomod-generate-sbom)

### GoReleaser 🚀

The recommended way of integrating with [GoReleaser](https://goreleaser.com/) is via its [*sbom* feature](https://goreleaser.com/customization/sbom/).
You can find some example configurations for each *cyclonedx-gomod* command below, given the following [`builds`](https://goreleaser.com/customization/build/):

```yaml
builds:
- env:
  - CGO_ENABLED=0
  goos:
  - linux
  - windows
  - darwin
  goarch:
  - amd64
  - arm64
  tags:
  - foo
  - bar
```

```yaml
# app command:
# - generate a SBOM for each binary built
# - provide build context via environment variables

sboms:
- documents:
  - "{{ .Binary }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}.bom.json"
  artifacts: binary
  cmd: cyclonedx-gomod
  args: ["app", "-licenses", "-json", "-output", "$document", "../"]
  env:
  - GOARCH={{ .Arch }}
  - GOOS={{ .Os }}
  - GOFLAGS=-tags=foo,bar
```

```yaml
# bin command:
# - generate a SBOM for each binary built

sboms:
- documents:
  - "{{ .Binary }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}.bom.json"
  artifacts: binary
  cmd: cyclonedx-gomod
  args: ["bin", "-json", "-output", "$document", "$artifact"]
```

```yaml
# mod command:
# - generate a single SBOM for the entire module

sboms:
- documents:
  - bom.json
  artifacts: any
  cmd: cyclonedx-gomod
  args: [ "mod", "-licenses", "-std", "-json", "-output", "$document", "../" ]
```

GoReleaser will execute `cmd`s in its `dist` directory, which is a subdirectory of the project root. 
Because `app` and `mod` both expect the module's root directory as an argument, `../` must be provided.

### Docker 🐳

```shell
$ docker run -it --rm \
    -v "/path/to/mymodule:/usr/src/mymodule" \
    -v "$(pwd):/out" \
    cyclonedx/cyclonedx-gomod:v1 mod -json -output /out/bom.json /usr/src/mymodule
```

> The image is based on `golang:1.18-alpine`.  
> When using the `app` command, please keep in mind that the Go version may influence module selection.  
> We generally recommend using a [precompiled binary](https://github.com/CycloneDX/cyclonedx-gomod/releases) 
> and running it in the same environment in which you're building your application in.

### Library Usage

Starting with `v1.2.0`, *cyclonedx-gomod* can be used as a library as well:

```shell
go get -v github.com/CycloneDX/cyclonedx-gomod
```

Refer to the [documentation](https://pkg.go.dev/github.com/CycloneDX/cyclonedx-gomod) for details and examples.

> Be warned that *cyclonedx-gomod* is and will continue to be primarily a CLI tool.  
> While we'll only introduce breaking changes to the exposed APIs in accordance with semver,
> we will not invest in supporting older versions. If you intend on depending on our API,
> please assess if you'll be able to keep up. For example, we will move to the newest Go version
> shortly after its GA, and will almost definitely use backwards-incompatible features going forward.

## Important Notes

### Vendoring

Modules that use [vendoring](https://golang.org/ref/mod#go-mod-vendor) are, although in a limited manner, supported.  
Limitations are as follows:

* **No hashes.** Go doesn't copy all module files to the `vendor`, only those that are required to build
  and test the main module. Because [module checksums](#hashes) consider almost all files in a module's directory though, 
  calculating accurate hashes from the `vendor` directory is not possible. As a consequence, SBOMs for modules that use
  vendoring do not include component hashes.
* **License detection may fail.** Go doesn't always copy license files when vendoring modules, which may cause license detection to fail.

### Licenses

There is currently no standard way for developers to declare their module's license.  
Detecting licenses based on files in a repository is a non-trivial task, which is why *cyclonedx-gomod*  
uses [`go-license-detector`](https://github.com/go-enry/go-license-detector) to resolve module licenses.

While `go-license-detector`'s license matching *may* be accurate most of the time, SBOMs should state facts.  
This is why detected licenses are included as [evidences](https://cyclonedx.org/news/cyclonedx-v1.3-released/#copyright-and-license-evidence), 
rather than the `licenses` field directly.

> Detected licenses may be *asserted* using the `-assert-licenses` flag. When provided,
> *cyclonedx-gomod* will use the `licenses` field, instead of `evidences`. This can be
> helpful when the generated BOM is pushed to an analysis tool that does not yet handle
> evidences.

### Hashes

*cyclonedx-gomod* uses the same hashing algorithm Go uses for its [module authentication](https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md#module-authentication-with).  
[`vikyd/go-checksum`](https://github.com/vikyd/go-checksum#calc-checksum-of-module-directory) does a great job of
explaining what exactly that entails. In essence, the hash you see in an SBOM should be the same as in your `go.sum` file,
just in a different format. This is because the CycloneDX specification enforces hashes to be provided in hex encoding,
while Go uses base64 encoded values.

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

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/CycloneDX/cyclonedx-gomod)

Pull requests are welcome. But please read the
[CycloneDX contributing guidelines](https://github.com/CycloneDX/.github/blob/master/CONTRIBUTING.md) first.

It is generally expected that pull requests will include relevant tests. Tests are automatically run against all
supported Go versions for every pull request.
