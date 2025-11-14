# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Copy root package files and workspace config
COPY package.json package-lock.json* ./

# Copy all workspace packages
COPY packages/shared ./packages/shared
COPY packages/frontend ./packages/frontend

# Install all dependencies (npm workspaces will link them)
RUN npm install

# Build the shared package first
RUN npm run build -w packages/shared

# Build the frontend
RUN npm run build -w packages/frontend

# Runtime stage - Nginx to serve static files
FROM nginx:alpine

# Copy built files from the frontend workspace
COPY --from=builder /app/packages/frontend/dist /usr/share/nginx/html

# Copy nginx configuration
COPY packages/frontend/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
