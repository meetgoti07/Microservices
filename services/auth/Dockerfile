# Auth Service Dockerfile (Bun + TypeScript)

FROM oven/bun:1.1-alpine AS base

WORKDIR /app

# Install dependencies
FROM base AS deps

# Copy package files
COPY package.json bun.lock ./

# Install dependencies
RUN bun install --frozen-lockfile

# Build stage
FROM base AS builder

# Copy dependencies
COPY --from=deps /app/node_modules ./node_modules

# Copy source code
COPY . .

# Install dev dependencies for build
RUN bun install --frozen-lockfile

# Production stage
FROM base AS runner

WORKDIR /app

ENV NODE_ENV=production

# Copy necessary files
COPY --from=deps /app/node_modules ./node_modules
COPY --from=builder /app/src ./src
COPY --from=builder /app/lib ./lib
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/tsconfig.json ./tsconfig.json

# Expose port
EXPOSE 3001

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3001/api/auth || exit 1

# Start the application
CMD ["bun", "run", "src/index.ts"]
