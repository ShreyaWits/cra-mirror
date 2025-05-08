# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]
- **Enhanced webhook registration and added retries**
  - Updated tests to reflect new success response structure.
  - Adjusted mock service setup.
  - Modified `RegisterWebhook` handler to return a success response DTO.
  - [`a8072ea`](https://example.com/commit/a8072ea)

---

## 📅 May 1, 2025

### ✨ Features

#### `config-service`
## May 8, 2025
    - Updates docker configurations for local development.
    - Configures temporal endpoint via environment variables.
    - Fixes workflow id for uniqueness.
    - Adds .dockerignore file to exclude unnecessary files.

    - Adds a webhook retry mechanism using Temporal workflows for reliability.
    - Enables configuring the Temporal client via the `TEMPORAL_ENDPOINT` env var.
    - Removes `.env` and `temporal-development-sql-config.yaml`, simplifies env config.
    - Adds a `CHANGELOG.md` to track changes.
    - Modifies the Webhook service to create temporal workflows for webhook retries.
    - Updates config service to start workflows for webhook notifications.
    - Removes config metadata from etcd to simplify the config structure.
    - Implements EnvRules in `env_config.go` for validation.
- **Enabled recover middleware with stack trace**
  - Middleware added to catch and log panics.
  - Stack trace enabled in recover configuration.
  - [`b8a3c0c`](https://example.com/commit/b8a3c0c)

- **Configured worker and Temporal client**
  - Added Docker support for worker service.
  - Enabled Temporal client via env variable `TEMPORAL_SERVER_URL`.
  - [`fd325a7`](https://example.com/commit/fd325a7)