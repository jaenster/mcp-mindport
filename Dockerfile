# MindPort MCP Server Production Dockerfile
FROM node:18-alpine

# Set working directory
WORKDIR /app

# Create mindport user
RUN addgroup -g 1001 -S mindport && \
    adduser -S mindport -u 1001

# Copy package files
COPY package*.json ./
COPY tsconfig.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY src/ ./src/
COPY dist/ ./dist/

# Copy site build
COPY site/package*.json ./site/
RUN cd site && npm ci --only=production
COPY site/ ./site/

# Create data directory
RUN mkdir -p /data/.config/mindport && \
    chown -R mindport:mindport /data

# Switch to mindport user
USER mindport

# Set environment variables
ENV NODE_ENV=production
ENV MCP_MINDPORT_STORE_PATH=/data/.config/mindport/storage.db
ENV MCP_MINDPORT_CONFIG=/data/.config/mindport/config.yaml

# Expose ports
EXPOSE 3001

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD node dist/index.js --health || exit 1

# Default command
CMD ["node", "dist/index.js", "--web"]

# Labels
LABEL org.opencontainers.image.title="MindPort MCP Server"
LABEL org.opencontainers.image.description="High-performance Model Context Protocol server with web interface"
LABEL org.opencontainers.image.version="1.0.0"
LABEL org.opencontainers.image.authors="MindPort AI"