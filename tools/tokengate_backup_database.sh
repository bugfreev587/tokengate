#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  DATABASE_URL="postgresql://..." tools/tokengate_backup_database.sh

Optional:
  TOKENGATE_BACKUP_DIR=backups
  TOKENGATE_BACKUP_PREFIX=tokengate

Creates a compressed custom-format PostgreSQL dump with pg_dump -Fc.
The script is read-only against the database.
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if ! command -v pg_dump >/dev/null 2>&1; then
  printf 'ERROR pg_dump is required but was not found in PATH.\n' >&2
  printf 'Install PostgreSQL client tools locally, or run this from the Railway backend image.\n' >&2
  exit 1
fi

database_url="${DATABASE_URL:-}"
if [[ -z "$database_url" ]]; then
  printf 'ERROR DATABASE_URL is required.\n' >&2
  usage >&2
  exit 1
fi

if [[ "$database_url" == *'${{'* || "$database_url" == *'YOUR_'* || "$database_url" == *'example.com'* ]]; then
  printf 'ERROR DATABASE_URL still looks like a placeholder.\n' >&2
  exit 1
fi

backup_dir="${TOKENGATE_BACKUP_DIR:-backups}"
backup_prefix="${TOKENGATE_BACKUP_PREFIX:-tokengate}"
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
output="${backup_dir}/${backup_prefix}-${timestamp}.dump"

mkdir -p "$backup_dir"

printf 'Starting TokenGate database backup...\n'
printf 'Output: %s\n' "$output"

pg_dump \
  --format=custom \
  --no-owner \
  --no-privileges \
  --file="$output" \
  "$database_url"

bytes="$(wc -c < "$output" | tr -d ' ')"
printf 'PASS backup complete: %s bytes\n' "$bytes"
printf 'Restore drill command:\n'
printf '  pg_restore --clean --if-exists --no-owner --no-privileges --dbname "$RESTORE_DATABASE_URL" "%s"\n' "$output"
