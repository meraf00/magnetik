set -e

(
  cd "$(dirname "$0")"
  go build -o /tmp/build-magnetik app/*.go
)

exec /tmp/build-magnetik "$@"