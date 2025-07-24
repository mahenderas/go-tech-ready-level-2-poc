
#!/bin/sh
# Wait for PostgreSQL to be ready before starting the service
set -e
host="$1"
port="$2"
shift 2
until nc -z "$host" "$port"; do
  echo "Waiting for postgres at $host:$port..."
  sleep 2
done
echo "Postgres is up!"
exec "$@"

