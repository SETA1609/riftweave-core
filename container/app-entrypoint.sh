#!/bin/sh
# Start riftweave-server on the loopback, wait for it, then run nginx in the foreground.
# If the server does not come up, the container exits with its error in the log.
set -eu
riftweave-server --data /srv/ruleset/data --addr 127.0.0.1:8081 &
server=$!
for _ in $(seq 1 30); do
  wget -q -O /dev/null http://127.0.0.1:8081/v1/commands 2> /dev/null && exec nginx -g 'daemon off;'
  kill -0 "$server" 2> /dev/null || { echo "app-entrypoint: riftweave-server stopped" >&2; exit 1; }
  sleep 1
done
echo "app-entrypoint: riftweave-server did not answer in 30 seconds" >&2
exit 1
