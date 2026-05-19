#!/bin/sh
set -eu

. "$(dirname "$0")/target_version_guard.sh"

set +e
python3 "$(dirname "$0")/prepare.py" > /tmp/esrally-prepare.out 2>&1
status=$?
set -e
cat /tmp/esrally-prepare.out | tee -a "${ESRALLY_LOG_FILE}"
exit "$status"
