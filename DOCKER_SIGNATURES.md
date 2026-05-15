# Docker Image Signatures

All official Traceway Docker images are cryptographically signed using [Cosign](https://docs.sigstore.dev/cosign/overview/), a CNCF tool for container image signing and verification.

## Why Image Signatures Matter

Signed images provide:
- **Authenticity**: Verify the image comes from the official Traceway project
- **Integrity**: Ensure the image hasn't been modified or tampered with
- **Transparency**: Signatures are publicly verifiable using GitHub's OIDC token

## Available Images

Traceway publishes three Docker image variants to GitHub Container Registry (GHCR):

| Image | Purpose | Size | Best For |
|-------|---------|------|----------|
| `ghcr.io/tracewayapp/traceway:v*` | **Full** — ClickHouse + PostgreSQL + supervisord | ~600MB | Large deployments, multi-service setup |
| `ghcr.io/tracewayapp/traceway:v*-minimal` | **Minimal** — single binary with external databases | ~20-30MB | Kubernetes, scalable setups, external database infrastructure |
| `ghcr.io/tracewayapp/traceway:v*-sqlite` | **SQLite** — embedded SQLite, single binary, zero dependencies | ~50-80MB | Small VPS, testing, single-server deployments, embedded mode |

All images are cryptographically signed. Pull the latest release:

```bash
# Full image (all services included)
docker pull ghcr.io/tracewayapp/traceway:latest

# Minimal image (external databases)
docker pull ghcr.io/tracewayapp/traceway:minimal

# SQLite image (embedded, no external dependencies)
docker pull ghcr.io/tracewayapp/traceway:sqlite
```

## Verifying Signatures

### Install Cosign

**macOS:**
```bash
brew install cosign
```

**Linux:**
```bash
wget https://github.com/sigstore/cosign/releases/latest/download/cosign-linux-amd64
sudo mv cosign-linux-amd64 /usr/local/bin/cosign
sudo chmod +x /usr/local/bin/cosign
```

**Windows (PowerShell):**
```powershell
Invoke-WebRequest -Uri "https://github.com/sigstore/cosign/releases/latest/download/cosign-windows-amd64.exe" -OutFile cosign.exe
# Add cosign.exe to your PATH
```

### Verify Signature

```bash
# Full image
cosign verify \
  --certificate-identity-regexp '.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  ghcr.io/tracewayapp/traceway:latest

# Minimal image
cosign verify \
  --certificate-identity-regexp '.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  ghcr.io/tracewayapp/traceway:minimal

# SQLite image
cosign verify \
  --certificate-identity-regexp '.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  ghcr.io/tracewayapp/traceway:sqlite
```

### Example Output

A successful verification looks like:
```
Verification successful!
```

If verification fails, the image may have been tampered with and should not be used.

## How Signatures Are Generated

1. **Build & Push**: GitHub Actions builds the Docker image and pushes it to GHCR
2. **Sign**: Cosign signs the image using GitHub's OIDC token (keyless signing)
3. **Store**: Signature is stored in GHCR alongside the image

## Using in Docker Compose

```yaml
version: '3.8'

services:
  traceway:
    image: ghcr.io/tracewayapp/traceway:latest
    # Optionally verify the signature before running:
    # Run: cosign verify ... ghcr.io/tracewayapp/traceway:latest
    ports:
      - "80:80"
      - "8082:8082"
    volumes:
      - clickhouse_data:/var/lib/clickhouse
      - postgres_data:/var/lib/postgresql/data
    environment:
      GIN_MODE: release

volumes:
  clickhouse_data:
  postgres_data:
```

## Minimal Image with External Databases

For scalable deployments with external ClickHouse/PostgreSQL:

```yaml
version: '3.8'

services:
  traceway:
    image: ghcr.io/tracewayapp/traceway:minimal
    ports:
      - "80:80"
      - "8082:8082"
    environment:
      GIN_MODE: release
      CLICKHOUSE_SERVER: clickhouse:9000
      POSTGRES_HOST: postgres
      POSTGRES_PORT: 5432
      # ... other env vars

  clickhouse:
    image: clickhouse/clickhouse-server:latest
    # ... configuration

  postgres:
    image: postgres:15
    # ... configuration
```

## SQLite Image (Embedded, Zero Dependencies)

For small deployments, testing, or single-server setups:

```yaml
version: '3.8'

services:
  traceway:
    image: ghcr.io/tracewayapp/traceway:sqlite
    ports:
      - "80:80"
      - "8082:8082"
    volumes:
      - traceway_data:/data
    environment:
      GIN_MODE: release

volumes:
  traceway_data:
```

The SQLite image:
- Stores all data in `/data/traceway.db` (SQLite database)
- Stores blobs in `/data/storage` (if using local storage)
- Requires only one volume mount for all persistent state
- Perfect for VPS, small instances, or development

To persist data across restarts, mount a host folder or named volume at `/data`:

```bash
docker run -v /var/lib/traceway:/data ghcr.io/tracewayapp/traceway:sqlite
```

## Troubleshooting

**"Verification failed" error:**
- Ensure you're using the correct image URI (with version tag or `latest`)
- Check your internet connection (Cosign needs to fetch OIDC tokens)
- Try verifying a different version tag

**"cosign not found":**
- Ensure Cosign is installed and in your `$PATH`
- Run `cosign version` to verify installation

**"Certificate verification failed":**
- This indicates the image signature is invalid
- Do not use this image — report it to the Traceway team

## More Information

- [Cosign Documentation](https://docs.sigstore.dev/cosign/overview/)
- [Traceway Self-Hosting Guide](https://docs.tracewayapp.com/server)
- [SBOM (Software Bill of Materials)](https://docs.sigstore.dev/cosign/sbom/)

## Questions?

If you have questions about Docker image security or signatures, [open an issue](https://github.com/tracewayapp/traceway/issues) or ask in the [Traceway Discord](https://discord.gg/9tPn2SB3).
