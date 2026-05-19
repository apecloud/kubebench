#!/bin/sh
set -eu

set -- race --pipeline=benchmark-only --target-hosts "$TARGET_HOSTS" --track-path "$TRACK_PATH" --offline --on-error "$ON_ERROR" --report-format "$REPORT_FORMAT" --report-file "$REPORT_FILE"
if [ -n "$CHALLENGE" ]; then set -- "$@" --challenge "$CHALLENGE"; fi
if [ -n "$TRACK_PARAMS" ]; then set -- "$@" --track-params "$TRACK_PARAMS"; fi
if [ -n "$CLIENT_OPTIONS" ]; then set -- "$@" --client-options "$CLIENT_OPTIONS"; fi
if [ -n "$TELEMETRY" ]; then set -- "$@" --telemetry "$TELEMETRY"; fi
if [ -n "$EXTRA_ARGS" ]; then set -- "$@" $EXTRA_ARGS; fi

set +e
esrally "$@" > /tmp/esrally.out 2>&1
status=$?
cat /tmp/esrally.out | tee "${ESRALLY_LOG_FILE}"
if [ -f "$REPORT_FILE" ]; then
  echo "Rally CSV report:" | tee -a "${ESRALLY_LOG_FILE}"
  cat "$REPORT_FILE" | tee -a "${ESRALLY_LOG_FILE}"
fi
echo "$status" > "${ESRALLY_EXIT_FILE}"
exit "$status"
