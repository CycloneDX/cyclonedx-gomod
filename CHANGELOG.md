# Changelog

## v1.1.0 (unreleased)

### Enhancements

* `app`: Add `-packages` option to include packages in application SBOM ([#85](https://github.com/CycloneDX/cyclonedx-gomod/issues/85) via [#92](https://github.com/CycloneDX/cyclonedx-gomod/pull/92))
* `app`: The `-packages` and `-files` options are now also applied to the standard library component (when `-std` is used) ([#84](https://github.com/CycloneDX/cyclonedx-gomod/issues/84) via [#92](https://github.com/CycloneDX/cyclonedx-gomod/pull/92))

### Breaking Changes

* `app`: `-files` can now only be used in conjunction with `-packages`

### Dependency Updates

* Update `github.com/rs/zerolog` from `v1.25.0` to `v1.26.0`

### Building and Packaging

* Bump `golang` container base images from `1.17.2` to `1.17.3` (via [#95](https://github.com/CycloneDX/cyclonedx-gomod/pull/95))
* Reference container base images by their SHA digest ([#89](https://github.com/CycloneDX/cyclonedx-gomod/issues/89) via [#90](https://github.com/CycloneDX/cyclonedx-gomod/pull/90))
* Introduce multi-platform container image builds ([#87](https://github.com/CycloneDX/cyclonedx-gomod/issues/87) via [#90](https://github.com/CycloneDX/cyclonedx-gomod/pull/90))

## v1.0.0

### Enhancements

* Introduce multi-command CLI ([#42](https://github.com/CycloneDX/cyclonedx-gomod/issues/42) via [#45](https://github.com/CycloneDX/cyclonedx-gomod/pull/45))
* Output SBOMs in v1.3 of the CycloneDX specification ([#43](https://github.com/CycloneDX/cyclonedx-gomod/issues/43) via [`5bab19b`](https://github.com/CycloneDX/cyclonedx-gomod/commit/5bab19bbed9c6de22112ebeb2f71691c4b4163f5))
* Add support for application SBOMs ([#44](https://github.com/CycloneDX/cyclonedx-gomod/issues/44) via [#50](https://github.com/CycloneDX/cyclonedx-gomod/pull/50))
  * Also addresses [#20](https://github.com/CycloneDX/cyclonedx-gomod/issues/20) (thanks [dlorenc](https://github.com/dlorenc) for reporting!)
* Add support for binary SBOMs ([#21](https://github.com/CycloneDX/cyclonedx-gomod/issues/21) via [#46](https://github.com/CycloneDX/cyclonedx-gomod/pull/46))
* Include applicable build constraints in application SBOMs ([#29](https://github.com/CycloneDX/cyclonedx-gomod/issues/29) via [#59](https://github.com/CycloneDX/cyclonedx-gomod/pull/59))
* Add license detection support for binary SBOMs ([#51](https://github.com/CycloneDX/cyclonedx-gomod/issues/51) via [#52](https://github.com/CycloneDX/cyclonedx-gomod/pull/52))
* Generate pseudo versions using `golang.org/x/mod` ([#55](https://github.com/CycloneDX/cyclonedx-gomod/issues/55) via [#57](https://github.com/CycloneDX/cyclonedx-gomod/pull/57))
* Use [license evidence](https://cyclonedx.org/news/cyclonedx-v1.3-released/#copyright-and-license-evidence) for detected licenses ([#40](https://github.com/CycloneDX/cyclonedx-gomod/issues/40) via [#49](https://github.com/CycloneDX/cyclonedx-gomod/pull/49))
* Build with and test against Go 1.17 (via [#54](https://github.com/CycloneDX/cyclonedx-gomod/pull/54))
* Introduce improved logging (via [#46](https://github.com/CycloneDX/cyclonedx-gomod/pull/46))
* Add indication for which application the SBOM was generated for ([#67](https://github.com/CycloneDX/cyclonedx-gomod/issues/67) via [#71](https://github.com/CycloneDX/cyclonedx-gomod/pull/71))
* Slightly reduce threshold for license detection confidence, and log a debug message if this threshold isn't met ([#79](https://github.com/CycloneDX/cyclonedx-gomod/issues/79) via [#80](https://github.com/CycloneDX/cyclonedx-gomod/pull/80))
  * Thanks [TheDiveO](https://github.com/TheDiveO) for reporting!

### Fixes

* Fix annotated tags not being recognized as versions ([#56](https://github.com/CycloneDX/cyclonedx-gomod/issues/56) via [#57](https://github.com/CycloneDX/cyclonedx-gomod/pull/57))
* Fix normalized versions interfering with hash calculation ([#58](https://github.com/CycloneDX/cyclonedx-gomod/issues/58) via [#60](https://github.com/CycloneDX/cyclonedx-gomod/pull/60))
* Fix `app` command missing dependencies when `main` package is spread across multiple files ([#75](https://github.com/CycloneDX/cyclonedx-gomod/issues/75) via [#78](https://github.com/CycloneDX/cyclonedx-gomod/pull/78))
  * Also addresses [#76](https://github.com/CycloneDX/cyclonedx-gomod/issues/76) (thanks [TheDiveO](https://github.com/TheDiveO) for reporting!) 

### Breaking Changes

* The CLI now consists of multiple subcommands, thus being incompatible with the CLI in cyclonedx-gomod `v0.x`
* Detected licenses (when using the `-licenses` flag) will now use the `components/evidence/licenses` node instead of `components/licenses`. Tools that consume SBOMs and don't support CycloneDX v1.3 yet may not recognize those licenses
* Version normalization has been removed ([#60](https://github.com/CycloneDX/cyclonedx-gomod/pull/60)). As a consequence, `+incompatible` suffixes and `v` prefixes (`-novprefix` flag in `v0.x`) are not trimmed anymore
* The `-reproducible` flag has been removed (via [`9b45f4a`](https://github.com/CycloneDX/cyclonedx-gomod/commit/9b45f4a0e905dc89bef1d238c28de908bd4163a0))

### Dependency Updates

* Update `github.com/CycloneDX/cyclonedx-go` from `v0.3.0` to `v0.4.0` (via [`5bab19b`](https://github.com/CycloneDX/cyclonedx-gomod/commit/5bab19bbed9c6de22112ebeb2f71691c4b4163f5))
* Update `golang.org/x/mod` from `v0.4.2` to `v0.5.1` (via [#57](https://github.com/CycloneDX/cyclonedx-gomod/pull/57) and [`088f0e3`](https://github.com/CycloneDX/cyclonedx-gomod/commit/088f0e30e6aa80a37f767651877cf943563960a4))
* Update `golang.org/x/crypto` from `v0.0.0-20210711020723-a769d52b0f97` to `v0.0.0-20210817164053-32db794688a5` (via [`75ae52a`](https://github.com/CycloneDX/cyclonedx-gomod/commit/75ae52ac039d9d702a1861c9625d0a14116097ce))

### Building and Packaging

* Produce and publish an SBOM for each binary built when releasing (via [#62](https://github.com/CycloneDX/cyclonedx-gomod/pull/62))
* Builds for `windows/386` and `linux/386` have been dropped (via [#62](https://github.com/CycloneDX/cyclonedx-gomod/pull/62))
* Use standard Go notation for architectures in release artifact names (via [#62](https://github.com/CycloneDX/cyclonedx-gomod/pull/62))
  * e.g. `cyclonedx-gomod_1.0.0_windows_x64.zip` is now `cyclonedx-gomod_1.0.0_windows_amd64.zip`
