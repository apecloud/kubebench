import os
import sys

log_file = os.environ["ESRALLY_LOG_FILE"]


class Tee:
    def __init__(self, stream, log_path):
        self.stream = stream
        self.log = open(log_path, "a", encoding="utf-8")

    def write(self, data):
        self.stream.write(data)
        self.log.write(data)
        self.flush()

    def flush(self):
        self.stream.flush()
        self.log.flush()


sys.stdout = Tee(sys.stdout, log_file)
sys.stderr = Tee(sys.stderr, log_file)

index_name = os.environ["INDEX_NAME"]
profile = os.environ["DATA_PROFILE"]
document_count = int(os.environ["DOCUMENT_COUNT"])
workload = os.environ.get("WORKLOAD", "all")
target_version = os.environ.get("TARGET_VERSION", "").strip()

supported_profiles = (
    "logs",
    "metrics",
    "http_logs",
    "metricbeat",
    "geonames",
    "nyc_taxis",
    "noaa",
    "nested",
    "pmc",
    "so",
    "dense_vector",
)
supported_workloads = ("index", "search", "mixed", "all")


def validate_target_version():
    if not target_version:
        return 0

    target_major = target_version.split(".", 1)[0]
    if not target_major.isdigit():
        print(
            "invalid targetVersion "
            f"{target_version}; expected an Elasticsearch version like 6.8.23, 7.17.0, or 8.12.2",
            file=sys.stderr,
        )
        sys.exit(1)

    target_major_version = int(target_major)
    if target_major_version < 6:
        print(f"generated ESRally data mode supports targetVersion 6 or newer; got {target_version}", file=sys.stderr)
        sys.exit(1)

    return target_major_version


target_major_version = validate_target_version()

if profile not in supported_profiles:
    print(f"unsupported dataProfile {profile!r}; expected one of {', '.join(supported_profiles)}", file=sys.stderr)
    sys.exit(1)

if document_count < 1:
    print(f"documentCount must be >= 1; got {document_count}", file=sys.stderr)
    sys.exit(1)

if workload not in supported_workloads:
    print(f"unsupported workload {workload!r}; expected one of {', '.join(supported_workloads)}", file=sys.stderr)
    sys.exit(1)

if profile == "dense_vector" and target_major_version and target_major_version < 7:
    print("dataProfile dense_vector requires Elasticsearch 7 or newer", file=sys.stderr)
    sys.exit(1)

print(
    "Validated ESRally generated workload: "
    f"profile={profile} workload={workload} documentCount={document_count} "
    f"targetVersion={target_version or 'unspecified'} index={index_name}"
)
