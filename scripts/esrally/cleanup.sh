#!/bin/sh
set -eu

. "$(dirname "$0")/target_version_guard.sh"

echo "Deleting generated ESRally index ${INDEX_NAME} from ${TARGET_URL}" | tee -a "${ESRALLY_LOG_FILE}"
: > /tmp/esrally-cleanup.out
if [ -n "$ES_USERNAME" ] || [ -n "$ES_PASSWORD" ]; then
  status=$(curl -sS -o /tmp/esrally-cleanup.out -w "%{http_code}" -u "${ES_USERNAME}:${ES_PASSWORD}" -X DELETE "${TARGET_URL}/${INDEX_NAME}" || true)
else
  status=$(curl -sS -o /tmp/esrally-cleanup.out -w "%{http_code}" -X DELETE "${TARGET_URL}/${INDEX_NAME}" || true)
fi
cat /tmp/esrally-cleanup.out | tee -a "${ESRALLY_LOG_FILE}"
case "$status" in
  200|202|404) echo "Cleanup finished with HTTP ${status}" | tee -a "${ESRALLY_LOG_FILE}" ;;
  *) echo "Failed to delete index ${INDEX_NAME}: HTTP ${status}" | tee -a "${ESRALLY_LOG_FILE}"; exit 1 ;;
esac
