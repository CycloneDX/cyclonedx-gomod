# Changelog

## v1.0.0

### Enhancements

* Introduce multi-command CLI ([#42](https://github.com/CycloneDX/cyclonedx-gomod/issues/42) via [#45](https://github.com/CycloneDX/cyclonedx-gomod/pull/45))
* Output SBOMs in v1.3 of the CycloneDX specification ([#43](https://github.com/CycloneDX/cyclonedx-gomod/issues/43) via [`5bab19b`](https://github.com/CycloneDX/cyclonedx-gomod/commit/5bab19bbed9c6de22112ebeb2f71691c4b4163f5))
* Add support for application SBOMs ([#44](https://github.com/CycloneDX/cyclonedx-gomod/issues/44) via [#50](https://github.com/CycloneDX/cyclonedx-gomod/pull/50))
* Add support for binary SBOMs ([#21](https://github.com/CycloneDX/cyclonedx-gomod/issues/21) via [#46](https://github.com/CycloneDX/cyclonedx-gomod/pull/46))
* Include applicable build constraints in application SBOMs ([#29](https://github.com/CycloneDX/cyclonedx-gomod/issues/29) via [#59](https://github.com/CycloneDX/cyclonedx-gomod/pull/59))
* Add license detection support for binary SBOMs ([#51](https://github.com/CycloneDX/cyclonedx-gomod/issues/51) via [#52](https://github.com/CycloneDX/cyclonedx-gomod/pull/52))
* Generate pseudo versions using `golang.org/x/mod` ([#55](https://github.com/CycloneDX/cyclonedx-gomod/issues/55) via [#57](https://github.com/CycloneDX/cyclonedx-gomod/pull/57))
* Use [license evidence](https://cyclonedx.org/news/cyclonedx-v1.3-released/#copyright-and-license-evidence) for detected licenses ([#40](https://github.com/CycloneDX/cyclonedx-gomod/issues/40) via [#49](https://github.com/CycloneDX/cyclonedx-gomod/pull/49))
* Build with and test against Go 1.17 (via [#54](https://github.com/CycloneDX/cyclonedx-gomod/pull/54))
* Produce and publish an SBOM for each binary built when releasing (via [#62](https://github.com/CycloneDX/cyclonedx-gomod/pull/62))

### Fixes

* Fix annotated tags not being recognized as versions ([#56](https://github.com/CycloneDX/cyclonedx-gomod/issues/56) via [#57](https://github.com/CycloneDX/cyclonedx-gomod/pull/57))
* Fix normalized versions interfering with hash calculation ([#58](https://github.com/CycloneDX/cyclonedx-gomod/issues/58) via [#60](https://github.com/CycloneDX/cyclonedx-gomod/pull/60))

### Breaking Changes

* Detected licenses (when using the `-licenses` flag) will now use the `components/evidence/licenses` node instead of `components/licenses`. Tools that consume SBOMs and don't support CycloneDX v1.3 yet may not recognize those licenses. 
* Version normalization has been removed ([#60](https://github.com/CycloneDX/cyclonedx-gomod/pull/60)). As a consequence, `+incompatible` suffixes and `v` prefixes (`-noprefix` flag in cyclonedx-gomod v0.x) are not trimmed anymore.

### Dependency Updates

* Update `github.com/CycloneDX/cyclonedx-go` from `v0.3.0` to `v0.4.0` (via [`5bab19b`](https://github.com/CycloneDX/cyclonedx-gomod/commit/5bab19bbed9c6de22112ebeb2f71691c4b4163f5))
* Update `golang.org/x/mod` from `v0.4.2` to `v0.5.0` (via [#57](https://github.com/CycloneDX/cyclonedx-gomod/pull/57))
* Update `golang.org/x/crypto` from `v0.0.0-20210711020723-a769d52b0f97` to `v0.0.0-20210817164053-32db794688a5` (via [`75ae52a`](https://github.com/CycloneDX/cyclonedx-gomod/commit/75ae52ac039d9d702a1861c9625d0a14116097ce))
