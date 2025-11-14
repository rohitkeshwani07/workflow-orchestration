# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Copy shared package first
COPY ../../packages/shared /app/packages/shared

# Copy package files
COPY package.json ./
COPY ../../package.json /app/root-package.json

# Install dependencies
WORKDIR /app/packages/shared
RUN npm install

WORKDIR /app
RUN npm install

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Runtime stage - Nginx to serve static files
FROM nginx:alpine

# Copy built files
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
