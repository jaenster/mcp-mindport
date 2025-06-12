# Build Instructions

This document describes how to build MCP MindPort from source and details about supported platforms.

## Quick Start

### Prerequisites

- Go 1.21 or later
- Git (for version information)

### Local Development Build

```bash
# Build for current platform
go build -o mcp-mindport .

# Build with version information
go build -ldflags="-X main.Version=dev -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o mcp-mindport .
```

### Using Build Script

```bash
# Build for current platform only
./scripts/build.sh --local

# Build for all platforms
./scripts/build.sh

# Build specific version
./scripts/build.sh --version v1.0.0

# Build for specific platforms
./scripts/build.sh --platforms linux/amd64,darwin/amd64,windows/amd64
```

## Supported Platforms

MCP MindPort supports cross-compilation for a wide range of platforms and architectures:

### Primary Platforms (Tier 1)
- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64

### Secondary Platforms (Tier 2)
- **Linux**: 386, arm, mips64le, ppc64le, riscv64, s390x
- **Windows**: 386, arm64
- **FreeBSD**: amd64, arm64
- **OpenBSD**: amd64, arm64

### Specialized Platforms (Tier 3)
- **NetBSD**: amd64
- **DragonFly BSD**: amd64
- **Solaris**: amd64
- **AIX**: ppc64
- **Android**: amd64, arm64
- **WebAssembly**: js/wasm

## Build Flags

The build process includes several build-time variables:

```bash
go build -ldflags="-X main.Version=v1.0.0 -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) -X main.GitCommit=$(git rev-parse HEAD)"
```

### Build Variables

- `main.Version`: Version string (e.g., "v1.0.0")
- `main.BuildTime`: ISO 8601 timestamp of build
- `main.GitCommit`: Git commit hash

### Build Optimization

For release builds, we use optimization flags:

```bash
-ldflags="-s -w"  # Strip debug info and symbol table
```

## Cross-Compilation

### Manual Cross-Compilation

```bash
# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o mcp-mindport-linux-arm64 .

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o mcp-mindport-windows-amd64.exe .

# macOS Universal Binary (requires additional steps)
GOOS=darwin GOARCH=amd64 go build -o mcp-mindport-darwin-amd64 .
GOOS=darwin GOARCH=arm64 go build -o mcp-mindport-darwin-arm64 .
```

### WebAssembly Build

```bash
GOOS=js GOARCH=wasm go build -o mcp-mindport.wasm .
```

Note: WebAssembly builds require a JavaScript runtime environment.

### Android Build

```bash
GOOS=android GOARCH=arm64 go build -o mcp-mindport-android-arm64 .
```

Note: Android builds may require additional configuration for mobile deployment.

## Docker Build

### Build Docker Image

```bash
# Build for current platform
docker build -t mcp-mindport .

# Build with version
docker build --build-arg VERSION=v1.0.0 -t mcp-mindport:v1.0.0 .

# Multi-platform build
docker buildx build --platform linux/amd64,linux/arm64 -t mcp-mindport .
```

### Docker Compose (Development)

```yaml
version: '3.8'
services:
  mcp-mindport:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
```

## CI/CD Pipeline

The project includes comprehensive GitHub Actions workflows:

### Continuous Integration (`ci.yml`)
- Runs on every push/PR
- Tests on multiple Go versions (1.21, 1.22)
- Cross-compilation validation
- Security scanning
- Code quality checks

### Release Pipeline (`release.yml`)
- Manual trigger only
- Version validation (prevents downgrades)
- Multi-platform builds (30+ platforms)
- Docker image publishing
- GitHub release creation
- Artifact signing

### Build Validation (`validate-build.yml`)
- Tests building on all supported platforms
- Validates cross-compilation works
- Ensures no platform-specific issues

## Release Process

### Creating a Release

1. Ensure you're on the `main` branch
2. Go to GitHub Actions → Release workflow
3. Click "Run workflow"
4. Enter the version (e.g., `v1.2.3`)
5. Select release type (release/prerelease)
6. Add optional release notes

### Version Requirements

The release system enforces semantic versioning:

- **Patch**: `v1.0.0` → `v1.0.1` (bug fixes, security patches)
- **Minor**: `v1.0.0` → `v1.1.0` (new features, backward compatible)
- **Major**: `v1.0.0` → `v2.0.0` (breaking changes)
- **Prerelease**: `v1.1.0-alpha.1`, `v1.1.0-beta.1`, `v1.1.0-rc.1`

### Platform Coverage

Release builds include:
- 30+ platform/architecture combinations
- Compressed archives (.tar.gz, .zip)
- SHA256 checksums
- Docker images (AMD64, ARM64)
- Container registry publishing (Docker Hub, GHCR)

## Dependencies

### Runtime Dependencies
- None (statically linked binary)

### Build Dependencies
- Go 1.21+ (1.22 recommended)
- Git (for version info)
- Docker (for container builds)

### Optional Dependencies
- `staticcheck` (linting)
- `gosec` (security scanning)
- `nancy` (dependency vulnerability scanning)

## Troubleshooting

### Common Build Issues

1. **CGO Errors**: Ensure `CGO_ENABLED=0` for pure Go builds
2. **Version Parsing**: Check Git is available for commit hash
3. **Cross-compilation**: Some platforms may require build constraints
4. **WebAssembly**: Requires `GOOS=js GOARCH=wasm`

### Platform-Specific Notes

- **Windows**: Anti-virus software may flag binaries
- **macOS**: Binaries may need code signing for distribution
- **Android**: Requires additional APK packaging for mobile deployment
- **WebAssembly**: Needs JavaScript host environment

### Performance Notes

- Static linking increases binary size but eliminates dependencies
- Strip flags (`-s -w`) significantly reduce binary size
- Different architectures have varying performance characteristics

## Contributing

When adding new features that might affect cross-compilation:

1. Test on multiple platforms
2. Avoid platform-specific code when possible
3. Use build tags for platform-specific features
4. Update this documentation if adding new platform support

## Security

- All binaries are built with security flags
- Dependencies are scanned for vulnerabilities
- Release artifacts include SHA256 checksums
- Docker images use minimal base images with non-root users