# Build frontend assets
FROM node:22-alpine AS web-builder
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Build Go single binary
FROM golang:1.24-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
COPY --from=web-builder /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/llmmeter ./cmd/llmmeter

# Final lightweight runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /bin/llmmeter /usr/local/bin/llmmeter

ENV PORT=8080
ENV LLMMETER_DB=/data/finops.db

EXPOSE 8080
VOLUME ["/data"]

ENTRYPOINT ["llmmeter"]
CMD ["serve", "--port", "8080", "--db", "/data/finops.db"]
