#!/bin/bash

# Build script for MCP MindPort
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
VERSION=${VERSION:-"dev"}
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
OUTPUT_DIR=${OUTPUT_DIR:-"build"}
PLATFORMS=${PLATFORMS:-"linux/amd64,darwin/amd64,darwin/arm64,windows/amd64"}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -p|--platforms)
            PLATFORMS="$2"
            shift 2
            ;;
        --local)
            PLATFORMS="$(go env GOOS)/$(go env GOARCH)"
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -v, --version VERSION    Set version (default: dev)"
            echo "  -o, --output DIR         Output directory (default: build)"
            echo "  -p, --platforms LIST     Comma-separated list of OS/ARCH (default: multi-platform)"
            echo "      --local              Build only for current platform"
            echo "  -h, --help               Show this help"
            echo ""
            echo "Examples:"
            echo "  $0 --version v1.0.0"
            echo "  $0 --local"
            echo "  $0 --platforms linux/amd64,darwin/amd64"
            exit 0
            ;;
        *)
            echo "Unknown option $1"
            exit 1
            ;;
    esac
done

echo -e "${BLUE}Building MCP MindPort${NC}"
echo -e "${YELLOW}Version:    ${VERSION}${NC}"
echo -e "${YELLOW}Build Time: ${BUILD_TIME}${NC}"
echo -e "${YELLOW}Git Commit: ${GIT_COMMIT}${NC}"
echo -e "${YELLOW}Platforms:  ${PLATFORMS}${NC}"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build for each platform
IFS=',' read -ra PLATFORM_LIST <<< "$PLATFORMS"
for platform in "${PLATFORM_LIST[@]}"; do
    IFS='/' read -ra PLATFORM_PARTS <<< "$platform"
    GOOS=${PLATFORM_PARTS[0]}
    GOARCH=${PLATFORM_PARTS[1]}
    
    # Set binary name with platform suffix
    BINARY_NAME="mcp-mindport-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        BINARY_NAME="${BINARY_NAME}.exe"
    fi
    
    echo -e "${BLUE}Building for ${GOOS}/${GOARCH}...${NC}"
    
    # Build the binary
    env GOOS="$GOOS" GOARCH="$GOARCH" go build \
        -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o "${OUTPUT_DIR}/${BINARY_NAME}" .
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Successfully built ${BINARY_NAME}${NC}"
        
        # Get file size
        if command -v du >/dev/null 2>&1; then
            SIZE=$(du -h "${OUTPUT_DIR}/${BINARY_NAME}" | cut -f1)
            echo -e "   Size: ${SIZE}"
        fi
        
        # Create checksums
        if command -v sha256sum >/dev/null 2>&1; then
            (cd "$OUTPUT_DIR" && sha256sum "$BINARY_NAME" > "${BINARY_NAME}.sha256")
            echo -e "   Checksum created"
        elif command -v shasum >/dev/null 2>&1; then
            (cd "$OUTPUT_DIR" && shasum -a 256 "$BINARY_NAME" > "${BINARY_NAME}.sha256")
            echo -e "   Checksum created"
        fi
    else
        echo -e "${RED}Failed to build for ${GOOS}/${GOARCH}${NC}"
        exit 1
    fi
    
    echo ""
done

echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${YELLOW}Output directory: ${OUTPUT_DIR}${NC}"
echo ""
echo -e "${BLUE}Built binaries:${NC}"
ls -la "$OUTPUT_DIR"/ | grep -E '\.(exe|sha256)$|mcp-mindport'