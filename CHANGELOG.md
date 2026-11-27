# Changelog

## [0.8.0](https://github.com/7K-Inari/inari-api/compare/v0.7.0...v0.8.0) (2026-11-27)


### Features

* add deployment observability contracts ([f224904](https://github.com/7K-Inari/inari-api/commit/f2249048ba73bcf43508f65d79bc52ca0f922c31))
* add deployment observability contracts ([a35cb86](https://github.com/7K-Inari/inari-api/commit/a35cb86e3c5ef8393bb812046c814ecdcfccbf4a))
* add generic provisioning contract ([42e1b48](https://github.com/7K-Inari/inari-api/commit/42e1b48da8bfbdd0e511050b5b41d8aefbd2dc86))
* add generic provisioning contract ([976ac4d](https://github.com/7K-Inari/inari-api/commit/976ac4df7d73d983b1b53ff6406172332181c572))
* add org resource target REST contracts ([6779924](https://github.com/7K-Inari/inari-api/commit/6779924724c607c108a4a9487d22d3d15077c6d8))
* add org resource target REST contracts ([7af04a0](https://github.com/7K-Inari/inari-api/commit/7af04a00890b879733b86342aee570798cffe392))
* add preflight application schemas ([cdab03f](https://github.com/7K-Inari/inari-api/commit/cdab03fec6436e6b7c19e3ebf3d144f835cbaf4d))
* add verification trust and agent contracts ([9141c0a](https://github.com/7K-Inari/inari-api/commit/9141c0a793b211cf8eb94cd5fe56779df1760ec4))
* **agent:** add verification report stream contract ([14c302a](https://github.com/7K-Inari/inari-api/commit/14c302a03a71d11a752bfb60eb0ca9cc21de3a55))
* expose application execution results on plans ([dc6ef4d](https://github.com/7K-Inari/inari-api/commit/dc6ef4d8c2984a3bf107309153bac0f18c7cefe6))
* expose deploy artifact digest in REST clients ([6769cdc](https://github.com/7K-Inari/inari-api/commit/6769cdc888d9dc1d384bec310be60c437dad359d))
* **extension:** add versioned contract declarations ([dcd3915](https://github.com/7K-Inari/inari-api/commit/dcd3915e1c7b7328775345e9ae5d653452662b69))
* **extensions:** add versioned plugin contracts ([b5e7e0b](https://github.com/7K-Inari/inari-api/commit/b5e7e0bc36e96b9164e4832b8087f44add06797d))
* **extensions:** generate contract metadata clients ([970e458](https://github.com/7K-Inari/inari-api/commit/970e45820471e19094e27c763412c3f8e6b8aa1f))
* **openapi:** add promotion and run link contracts ([03da85e](https://github.com/7K-Inari/inari-api/commit/03da85e5b3e004577b040dbac4bd86a6dd863c74))
* **openapi:** publish verification policy contracts ([3464a83](https://github.com/7K-Inari/inari-api/commit/3464a8388f5a96827623c19c6dcc6f1bf9b4e57b))
* publish deployment plan REST contracts ([6c94552](https://github.com/7K-Inari/inari-api/commit/6c94552d4b42fba62f5c2fe4b2e01b6a2e99d168))
* **trustedsource:** enforce signed runtime evidence eligibility ([963d871](https://github.com/7K-Inari/inari-api/commit/963d871e5090694e4a1664864b15c43aebc3b395))
* **verification:** add deterministic run identity contract ([d23453b](https://github.com/7K-Inari/inari-api/commit/d23453ba6266ed07b49ab0484853f67b2153da16))


### Bug Fixes

* align observability verification profiles ([34d8e0d](https://github.com/7K-Inari/inari-api/commit/34d8e0db977eb29ac04da8f3eece5224dbcb64a7))
* align observability verification with generic profiles ([8d5dfaa](https://github.com/7K-Inari/inari-api/commit/8d5dfaa6e2adfe5592a991cf296d8d53aa91ab87))
* disambiguate observability capability schema ([92fac9d](https://github.com/7K-Inari/inari-api/commit/92fac9d4765700aafb5382194f756d9a8b139264))
* **extension:** accept npm sdk ranges ([5fe7476](https://github.com/7K-Inari/inari-api/commit/5fe74760b5fc3150355693bb1fa39e61cb978d36))
* **openapi:** disambiguate verification grant status schema ([a99c67f](https://github.com/7K-Inari/inari-api/commit/a99c67f18be8f523b484a038b0eaa09b3d122eaf))
* **openapi:** stabilize embedded contract generation ([34169b9](https://github.com/7K-Inari/inari-api/commit/34169b962354c7b760fb42b160f9e0bf0cfe767e))
* restrict unsigned sources to development ([fbb8a5f](https://github.com/7K-Inari/inari-api/commit/fbb8a5fd4de18e702f338fa586a564b71cf5b6d9))
* restrict unsigned trusted sources to development ([5ee597d](https://github.com/7K-Inari/inari-api/commit/5ee597d2eb0107a29e2076a414e840a823554df0))
* tighten verification result schema ([467dc36](https://github.com/7K-Inari/inari-api/commit/467dc36a77f961dc04d2598c522cb5f09519082f))

## [0.7.0](https://github.com/7K-Inari/inari-api/compare/v0.6.0...v0.7.0) (2026-11-14)


### ⚠ BREAKING CHANGES

* migrate generated clients to connect-go v2 and protobuf-es v2 ([#59](https://github.com/7K-Inari/inari-api/issues/59))
* add inari.tunnel.v2 with multi-valued headers

### Features

* add inari.tunnel.v1 contract for kubectl gateway tunnel ([#34](https://github.com/7K-Inari/inari-api/issues/34)) ([25e482f](https://github.com/7K-Inari/inari-api/commit/25e482f0464f3ed70c61c893509aaeae76c6624f))
* add inari.tunnel.v2 with multi-valued headers ([c10b708](https://github.com/7K-Inari/inari-api/commit/c10b708d8a111546089a6aa25ac0d8315169fded))
* add trusted source and admission contracts ([#83](https://github.com/7K-Inari/inari-api/issues/83)) ([2174844](https://github.com/7K-Inari/inari-api/commit/21748447614b6c988dae71cb253bf3dab2d4287f))
* **agent/v1:** add baseline_repo to RegisterClusterResponse and GitopsConfigUpdate command ([0464651](https://github.com/7K-Inari/inari-api/commit/0464651127cd6d4b6ed31d058ab040c9cc989ed5))
* **extensions/v1:** add AgentGateway contract ([#82](https://github.com/7K-Inari/inari-api/issues/82)) ([9ae9821](https://github.com/7K-Inari/inari-api/commit/9ae982121e013c32b775b75067137b7ac3b328c7))
* migrate generated clients to connect-go v2 and protobuf-es v2 ([#59](https://github.com/7K-Inari/inari-api/issues/59)) ([180bdb0](https://github.com/7K-Inari/inari-api/commit/180bdb0b4fed3a13346c3e2202a1012ba3c02182))
* **release:** per-merge edge releases (vX.Y.Z-&lt;sha&gt; + moving edge channel) ([#22](https://github.com/7K-Inari/inari-api/issues/22)) ([e384273](https://github.com/7K-Inari/inari-api/commit/e384273d2619c3e81fdc737ad4848a788037f63b))
* shared cluster-ID charset testdata ([6735649](https://github.com/7K-Inari/inari-api/commit/6735649947dc2c40be756fe7bbc411a1b266d3a4))
* Wave 1 contracts for extension OIDC pass-through and per-user Git credential redemption ([#20](https://github.com/7K-Inari/inari-api/issues/20)) ([52ea583](https://github.com/7K-Inari/inari-api/commit/52ea5837ecacf8cab48b8bd26ead0301ce4c65b5))

## [0.6.0](https://github.com/7K-Inari/inari-api/compare/v0.5.2...v0.6.0) (2026-09-14)


### Features

* **agent/v1:** add SecretStoreApply/SecretStoreDelete commands and event types ([#18](https://github.com/7K-Inari/inari-api/issues/18)) ([b01da6a](https://github.com/7K-Inari/inari-api/commit/b01da6a8ca27b94b7219ed5f5b62fafbf7fa6406))

## [0.5.2](https://github.com/7K-Inari/inari-api/compare/v0.5.1...v0.5.2) (2026-08-21)


### Bug Fixes

* add repository field for npm provenance validation ([#15](https://github.com/7K-Inari/inari-api/issues/15)) ([6576a47](https://github.com/7K-Inari/inari-api/commit/6576a47fab8b0c706b876d892bd8786e9d4760be))

## [0.5.1](https://github.com/7K-Inari/inari-api/compare/v0.5.0...v0.5.1) (2026-08-21)


### Bug Fixes

* **ci:** pin npm ^11 for trusted publishing ([#13](https://github.com/7K-Inari/inari-api/issues/13)) ([f3ffff8](https://github.com/7K-Inari/inari-api/commit/f3ffff8163793f8c2bc18b83fc41917cf7ed671e))

## [0.5.0](https://github.com/7K-Inari/inari-api/compare/v0.4.0...v0.5.0) (2026-08-21)


### Features

* **agent:** real EventStream protocol with registration, events, and commands ([44002b5](https://github.com/7K-Inari/inari-api/commit/44002b549ee01583171a038e8dd0a4e9b579c387))
* **agent:** real EventStream protocol with registration, events, and commands ([5b154ed](https://github.com/7K-Inari/inari-api/commit/5b154edc7a141f6c82085928ed9a99d62334759a))
* **agent:** regenerate Go/TS clients, add contract tests, bump 0.2.0 ([5e883a3](https://github.com/7K-Inari/inari-api/commit/5e883a366574557a53291f85c02154e5f3e3b095))
* bootstrap contract repo (buf protos, OpenAPI skeleton, codegen, CI) ([5bee67f](https://github.com/7K-Inari/inari-api/commit/5bee67f93aadf2300c9557f086775e164e7b1101))
* full REST surface in OpenAPI contract with typed Go/TS clients ([#6](https://github.com/7K-Inari/inari-api/issues/6)) ([76d69f9](https://github.com/7K-Inari/inari-api/commit/76d69f96bebeff944f22c59370ca32e0b7506219))
* **openapi:** add REST skeleton spec with Go/TS codegen configs ([ce7616a](https://github.com/7K-Inari/inari-api/commit/ce7616aa7a32266011811f2a2dc08bd9391968ee))
* **plugin/v1:** add auth context, capabilities, invoke envelope, health and gRPC stubs ([#4](https://github.com/7K-Inari/inari-api/issues/4)) ([12b34d0](https://github.com/7K-Inari/inari-api/commit/12b34d04d2cafa3c179375d2103087f9b1ac1921))
* **proto:** add buf v2 setup with placeholder agent/plugin contracts ([a17e970](https://github.com/7K-Inari/inari-api/commit/a17e970af2c842b9df6d5bbb78a8485655fb8ba2))

## [0.4.0](https://github.com/7K-Inari/inari-api/compare/v0.3.0...v0.4.0) (2026-08-21)


### Features

* full REST surface in OpenAPI contract with typed Go/TS clients ([#6](https://github.com/7K-Inari/inari-api/issues/6)) ([76d69f9](https://github.com/7K-Inari/inari-api/commit/76d69f96bebeff944f22c59370ca32e0b7506219))

## [0.3.0](https://github.com/7K-Inari/inari-api/compare/v0.2.0...v0.3.0) (2026-08-20)


### Features

* **plugin/v1:** add auth context, capabilities, invoke envelope, health and gRPC stubs ([#4](https://github.com/7K-Inari/inari-api/issues/4)) ([12b34d0](https://github.com/7K-Inari/inari-api/commit/12b34d04d2cafa3c179375d2103087f9b1ac1921))

## [0.2.0](https://github.com/7K-Inari/inari-api/compare/v0.1.0...v0.2.0) (2025-08-14)

### Features

* **agent:** real EventStream protocol with registration, events, and commands ([5b154ed](https://github.com/7K-Inari/inari-api/commit/5b154ed))
