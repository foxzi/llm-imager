# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.4] - 2026-02-22

### Added
- Dry-run mode (--dry-run flag or dry_run config option) for generating placeholder images without API calls

### Changed
- Config loading now cascades: system -> user -> local (later overrides earlier)
- System config path is now /etc/llm-imager/llm-imager.yaml

### Fixed
- Package names in release workflow using envsubst

## [0.1.0] - 2025-02-22

### Added
- LICENSE file (MIT)
- Provider details documentation in README
- Advanced usage examples in README
- Troubleshooting section in README
- Contributing guidelines in README
- CI workflow for testing and linting
- Release workflow for building binaries (amd64, arm64)
- DEB and RPM package generation via nfpm
- GitHub Pages package repository (APT and RPM)
- golangci-lint configuration

### Changed
- Updated OpenRouter models in README to match current API

## [0.1.0] - 2025-02-22

### Added
- Initial project structure
- CLI interface with Cobra
- Configuration system with Viper (env + YAML support)
- Provider interface and registry
- OpenAI provider (DALL-E 2, DALL-E 3, GPT Image 1)
- Google Gemini provider (gemini-2.5-flash-image)
- OpenRouter provider (universal proxy)
- Stability AI provider (Stable Image Core, Ultra, SD3)
- Replicate provider (FLUX, SDXL)
- Image output writer (PNG, JPEG, WebP)
- List command for providers and models
- Version command
