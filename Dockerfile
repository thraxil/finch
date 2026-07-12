FROM golang:1.25-bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /finch .

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates sqlite3 && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

COPY --from=builder /finch /bin/finch
COPY templates /workspace/templates
COPY media /workspace/media
COPY schema.sql /workspace/schema.sql
COPY seed.sql /workspace/seed.sql

# Environment variables matching previous fly.toml defaults
ENV FINCH_DB_FILE="/data/database.db"
ENV FINCH_PORT="8000"
ENV FINCH_MEDIA_DIR="/workspace/media"
ENV FINCH_TEMPLATE_DIR="/workspace/templates"
ENV FINCH_ITEMS_PER_PAGE="50"

EXPOSE 8000

CMD ["/bin/finch"]
