# Conventional Release Configuration

# This project follows semantic versioning (semver.org)

## Version Format

# - Major: Breaking changes (X.0.0)

# - Minor: New features, backwards compatible (0.X.0)

# - Patch: Bug fixes, backwards compatible (0.0.X)

# - Prerelease: Development versions (0.0.0-prerelease)

## Commit Convention

# This project can optionally follow conventional commits:

# - feat: A new feature

# - fix: A bug fix

# - docs: Documentation only changes

# - style: Changes that do not affect the meaning of the code

# - refactor: A code change that neither fixes a bug nor adds a feature

# - test: Adding missing tests

# - chore: Changes to the build process or auxiliary tools

## Release Process

# 1. Use `make version-info` to see current version

# 2. Use `make release-{patch|minor|major|pre}` to trigger releases

# 3. GitHub Actions will automatically:

# - Increment the version using git tags

# - Build binaries for all platforms

# - Create a GitHub release with assets

# - Generate release notes

## Tools Used

# - github-tag-action: Standard semantic versioning with git tags

# - GitHub Actions: Multi-platform builds and releases

# - GitHub CLI: Local release triggering

## Manual Release

# You can also create releases manually by pushing tags:

# ```bash

# git tag v1.0.0

# git push origin v1.0.0

# ```
