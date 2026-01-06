# Changelog

# [unreleased]

## <!-- 1 -->🐛 Bug Fixes

- Guard Enabled for nil level, apply ReplaceAttr ([7113805](https://github.com/goliatone/go-logger/commit/7113805e97806e43c41dc24d861cfa9e7540b658))  - (goliatone)
- Propagate missing stdout, handlerWrapper tec from logger ([cfae71a](https://github.com/goliatone/go-logger/commit/cfae71a6aadd5f742e2914f084e1076bbcc24b5d))  - (goliatone)

## <!-- 2 -->🚜 Refactor

- Error logging into errorWithSkip ([94f281a](https://github.com/goliatone/go-logger/commit/94f281aec5ddc26355c166b7682dee9d4f8ab4a4))  - (goliatone)

## <!-- 7 -->⚙️ Miscellaneous Tasks

- Update tests ([eb1159e](https://github.com/goliatone/go-logger/commit/eb1159e09eee6c9f15feaa83b68412a2ad9ee227))  - (goliatone)

# [0.5.0](https://github.com/goliatone/go-logger/compare/v0.4.1...v0.5.0) - (2026-01-06)

## <!-- 12 -->🔖 Releases

- V0.5.0 ([1853856](https://github.com/goliatone/go-logger/commit/1853856e3350508a58aaf4df81c47b9d0dacaae1))  - (goliatone)

## <!-- 16 -->➕ Add

- BaseLogger has a handlerWrapper ([d4ca239](https://github.com/goliatone/go-logger/commit/d4ca239aa348a454c47f353c8aa987652f3a79a2))  - (goliatone)
- New options to configure writer and wrapper ([c3466eb](https://github.com/goliatone/go-logger/commit/c3466ebda681c36dec4502087575f99f075de614))  - (goliatone)

## <!-- 7 -->⚙️ Miscellaneous Tasks

- Update tests ([d8a8ecd](https://github.com/goliatone/go-logger/commit/d8a8ecdb9dee884c61e8f2e7270586cccc3fa520))  - (goliatone)
- Update readme ([6fd3167](https://github.com/goliatone/go-logger/commit/6fd316775192a387230247923217ffdfff327a5a))  - (goliatone)

# [0.4.1](https://github.com/goliatone/go-logger/compare/v0.3.0...v0.4.1) - (2025-10-17)

## <!-- 12 -->🔖 Releases

- V0.4.0 ([1d3fcbf](https://github.com/goliatone/go-logger/commit/1d3fcbfa3fb4fa9eb89709d676519240c34fa3af))  - (goliatone)

## <!-- 16 -->➕ Add

- WithFields method ([15e94a4](https://github.com/goliatone/go-logger/commit/15e94a454d9c223e07925f9837f70e910d631fa4))  - (goliatone)
- Handle case we send source in attr map ([1fa74ec](https://github.com/goliatone/go-logger/commit/1fa74ec41b461893dcfc76cf3759221931a03afc))  - (goliatone)

## <!-- 22 -->🚧 WIP

- Line output ([23b6a32](https://github.com/goliatone/go-logger/commit/23b6a329e58b9263e6a64e8344ed79382f8f9a43))  - (goliatone)

## <!-- 3 -->📚 Documentation

- Update changelog for v0.3.0 ([d5499cd](https://github.com/goliatone/go-logger/commit/d5499cd0c733e40b5aeb1b63a3394eadeaff87e8))  - (goliatone)

## <!-- 7 -->⚙️ Miscellaneous Tasks

- Update release task ([006f62a](https://github.com/goliatone/go-logger/commit/006f62a270da4cf28fdd6deb8ad379d2483b19ea))  - (goliatone)
- Update readme ([5b085d4](https://github.com/goliatone/go-logger/commit/5b085d4bff643a5f6bf0895ee1d4a949b2795124))  - (goliatone)

# [0.3.0](https://github.com/goliatone/go-logger/compare/v0.2.0...v0.3.0) - (2025-07-11)

## <!-- 1 -->🐛 Bug Fixes

- Check level for adding source ([4d59988](https://github.com/goliatone/go-logger/commit/4d59988f2e652cd9dee8dedc68ce9c80f5ca593d))  - (goliatone)
- Use custom log function to properly build stack source ([daff58e](https://github.com/goliatone/go-logger/commit/daff58ebba1a64c4ca2737ed2066f4a192050b53))  - (goliatone)

## <!-- 13 -->📦 Bumps

- Bump version: v0.3.0 ([2689730](https://github.com/goliatone/go-logger/commit/2689730bcf2c832cec9a84d07484469e735914fd))  - (goliatone)

## <!-- 16 -->➕ Add

- WithAddSource option ([1ce1ef7](https://github.com/goliatone/go-logger/commit/1ce1ef7dd9427c2d79523e4339431aa04b5461dc))  - (goliatone)

## <!-- 3 -->📚 Documentation

- Update changelog for v0.2.0 ([1f4d237](https://github.com/goliatone/go-logger/commit/1f4d23716ce70358d281a4599ace301c67f61c7a))  - (goliatone)

# [0.2.0](https://github.com/goliatone/go-logger/compare/v0.1.2...v0.2.0) - (2025-07-10)

## <!-- 1 -->🐛 Bug Fixes

- How we handle rich output in console ([ee91ed2](https://github.com/goliatone/go-logger/commit/ee91ed2137c2128fa6f78b57b4f4ae8cea02c939))  - (goliatone)
- Logger should make new copies ([ae86263](https://github.com/goliatone/go-logger/commit/ae86263274e4e99dfa011e46da3b8411d6499489))  - (goliatone)

## <!-- 13 -->📦 Bumps

- Bump version: v0.2.0 ([a8a930d](https://github.com/goliatone/go-logger/commit/a8a930d97888d062f3ac22f86526feba05d72fb3))  - (goliatone)

## <!-- 16 -->➕ Add

- Test for package ([dc77ce7](https://github.com/goliatone/go-logger/commit/dc77ce751b563523521f7c8d23921757bd284eaa))  - (goliatone)
- RichErrorHandler to augment error log output ([459b3ce](https://github.com/goliatone/go-logger/commit/459b3ce9e262ddd14970a217a63a8c83094c9dca))  - (goliatone)

## <!-- 2 -->🚜 Refactor

- Enable testing by using mock osExit ([0d5f310](https://github.com/goliatone/go-logger/commit/0d5f310cdca8e216cfd8d40b55a2a3c640f4d207))  - (goliatone)

## <!-- 3 -->📚 Documentation

- Update changelog for v0.1.2 ([87a7861](https://github.com/goliatone/go-logger/commit/87a7861a61e2b31177988b1b0b731d4c4303645f))  - (goliatone)

## <!-- 30 -->📝 Other

- PR [#1](https://github.com/goliatone/go-logger/pull/1): rich error output ([6fe3a46](https://github.com/goliatone/go-logger/commit/6fe3a4667776113e8a5aa8a106bcf287956e2149))  - (goliatone)

## <!-- 7 -->⚙️ Miscellaneous Tasks

- Update deps ([34d6984](https://github.com/goliatone/go-logger/commit/34d6984da1658251279698069320633149f06491))  - (goliatone)

# [0.1.2](https://github.com/goliatone/go-logger/compare/v0.1.1...v0.1.2) - (2025-06-26)

## <!-- 1 -->🐛 Bug Fixes

- Output ([4268fea](https://github.com/goliatone/go-logger/commit/4268fea9129edaba580bec1531a81e65bf5b6353))  - (goliatone)
- Cliff setup ([4577a39](https://github.com/goliatone/go-logger/commit/4577a39147d082a2b27906774e67b714d656af08))  - (goliatone)

## <!-- 13 -->📦 Bumps

- Bump version: v0.1.2 ([da7b483](https://github.com/goliatone/go-logger/commit/da7b4836c6e30661f1dfc271863fee24af520879))  - (goliatone)

## <!-- 3 -->📚 Documentation

- Update changelog for v0.1.1 ([cbb86b5](https://github.com/goliatone/go-logger/commit/cbb86b5d86703d834aeb883d96039900282cc5f2))  - (goliatone)

# [0.1.1](https://github.com/goliatone/go-logger/compare/v0.1.0...v0.1.1) - (2025-04-13)

## <!-- 13 -->📦 Bumps

- Bump version: v0.1.1 ([f4ef7d5](https://github.com/goliatone/go-logger/commit/f4ef7d51d6d36a110c8e24cd7e6c72875a8c6ad7))  - (goliatone)

## <!-- 16 -->➕ Add

- Message key ([0680346](https://github.com/goliatone/go-logger/commit/0680346dce97d67b9a38df14e79592a577dc7b55))  - (goliatone)
- Fatal supports getting exit code from err.Code() ([f516a56](https://github.com/goliatone/go-logger/commit/f516a5671c231381f95f40e4004174d4dde78be0))  - (goliatone)
- Support for code and status in errors ([79a1033](https://github.com/goliatone/go-logger/commit/79a10337fce5dce171593583f59bc09c6291163b))  - (goliatone)
- More info to Error ([10a2f00](https://github.com/goliatone/go-logger/commit/10a2f003c82441876d8e3803de98a196755ea920))  - (goliatone)
- New levels ([66d9e5a](https://github.com/goliatone/go-logger/commit/66d9e5a3ce6347dc23b23e148215103a682486d7))  - (goliatone)
- Args helper ([aa5467a](https://github.com/goliatone/go-logger/commit/aa5467a95246d12c8ac480be8078c8c8946b3700))  - (goliatone)
- Taskfile handle release with latest .version ([40781af](https://github.com/goliatone/go-logger/commit/40781af43c580a155e3662969c580e74ad672eee))  - (goliatone)

## <!-- 22 -->🚧 WIP

- Support statck info in errors ([fe2aa9f](https://github.com/goliatone/go-logger/commit/fe2aa9f07aa6ea36c363db6762d55f2eb3704e73))  - (goliatone)

## <!-- 3 -->📚 Documentation

- Update changelog for v0.1.0 ([97c1b7c](https://github.com/goliatone/go-logger/commit/97c1b7cf12f8d73b66f5e9c62dcb2d170df81f42))  - (goliatone)

# [0.1.0](https://github.com/goliatone/go-logger/tree/v0.1.0) - (2025-04-12)

## <!-- 1 -->🐛 Bug Fixes

- Color console handler ([2eb7323](https://github.com/goliatone/go-logger/commit/2eb7323484b5ded2cfca45666b74d79b519a24f0))  - (goliatone)

## <!-- 13 -->📦 Bumps

- Bump version: v0.1.0 ([ff0741a](https://github.com/goliatone/go-logger/commit/ff0741a4f6faec0091fd754a75dee83a37ce0481))  - (goliatone)

## <!-- 14 -->🎉 Initial Commit

- Initial commit ([3f6a101](https://github.com/goliatone/go-logger/commit/3f6a10105401aca515be665876d35164f3c9ac24))  - (goliatone)

## <!-- 16 -->➕ Add

- Logger name formatting on output ([991afa2](https://github.com/goliatone/go-logger/commit/991afa2bbeefbb1896b64bb3e7f20756eb04f924))  - (goliatone)
- Clean up interafce ([987f1a9](https://github.com/goliatone/go-logger/commit/987f1a9a669f7e83161bc340eb7eb394c1c6594f))  - (goliatone)
- Initial type definition; ([2acf954](https://github.com/goliatone/go-logger/commit/2acf9543d05c674ed21178ceda8b597d6d5e22a1))  - (goliatone)

## <!-- 2 -->🚜 Refactor

- Make writer configurable ([aeb4c41](https://github.com/goliatone/go-logger/commit/aeb4c41528e9494284c8b50abf5a93f7c3199a10))  - (goliatone)

<!-- generated by git-cliff -->
