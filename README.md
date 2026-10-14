# Camera API

A Go backend API for managing PTZ cameras and recorded content, with support for **ONVIF**, **Hikvision ISAPI**, and **S3-compatible storage**.

## Features

- JWT authentication and role-based access control
- Camera management
- Camera configuration management
- PTZ movement, presets, and home position
- Camera content search and download
- Camera snapshot capture
- S3-compatible storage integration with MinIO
- ONVIF support
- Hikvision ISAPI support
- Rate limiting
- HashiCorp Vault for configuration and secrets
- Swagger API documentation
- Docker support

## Architecture

```text
Handler
   ↓
Usecase
   ├── Repository ─── PostgreSQL
   ├── Camera Adapter
   │      ├── ONVIF
   │      └── ISAPI
   │
   └── Storage Adapter
          └── S3 / MinIO
```

## Requirements

- Go 1.25+
- PostgreSQL
- HashiCorp Vault
- S3-compatible storage (MinIO)
- Docker (optional)

## Configuration

Configuration and secrets are loaded through HashiCorp Vault.

Environment variables are used as bootstrap configuration:

```env
PRODUCT=camera
ENV=development
FILE=api
HOST=<vault-host>
TOKEN=<vault-token>
```

Do not commit real credentials or tokens.

## Run

Development:

```bash
make dev
```

Docker:

```bash
make prod
```

Check running containers:

```bash
make ps
```

View logs:

```bash
make logs
```

## Swagger

Generate Swagger documentation:

```bash
make swagger-init
```

When enabled, Swagger UI is available at:

```text
http://localhost:8085/swagger/index.html
```

Swagger can be enabled or disabled through the application configuration.

## Supported Protocols

| Protocol | Status |
|---|---|
| ONVIF | Supported |
| Hikvision ISAPI | Supported |

## License

This project is licensed under the [MIT License](LICENSE).
