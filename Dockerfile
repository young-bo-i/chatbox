# ============================================
# Stage 1: Build
# ============================================
FROM node:20-alpine AS builder

# Install build dependencies for native modules
RUN apk add --no-cache python3 make g++ git

WORKDIR /app

# Copy package files and npm config first for better layer caching
COPY package*.json ./
COPY .npmrc ./

# Install dependencies (skip postinstall as it contains electron-specific commands)
RUN npm install --ignore-scripts

# Copy source code
COPY . .

# Run only the necessary parts of postinstall for web build
RUN npx patch-package || true

# Build web version
RUN npm run build:web

# ============================================
# Stage 2: Production
# ============================================
FROM nginx:alpine

# Copy built files from builder stage
COPY --from=builder /app/release/app/dist/renderer /usr/share/nginx/html

# Copy custom nginx config for SPA routing
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
