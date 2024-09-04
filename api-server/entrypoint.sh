#!/bin/sh
set -e

# Function to output messages with timestamps
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Output all environment variables for debugging purposes
log "Environment Variables:"
printenv

DB_HOST=${POSTGRES_HOST:-db}
DB_PORT=${POSTGRES_PORT:-5432}

# Wait for the PostgreSQL database to be available
log "Checking if DATABASE environment variable is set to 'postgres'..."
if [ "$DATABASE" = "postgres" ]; then
    log "Waiting for PostgreSQL to be available on $DB_HOST:$DB_PORT..."

    # Use netcat to check if the PostgreSQL port is open
    while ! nc -z $DB_HOST $DB_PORT; do
        log "PostgreSQL is not available yet, retrying..."
        sleep 5
    done

    log "PostgreSQL started successfully on $DB_HOST:$DB_PORT"
fi

# Start the application
log "Starting the application..."
exec "$@"