# Release Process

This document describes the release process for `axelard`.

## Overview

The release process involves:
1. Testing the release candidate
2. Creating a release tag
3. Building and publishing binaries and Docker images
4. Updating documentation and announcing the release
5. Creating governance proposals for consensus-breaking upgrades

## Versioning

Axelar follows [semantic versioning](https://semver.org/):
- **Major releases** (`vX.0.0`): Breaking changes requiring coordinated network upgrades
- **Minor releases** (`vX.Y.0`): New features, typically consensus-breaking
- **Patch releases** (`vX.Y.Z`): Bug fixes and non-consensus-breaking changes

### Branching Strategy

- **Major/Minor releases**: Created from the `main` branch. The release workflow automatically creates a `releases/<major>.<minor>.x` branch (e.g., `releases/1.4.x`).
- **Patch releases**: Cherry-pick changes into the corresponding `releases/<major>.<minor>.x` branch.

## Pre-Release Checklist

Before starting a release:

- [ ] Upgrade tests pass against the previous release
- [ ] Any state migrations have been tested
- [ ] End-to-end tests pass on the release candidate commit
- [ ] Release documentation prepared in [axelar-contract-deployments/releases](https://github.com/axelarnetwork/axelar-contract-deployments/tree/main/releases)

## Release Steps

### 1. Prepare the Release Branch

For **patch releases**, cherry-pick the required changes into the release branch:
```bash
git checkout releases/<major>.<minor>.x
git cherry-pick <commit-hash>
git push
```

For **major/minor releases**, release directly from `main`.

### 2. Dry-Run the Release

Run the **Release: Create tag (dry run)** workflow from the target branch:
- Select the appropriate release type (major, minor, or patch)
- Verify the workflow output shows the expected tag (e.g., `v1.4.0`)
- This step does not create any tags or commits

### 3. Create the Release Tag

After verifying the dry-run output, run the **Release: Create tag** workflow from the same branch:
- Select the same release type used in the dry-run
- The workflow creates a release commit, pushes the tag, and for major/minor releases creates the release branch

### 4. Build and Publish

Run the **Release: Build and upload artifacts** workflow:
- Input the release tag created in step 3 (e.g., `v1.4.0`)
- Verify the tag matches what was created by the previous workflow
- The workflow:
  - Creates a [GitHub release](https://github.com/axelarnetwork/axelar-core/releases) with signed binaries
  - Publishes Docker images to [Docker Hub](https://hub.docker.com/r/axelarnet/axelar-core/tags)

### 5. Update Release Notes

Edit the GitHub release to add:
- Summary of changes
- Whether the release is consensus-breaking
- Required actions for node operators (if any)
- Upgrade schedule (if applicable)

### 6. Update Documentation

If build requirements changed (Go version, wasmvm dependency, node config):
1. Update [axelarate-community scripts](https://github.com/axelarnetwork/axelarate-community/tree/main/scripts)
2. Create an upgrade guide in [axelar-docs/upgrades](https://github.com/axelarnetwork/axelar-docs/tree/main/src/content/docs/resources/mainnet/upgrades)
3. Update [version variables](https://github.com/axelarnetwork/axelar-docs/blob/main/src/config/variables.ts) once the upgrade is live

### 7. Governance Proposal (Consensus-Breaking Only)

For consensus-breaking releases:
- **Testnet**: Schedule with at least 2 business days notice
- **Mainnet**: Schedule with at least 1 week notice

Create and announce the governance proposal through the appropriate channels.

## Workflow Reference

| Workflow | File | Purpose |
|----------|------|---------|
| Release: Create tag (dry run) | `release-create-tag-dry-run.yaml` | Preview the tag that will be created |
| Release: Create tag | `release-create-tag.yaml` | Create version tag and release commit |
| Release: Build and upload artifacts | `release-build-and-upload-artifacts.yaml` | Build binaries and Docker images |
