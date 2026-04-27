#!/usr/bin/env bash

#set -euxo pipefail

DATABASE="postgres"
DB_USER="${DATABASE_USER:=username}"
DB_PASSWORD="${DATABASE_PASSWORD:=password}"
DB_NAME="${DATABASE_NAME:=merchant}"
DB_PORT="${DATABASE_PORT:=5432}"
DB_HOST="${DATABASE_HOST:=localhost}"
CONTAINER_NAME="${DATABASE}_container"
RESTART_CONTAINER="${RESTART_CONTAINER:=false}"
RUNNING_CONTAINER=$(docker ps --filter "name=$DATABASE" --format '{{.Names}}')
CONTAINER_NAME="${RUNNING_CONTAINER:-${DATABASE}_container}"


function run_container() {
  docker run \
    --rm \
    -e POSTGRES_USER="${DB_USER}" \
    -e POSTGRES_PASSWORD="${DB_PASSWORD}" \
    -e POSTGRES_DB="${DB_NAME}" \
    -p "${DB_PORT}":5432 \
    -d \
    --name "$CONTAINER_NAME" \
    ${DATABASE} \
    -N 1000 # maximum number of allowed connections
}

if [[ -n $RUNNING_CONTAINER ]]; then
  echo >&2 "There is a database container $RUNNING_CONTAINER already running"
  if ${RESTART_CONTAINER}; then
    echo >&2 "Kill database container"
    docker kill "${RUNNING_CONTAINER}"
    sleep 2
    echo >&2 "Start new database container"
    run_container
  else
    echo >&2 "You can kill container with command:"
    echo >&2 "docker kill ${RUNNING_CONTAINER}"
  fi
else
  run_container
fi

export PGPASSWORD="${DB_PASSWORD}"
until docker exec $CONTAINER_NAME pg_isready -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}"; do
  echo >&2 "Database is still unavailable - sleeping"
  sleep 1
done

# Ensure the database exists
DB_EXISTS=$(docker exec $CONTAINER_NAME psql -U "${DB_USER}" -d "postgres" -tAc "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}';")
if [[ "$DB_EXISTS" != "1" ]]; then
  echo >&2 "Database ${DB_NAME} does not exist. Creating..."
  docker exec $CONTAINER_NAME psql -U "${DB_USER}" -d "postgres" -c "CREATE DATABASE ${DB_NAME};"
else
  echo >&2 "Database ${DB_NAME} already exists."
fi
