# syntax=docker/dockerfile:1

FROM node:22-alpine AS web-builder
WORKDIR /src/web
RUN corepack enable && corepack prepare pnpm@10.29.3 --activate
COPY web/package.json web/pnpm-lock.yaml ./
RUN HUSKY=0 pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm exec vite build && pnpm exec vue-tsc --noEmit

FROM golang:1.25-alpine AS backend-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /out/commory ./cmd/commory

FROM alpine:3.22
RUN addgroup -S commory && adduser -S commory -G commory
WORKDIR /app
COPY --from=backend-builder /out/commory /usr/local/bin/commory
COPY msglayer/schema /app/msglayer/schema
COPY --from=web-builder /src/web/dist /app/web
RUN mkdir -p /data && chown -R commory:commory /data /app
USER commory
ENV COMMORY_LISTEN_ADDR=:3000 \
    COMMORY_DB_DRIVER=sqlite \
    COMMORY_DB_DSN=/data/commory.db \
    COMMORY_SCHEMA_ROOT=/app/msglayer/schema/v0.1/root.schema.json \
    COMMORY_WEB_ROOT=/app/web \
    COMMORY_ENV=production
EXPOSE 3000
CMD ["commory", "serve"]
