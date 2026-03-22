# Changelog

## [1.8.0](https://github.com/tjg184/agent-smith/compare/v1.7.0...v1.8.0) (2026-03-22)


### Features

* handle deep nested skills ([#86](https://github.com/tjg184/agent-smith/issues/86)) ([54843ae](https://github.com/tjg184/agent-smith/commit/54843ae5cbab7d3658a92a6aca2acb03f6e0ac8f))

## [1.7.0](https://github.com/tjg184/agent-smith/compare/v1.6.3...v1.7.0) (2026-03-18)


### Features

* add universal linking ([#81](https://github.com/tjg184/agent-smith/issues/81)) ([6f65e9a](https://github.com/tjg184/agent-smith/commit/6f65e9a441a49f36f9a765d38487b3a335d47741))
* new release ([#79](https://github.com/tjg184/agent-smith/issues/79)) ([c5575fa](https://github.com/tjg184/agent-smith/commit/c5575fa58a0c9fcf61c6827424b3ce931e403a0e))

## [1.6.3](https://github.com/tjg184/agent-smith/compare/v1.6.2...v1.6.3) (2026-03-15)


### Bug Fixes

* checksum was off ([#75](https://github.com/tjg184/agent-smith/issues/75)) ([6ec330f](https://github.com/tjg184/agent-smith/commit/6ec330f96586e703409d0705fbbbeb77816b94ae))
* install script was not handling version correctly ([#73](https://github.com/tjg184/agent-smith/issues/73)) ([229c2f1](https://github.com/tjg184/agent-smith/commit/229c2f183266a154ba789cf38dfff8c45f33c501))

## [1.6.2](https://github.com/tjg184/agent-smith/compare/v1.6.1...v1.6.2) (2026-03-14)


### Bug Fixes

* **materialize,uninstall:** preserve sourceURL to avoid ambiguous com… ([#70](https://github.com/tjg184/agent-smith/issues/70)) ([6c6e2e2](https://github.com/tjg184/agent-smith/commit/6c6e2e2066f3b3c3dda9809695a8ef34494ea7e9))

## [1.6.1](https://github.com/tjg184/agent-smith/compare/v1.6.0...v1.6.1) (2026-03-13)


### Bug Fixes

* **updater:** resolve 'no lock file entry' for monorepo subdir components ([#65](https://github.com/tjg184/agent-smith/issues/65)) ([9486664](https://github.com/tjg184/agent-smith/commit/94866641737eec572beda8e26cb04e213ea10fea))

## [1.6.0](https://github.com/tjg184/agent-smith/compare/v1.5.0...v1.6.0) (2026-03-13)


### Features

* automatically swith single skill install ([#63](https://github.com/tjg184/agent-smith/issues/63)) ([a0cf907](https://github.com/tjg184/agent-smith/commit/a0cf907f08c256ef29442d3a81f5de16cab48cba))

## [1.5.0](https://github.com/tjg184/agent-smith/compare/v1.4.2...v1.5.0) (2026-03-13)


### Features

* add share feature and improve find color ([#61](https://github.com/tjg184/agent-smith/issues/61)) ([e5627c6](https://github.com/tjg184/agent-smith/commit/e5627c6e28cfaddac92812d8e24f68b2eef50336))

## [1.4.2](https://github.com/tjg184/agent-smith/compare/v1.4.1...v1.4.2) (2026-03-12)


### Bug Fixes

* directory structure for skills, agents, commands and materialize ([#60](https://github.com/tjg184/agent-smith/issues/60)) ([be3e57c](https://github.com/tjg184/agent-smith/commit/be3e57c8321eeef3ea6909518819ecc8d57265cd))
* **update:** handle multi-source components and clean up update output ([#58](https://github.com/tjg184/agent-smith/issues/58)) ([59aeee0](https://github.com/tjg184/agent-smith/commit/59aeee09f024055b0c810fe7adcfb44feb65df78))

## [1.4.1](https://github.com/tjg184/agent-smith/compare/v1.4.0...v1.4.1) (2026-03-12)


### Bug Fixes

* **cmd:** remove duplicate command registrations from root.go ([#57](https://github.com/tjg184/agent-smith/issues/57)) ([78b49f3](https://github.com/tjg184/agent-smith/commit/78b49f3cec4e2f957eb1bd6051ed2bae303a087e))
* **materialize:** use local-first strategy for update command ([#55](https://github.com/tjg184/agent-smith/issues/55)) ([990f78a](https://github.com/tjg184/agent-smith/commit/990f78a205f97f148d055189154282389b2b0e45))

## [1.4.0](https://github.com/tjg184/agent-smith/compare/v1.3.5...v1.4.0) (2026-03-11)


### Features

* **profiles:** add profile rename command ([#53](https://github.com/tjg184/agent-smith/issues/53)) ([1b483dc](https://github.com/tjg184/agent-smith/commit/1b483dc8aeb2ba271b0d0c1e064a9e6148335791))

## [1.3.5](https://github.com/tjg184/agent-smith/compare/v1.3.4...v1.3.5) (2026-03-11)


### Bug Fixes

* **linker:** link commands/agents as flat .md symlinks ([#51](https://github.com/tjg184/agent-smith/issues/51)) ([0cf4a0e](https://github.com/tjg184/agent-smith/commit/0cf4a0eb8ee5b36b9566d825d02cf608746829c7))

## [1.3.4](https://github.com/tjg184/agent-smith/compare/v1.3.3...v1.3.4) (2026-03-09)


### Bug Fixes

* **detector:** prevent skills ending in -agents/-commands from being misclassified ([#49](https://github.com/tjg184/agent-smith/issues/49)) ([9b3e7f0](https://github.com/tjg184/agent-smith/commit/9b3e7f0bcde4644e28de6beb64f751515203c894))
* **profiles:** remove lock entry on component removal ([#50](https://github.com/tjg184/agent-smith/issues/50)) ([313357c](https://github.com/tjg184/agent-smith/commit/313357cb9b71c417a8f338096eabc6c3a794210e))
* symlinks are not showing up in profile list ([#47](https://github.com/tjg184/agent-smith/issues/47)) ([3085652](https://github.com/tjg184/agent-smith/commit/3085652a6a892ef60672157e150c7a972bd1f2ba))

## [1.3.3](https://github.com/tjg184/agent-smith/compare/v1.3.2...v1.3.3) (2026-03-03)


### Bug Fixes

* root commands and agents would not show up ([#42](https://github.com/tjg184/agent-smith/issues/42)) ([a98d6fc](https://github.com/tjg184/agent-smith/commit/a98d6fc4bd22abd7c34437ddd6cd13deeca11788))
