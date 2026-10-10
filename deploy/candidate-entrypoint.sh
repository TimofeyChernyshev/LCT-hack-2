#!/bin/sh
set -eu
mkdir -p /var/lib/fsp/uploads
chown -R app:app /var/lib/fsp/uploads
exec su-exec app /app/service
