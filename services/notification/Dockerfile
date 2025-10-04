# Notification Service Dockerfile (Node.js + Fastify + Socket.io)

FROM node:20-alpine AS base

# Install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app

# Dependencies stage
FROM base AS deps

# Copy package files
COPY package.json pnpm-lock.yaml ./

# Install dependencies
RUN pnpm install --frozen-lockfile --prod

# Production stage
FROM base AS runner

WORKDIR /app

ENV NODE_ENV=production

# Copy dependencies
COPY --from=deps /app/node_modules ./node_modules

# Copy application code
COPY . .

# Create non-root user
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 notifyuser

# Change ownership
RUN chown -R notifyuser:nodejs /app

# Switch to non-root user
USER notifyuser

# Expose port
EXPOSE 3005

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
  CMD node -e "require('http').get('http://localhost:3005/health', (r) => {process.exit(r.statusCode === 200 ? 0 : 1)})" || exit 1

# Start the application
CMD ["node", "server.js"]
