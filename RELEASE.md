# Release Process

This document describes the process to release a new version of `axelard`.

1. An upgrade test should be performed over the previous release. If a migration is involved, ensure that it's been tested. Ensure that the e2e tests are also successful on the latest commit.

2. A release doc should be created under [here](https://github.com/axelarnetwork/axelar-contract-deployments/tree/main/releases) with the motivation for the release, the actions to be taken for the upgrade, and the tests that need to be performed post-upgrade.

3. If the release is a patch, then the change needs to be cherry-picked into the corresponding `releases/*` branch, and a new release is created against that branch.

4. Run the [Release workflow](./.github/workflows/release.yaml) from the branch being released, and select the appropriate release type. A release commit will be pushed to that branch, and a release tag will be created.

5. Run the [Build workflow](./.github/workflows/build-docker-image-and-binaries.yaml) to create a build from the release tag. This workflow will create a [Github release](https://github.com/axelarnetwork/axelar-core/releases) with binaries, and publish a docker image to [Docker Hub](https://hub.docker.com/r/axelarnet/axelar-core/tags).

6. Update the Github release `Changelog` with the release notes. Detail whether it's a consensus-breaking release, and whether the upgrade is not scheduled yet or it can/should be applied by node operators right away.

7. If the build steps have changed, e.g. the `go` version, `wasmvm` dependency, node config, this needs to be included in the release notes and announced to node operators. The community [scripts](https://github.com/axelarnetwork/axelarate-community/tree/main) (e.g. [wasmvm version](https://github.com/axelarnetwork/axelarate-community/blob/main/scripts/node.sh#L74)) and axelar-docs might need to be updated as well.

8. An upgrade doc should be created in [axelar-docs](https://github.com/axelarnetwork/axelar-docs/tree/main/src/content/docs/resources/mainnet/upgrades) based on the release doc. The latest [versions](https://github.com/axelarnetwork/axelar-docs/blob/main/src/config/variables.ts) should be updated once the upgrade is live on the network.

9. A governance proposal is created and announced for a consensus-breaking upgrade. For testnet, the upgrade should be scheduled with at least a 2-business day window. For mainnet, the upgrade should be scheduled with at least a 1-week window.
