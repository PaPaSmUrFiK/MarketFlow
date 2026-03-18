#!/usr/bin/env bash
set -e

DB_CONTAINER=${DB_CONTAINER:-infra-postgres}
DB_ADMIN_USER=${DB_ADMIN_USER:-postgres}
DB_USER=${DB_USER:-identity_user}
#DB_USER=${DB_USER:?"DB_USER is required"}
DB_PASSWORD=${DB_PASSWORD:-identity_pass}
#DB_PASSWORD=${DB_PASSWORD:?"DB_PASSWORD is required"}


SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SQL_DIR="$SCRIPT_DIR/db"

echo "▶ Initializing MarketFlow database"
echo "▶ Container : $DB_CONTAINER"
echo "▶ Admin user: $DB_ADMIN_USER"
echo "▶ App role  : $DB_USER"

run_sql() {
  local file=$1
  echo "▶ Running $(basename "$file")"

  docker exec -i "$DB_CONTAINER" \
    psql -U "$DB_ADMIN_USER" < "$file"
}

run_sql_with_vars() {
    local file=$1
    echo "▶ Running $(basename "$file") (with vars)"
    docker exec -i "$DB_CONTAINER" \
        psql -U "$DB_ADMIN_USER" \
             -v ON_ERROR_STOP=1 \
             -1 \
             -c "SET app.identity_user = '$DB_USER';" \
             -c "SET app.identity_pass = '$DB_PASSWORD';" \
        < "$file"
}

run_sql  "$SQL_DIR/01-create-db.sql"
run_sql_with_vars  "$SQL_DIR/02-create-schemas.sql"

echo "✅ MarketFlow database bootstrap completed"