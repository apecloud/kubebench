import base64
import os
import ssl
import sys
import urllib.error
import urllib.request

target_url = os.environ["TARGET_URL"].rstrip("/")
index_name = os.environ["INDEX_NAME"]
target_version = os.environ.get("TARGET_VERSION", "").strip()
username = os.environ.get("ES_USERNAME", "")
password = os.environ.get("ES_PASSWORD", "")
insecure_skip_verify = os.environ.get("ES_INSECURE_SKIP_VERIFY", "").lower() == "true"
log_file = os.environ["ESRALLY_LOG_FILE"]


def log(message="", file=None):
    print(message, file=file)
    with open(log_file, "a", encoding="utf-8") as handle:
        handle.write(message + "\n")


def validate_target_version():
    if not target_version:
        return

    target_major_version = target_version.split(".", 1)[0]
    if not target_major_version.isdigit():
        log(
            "invalid targetVersion "
            f"{target_version}; expected an Elasticsearch version like 6.8.23, 7.17.0, or 8.12.2",
            file=sys.stderr,
        )
        sys.exit(1)

    if int(target_major_version) < 6:
        log(f"generated ESRally data mode supports targetVersion 6 or newer; got {target_version}", file=sys.stderr)
        sys.exit(1)


def delete_index():
    request = urllib.request.Request(f"{target_url}/{index_name}", method="DELETE")
    if username and password:
        token = base64.b64encode(f"{username}:{password}".encode("utf-8")).decode("ascii")
        request.add_header("Authorization", "Basic " + token)

    context = ssl._create_unverified_context() if target_url.startswith("https://") and insecure_skip_verify else None
    try:
        with urllib.request.urlopen(request, timeout=60, context=context) as response:
            return response.status, response.read().decode("utf-8", errors="replace")
    except urllib.error.HTTPError as err:
        return err.code, err.read().decode("utf-8", errors="replace")


validate_target_version()
log(f"Deleting generated ESRally index {index_name} from {target_url}")

try:
    status, body = delete_index()
except Exception as err:
    log(f"Failed to delete index {index_name}: {err}", file=sys.stderr)
    sys.exit(1)

if body:
    log(body.rstrip())

if status in (200, 202, 404):
    log(f"Cleanup finished with HTTP {status}")
else:
    log(f"Failed to delete index {index_name}: HTTP {status}", file=sys.stderr)
    sys.exit(1)
