FROM nginx:1.25-alpine

# Install dependencies for better performance
RUN apk add --no-cache \
    curl \
    bash

# Remove default config
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Copy custom error pages
COPY html/ /usr/share/nginx/html/

# Create cache directories
RUN mkdir -p /var/cache/nginx/menu && \
    mkdir -p /var/cache/nginx/static && \
    chown -R nginx:nginx /var/cache/nginx

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost/health || exit 1

EXPOSE 80 8081

CMD ["nginx", "-g", "daemon off;"]