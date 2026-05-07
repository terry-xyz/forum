# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0](https://platform.zone01.gr/git/ckotsalas/forum/compare/0.1.0...0.2.0) (2026-05-07)

### Added

- *(database)* add category comment and reaction persistence ([#25](https://platform.zone01.gr/git/ckotsalas/forum/commit/b2fce56646648160630d9f8810b43dae4c22ca11))
- *(handlers)* add category comment and reaction flows ([#26](https://platform.zone01.gr/git/ckotsalas/forum/commit/fe318886667cab1c6c840807667491cd510e21e0))

### Documentation

- *(CHANGELOG.md)* update repository links ([#24](https://platform.zone01.gr/git/ckotsalas/forum/commit/ed3bf3fa10b3b9de4148e71bd83cb5c0f4576cd0))
- *(database)* add explanatory code comments ([#28](https://platform.zone01.gr/git/ckotsalas/forum/commit/9bb24c24eccc4a2abb46899751b8cc12490574e1))

## 0.1.0 (2026-05-04)

### Added

- *(database)* add sqlite schema and user persistence ([#12](https://platform.zone01.gr/git/ckotsalas/forum/commit/ab908c2a07eb1b7c561184a3619b59a5653614bf))
- *(handlers)* add registration and login handlers ([#13](https://platform.zone01.gr/git/ckotsalas/forum/commit/c76dbceb2cacb822e288d1d0878db9e582591366))
- *(database)* add user lookup and post storage ([#19](https://platform.zone01.gr/git/ckotsalas/forum/commit/c4f3290656a5f3bcfd7be6c3817cdb8ff259a95b))
- *(handlers)* wire home and post flows ([#20](https://platform.zone01.gr/git/ckotsalas/forum/commit/5a20c9f65b578652d69726b7b269f97a55808200))

### Fixed

- *(handlers)* return home handler HTTP errors ([#17](https://platform.zone01.gr/git/ckotsalas/forum/commit/9aa8fead70f607bfdfb2c0d5a0e2c700a04f6d2d))
- *(handlers/auth.go)* set session cookie attributes ([#18](https://platform.zone01.gr/git/ckotsalas/forum/commit/b193bca8526ef49eac7a1e2d2c2d3decafab0f47))

### Documentation

- *(forum-instructions)* replace specs with forum instructions ([#10](https://platform.zone01.gr/git/ckotsalas/forum/commit/5fa442218311bec5899245315217332828641455))
- *(CHANGELOG.md)* document forum implementation changes ([#14](https://platform.zone01.gr/git/ckotsalas/forum/commit/82fa2f7ab51c7344dc97379dce00a0833b5a81f8))
- add targeted code comments ([#15](https://platform.zone01.gr/git/ckotsalas/forum/commit/8cc36efa954bab7ac3613f1de806d7d7dccc923b))
- *(CHANGELOG.md)* document code comments ([#16](https://platform.zone01.gr/git/ckotsalas/forum/commit/7686d3e918fd4d602500ce6730180265b3d8087c))
- add notes ([#22](https://platform.zone01.gr/git/ckotsalas/forum/commit/3ddfe7a0adacfd5076606c36b201fd5042ca598e))

### Other

- *(go.mod)* initialize Go module with sqlite driver ([#11](https://platform.zone01.gr/git/ckotsalas/forum/commit/b5d09cc3ff90ecd09e09034bd003f5103109a609))
