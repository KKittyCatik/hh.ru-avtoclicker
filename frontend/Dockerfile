# Stage 1: build
FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci --frozen-lockfile

COPY . .
RUN npm run build

# Stage 2: serve with nginx
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx-spa.conf /etc/nginx/conf.d/default.conf

RUN addgroup -S nginx && adduser -S nginx -G nginx 2>/dev/null || true

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s \
    CMD wget --no-verbose --tries=1 --spider http://localhost:80/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
