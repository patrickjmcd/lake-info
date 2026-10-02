# Changelog

## [1.7.1](https://github.com/patrickjmcd/lake-info/compare/v1.7.0...v1.7.1) (2026-09-17)


### Bug Fixes

* make Google Sheets export best-effort, store to Mongo first ([ae42df2](https://github.com/patrickjmcd/lake-info/commit/ae42df224f313072d536d00eed2679de507c60ed))
* skip duplicate measurements when storing lake info ([95a968a](https://github.com/patrickjmcd/lake-info/commit/95a968aebcb940a6b459d0b4a6077ebe3bb56627))

## [1.7.0](https://github.com/patrickjmcd/lake-info/compare/v1.6.3...v1.7.0) (2026-09-17)


### Features

* use CWMS Data API for table rock scrape, HTML fallback ([#25](https://github.com/patrickjmcd/lake-info/issues/25)) ([316b5be](https://github.com/patrickjmcd/lake-info/commit/316b5be162f0c98e700036cb63e643fe4a02369c))

## [1.6.3](https://github.com/patrickjmcd/lake-info/compare/v1.6.2...v1.6.3) (2026-09-17)


### Bug Fixes

* harden table rock scrape/store path ([3be9789](https://github.com/patrickjmcd/lake-info/commit/3be978921eba401fbbe4c0707ebab86e3a80d348))
* stop dropping the oldest data row in table rock scrape ([4f2125a](https://github.com/patrickjmcd/lake-info/commit/4f2125a2f36e58d131b90982bf4a20c56ecf704f))

## [1.6.2](https://github.com/patrickjmcd/lake-info/compare/v1.6.1...v1.6.2) (2026-01-18)


### Bug Fixes

* use custom HTTP client with insecure TLS for fetching records ([e1df5cd](https://github.com/patrickjmcd/lake-info/commit/e1df5cdcece011b775fd09d133588acc575ba482))

## [1.6.1](https://github.com/patrickjmcd/lake-info/compare/v1.6.0...v1.6.1) (2024-09-19)


### Bug Fixes

* uses gsheets for reading and writing ([19e0037](https://github.com/patrickjmcd/lake-info/commit/19e003786df00579476382837a94b38c35ee6b16))

## [1.6.0](https://github.com/patrickjmcd/lake-info/compare/v1.5.0...v1.6.0) (2024-09-19)


### Features

* store in google sheet too ([19ee439](https://github.com/patrickjmcd/lake-info/commit/19ee439b4f08c930ac37ceee7621ebbe31df681c))

## [1.5.0](https://github.com/patrickjmcd/lake-info/compare/v1.4.0...v1.5.0) (2024-05-27)


### Features

* consolidate logging and add dry-run option ([ea8a926](https://github.com/patrickjmcd/lake-info/commit/ea8a926a3a18705cfdf6b412b11d46bdfe11a302))

## [1.4.0](https://github.com/patrickjmcd/lake-info/compare/v1.3.1...v1.4.0) (2023-12-19)


### Features

* reorganize for creating an executable ([47a948c](https://github.com/patrickjmcd/lake-info/commit/47a948cb6933da151b2f41b79c92d8834f4fb109))

## [1.3.1](https://github.com/patrickjmcd/lake-info/compare/v1.3.0...v1.3.1) (2023-12-18)


### Bug Fixes

* use PAT secret ([113a991](https://github.com/patrickjmcd/lake-info/commit/113a991acf7c1dad946e7e0285301160ac57341a))

## [1.3.0](https://github.com/patrickjmcd/lake-info/compare/v1.2.0...v1.3.0) (2023-12-18)


### Features

* switch to using go and connect framework with MongoDB ([134d423](https://github.com/patrickjmcd/lake-info/commit/134d4235a40ecaa39c20f47d26d5fd0789d47bff))


### Bug Fixes

* add login step to container registry ([408a455](https://github.com/patrickjmcd/lake-info/commit/408a4551262e8601eece3b8f2bb58ba49f7feb93))
* add permissions to builder action ([aad2ea7](https://github.com/patrickjmcd/lake-info/commit/aad2ea792922d547ee21c15de4117e16dec3b39a))
