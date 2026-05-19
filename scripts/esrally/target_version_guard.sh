if [ -n "$TARGET_VERSION" ]; then
  target_major="${TARGET_VERSION%%.*}"
  case "$target_major" in
    ''|*[!0-9]*) echo "invalid targetVersion ${TARGET_VERSION}; expected an Elasticsearch version like 6.8.23, 7.17.0, or 8.12.2" | tee -a "${ESRALLY_LOG_FILE}"; exit 1 ;;
  esac
  if [ "$target_major" -lt 6 ]; then
    echo "generated ESRally data mode supports targetVersion 6 or newer; got ${TARGET_VERSION}" | tee -a "${ESRALLY_LOG_FILE}"; exit 1
  fi
fi
