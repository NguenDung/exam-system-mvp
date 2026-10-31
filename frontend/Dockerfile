# ==============================================================================
# Multi-stage Dockerfile for React/Vite Frontend
# ==============================================================================

# Stage 1: Build static assets
FROM node:20-alpine AS builder

WORKDIR /app

# Cache node_modules
COPY package.json package-lock.json ./
RUN npm ci

# Copy source code and build production bundle
COPY . .
ENV NODE_ENV=production
RUN npm run build

# Stage 2: Serve via Nginx Alpine
FROM nginx:alpine

# Remove default nginx html files
RUN rm -rf /usr/share/nginx/html/*

# Copy build artifacts from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy custom Nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
