FROM node:26.9.0-alpine3.24 AS web
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY web ./
COPY core/internal/presentation/http/ui/dist /app/core/internal/presentation/http/ui/dist
RUN npm run build

FROM golang:1.27.3-alpine3.24 AS core
WORKDIR /app/core
COPY core/go.mod core/go.sum ./
RUN go mod download
COPY core ./
COPY --from=web /app/core/internal/presentation/http/ui/dist ./internal/presentation/http/ui/dist
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /fluffy-giggle ./cmd/app

FROM alpine:3.24
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 app && mkdir -p /data && chown app /data
COPY --from=core /fluffy-giggle /app/fluffy-giggle
USER app
VOLUME ["/data"]
ENV SQLITE_PATH=/data/app.db HTTP_SERVER_HOST=0.0.0.0 HTTP_SERVER_PORT=8000
EXPOSE 8000
CMD ["/app/fluffy-giggle"]
