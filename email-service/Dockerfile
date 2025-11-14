FROM node:18-alpine AS builder

WORKDIR /app

COPY package*.json ./
COPY tsconfig*.json ./
COPY nest-cli.json ./

RUN npm ci

COPY src ./src

RUN npm run build

FROM node:18-alpine

# Install procps for pgrep (used in healthcheck)
RUN apk add --no-cache procps

WORKDIR /app

COPY package*.json ./

# Install only production dependencies
RUN npm ci --only=production

# Copy built application from builder stage
COPY --from=builder /app/dist ./dist

# Copy healthcheck script
COPY healthcheck.js ./

# Health check - verifies process is running and can connect to dependencies
HEALTHCHECK --interval=30s --timeout=5s --start-period=40s --retries=3 \
  CMD node healthcheck.js || exit 1

CMD ["node", "dist/main.js"]
