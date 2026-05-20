import datetime
import json
import os
import random
import sys
from pathlib import Path

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
track_path = Path(os.environ["GENERATED_TRACK_PATH"])
documents_file = Path(os.environ["DOCUMENTS_FILE"])

random.seed(7)
base_time = datetime.datetime(2024, 1, 1, tzinfo=datetime.timezone.utc)
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


def timestamp(delta):
    return (base_time + delta).isoformat().replace("+00:00", "Z")


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


def typed_mappings(properties):
    mappings = {"properties": properties}
    if target_major_version == 6:
        return {"_doc": mappings}
    return mappings


def index_body_for_profile():
    common_keyword = {"type": "keyword"}
    common_text = {"type": "text"}
    common_date = {"type": "date"}
    common_long = {"type": "long"}
    common_double = {"type": "double"}
    profile_mappings = {
        "logs": {
            "@timestamp": common_date,
            "service": common_keyword,
            "host": common_keyword,
            "method": common_keyword,
            "path": common_keyword,
            "status": common_long,
            "latency_ms": common_double,
            "bytes": common_long,
            "message": common_text,
        },
        "metrics": {
            "@timestamp": common_date,
            "host": common_keyword,
            "pod": common_keyword,
            "container": common_keyword,
            "namespace": common_keyword,
            "cpu_pct": common_double,
            "memory_mb": common_double,
            "disk_read_bytes": common_long,
            "disk_write_bytes": common_long,
            "network_rx_bytes": common_long,
            "network_tx_bytes": common_long,
        },
        "http_logs": {
            "@timestamp": common_date,
            "client_ip": {"type": "ip"},
            "method": common_keyword,
            "url": common_keyword,
            "status": common_long,
            "bytes": common_long,
            "response_time_ms": common_double,
            "user_agent": common_text,
            "referer": common_keyword,
        },
        "metricbeat": {
            "@timestamp": common_date,
            "host": {"properties": {"name": common_keyword}},
            "event": {"properties": {"dataset": common_keyword}},
            "system": {
                "properties": {
                    "cpu": {"properties": {"total": {"properties": {"pct": common_double}}}},
                    "memory": {"properties": {"actual": {"properties": {"used": {"properties": {"pct": common_double}}}}}},
                    "filesystem": {"properties": {"used": {"properties": {"pct": common_double}}}},
                },
            },
            "kubernetes": {
                "properties": {
                    "pod": {"properties": {"name": common_keyword}},
                    "namespace": common_keyword,
                },
            },
        },
        "geonames": {
            "name": common_text,
            "asciiname": common_keyword,
            "country_code": common_keyword,
            "feature_class": common_keyword,
            "feature_code": common_keyword,
            "population": common_long,
            "elevation": common_long,
            "location": {"type": "geo_point"},
        },
        "nyc_taxis": {
            "pickup_datetime": common_date,
            "dropoff_datetime": common_date,
            "pickup_location": {"type": "geo_point"},
            "dropoff_location": {"type": "geo_point"},
            "passenger_count": common_long,
            "trip_distance": common_double,
            "fare_amount": common_double,
            "payment_type": common_keyword,
            "vendor_id": common_keyword,
        },
        "noaa": {
            "@timestamp": common_date,
            "station_id": common_keyword,
            "station_name": common_text,
            "location": {"type": "geo_point"},
            "temperature_c": common_double,
            "precipitation_mm": common_double,
            "wind_speed_mps": common_double,
            "weather_type": common_keyword,
        },
        "nested": {
            "@timestamp": common_date,
            "order_id": common_keyword,
            "customer_id": common_keyword,
            "status": common_keyword,
            "items": {
                "type": "nested",
                "properties": {
                    "sku": common_keyword,
                    "category": common_keyword,
                    "quantity": common_long,
                    "price": common_double,
                },
            },
        },
        "pmc": {
            "title": common_text,
            "abstract": common_text,
            "body": common_text,
            "journal": common_keyword,
            "year": common_long,
            "authors": common_keyword,
        },
        "so": {
            "creation_date": common_date,
            "title": common_text,
            "body": common_text,
            "tags": common_keyword,
            "score": common_long,
            "answer_count": common_long,
            "accepted_answer": {
                "type": "nested",
                "properties": {
                    "body": common_text,
                    "score": common_long,
                    "user": common_keyword,
                },
            },
        },
        "dense_vector": {
            "@timestamp": common_date,
            "title": common_text,
            "text": common_text,
            "category": common_keyword,
            "embedding": {"type": "dense_vector", "dims": 8},
        },
    }
    return {
        "settings": {"number_of_shards": 1, "number_of_replicas": 0},
        "mappings": typed_mappings(profile_mappings[profile]),
    }


def log_doc(i):
    status = random.choice([200, 200, 200, 201, 204, 301, 400, 404, 500])
    service = random.choice(["api", "checkout", "search", "billing"])
    path = random.choice(["/api/search", "/api/orders", "/api/cart", "/health", "/login"])
    return {
        "@timestamp": timestamp(datetime.timedelta(seconds=i)),
        "service": service,
        "host": f"host-{i % 12}",
        "method": random.choice(["GET", "POST", "PUT", "DELETE"]),
        "path": path,
        "status": status,
        "latency_ms": round(random.uniform(2, 900), 3),
        "bytes": random.randint(128, 65536),
        "message": f"{service} handled {path} with status {status}",
    }


def metrics_doc(i):
    return {
        "@timestamp": timestamp(datetime.timedelta(seconds=i)),
        "host": f"node-{i % 10}",
        "pod": f"pod-{i % 30}",
        "container": random.choice(["app", "sidecar", "worker"]),
        "namespace": random.choice(["default", "search", "payments"]),
        "cpu_pct": round(random.uniform(0, 100), 3),
        "memory_mb": round(random.uniform(128, 32768), 3),
        "disk_read_bytes": random.randint(0, 10485760),
        "disk_write_bytes": random.randint(0, 10485760),
        "network_rx_bytes": random.randint(0, 10485760),
        "network_tx_bytes": random.randint(0, 10485760),
    }


def http_logs_doc(i):
    status = random.choice([200, 200, 200, 200, 201, 204, 301, 302, 400, 401, 404, 429, 500, 503])
    url = random.choice(["/login", "/products", "/checkout", "/search", "/api/v1/orders", "/static/app.js"])
    return {
        "@timestamp": timestamp(datetime.timedelta(milliseconds=i * 20)),
        "client_ip": f"10.{i % 255}.{(i * 7) % 255}.{(i * 13) % 255}",
        "method": random.choice(["GET", "GET", "POST", "PUT", "DELETE"]),
        "url": url,
        "status": status,
        "bytes": random.randint(256, 1048576),
        "response_time_ms": round(random.lognormvariate(4.2, 0.8), 3),
        "user_agent": random.choice(["Mozilla/5.0", "curl/8.0", "kube-probe/1.28", "Elastic-Heartbeat/8.x"]),
        "referer": random.choice(["-", "https://example.com", "https://search.example.com"]),
    }


def metricbeat_doc(i):
    namespace = random.choice(["default", "observability", "payments", "search"])
    return {
        "@timestamp": timestamp(datetime.timedelta(seconds=i)),
        "host": {"name": f"node-{i % 24}"},
        "event": {"dataset": random.choice(["system.cpu", "system.memory", "system.filesystem", "kubernetes.pod"])},
        "system": {
            "cpu": {"total": {"pct": round(random.random(), 4)}},
            "memory": {"actual": {"used": {"pct": round(random.random(), 4)}}},
            "filesystem": {"used": {"pct": round(random.random(), 4)}},
        },
        "kubernetes": {
            "pod": {"name": f"{random.choice(['api', 'worker', 'search'])}-{i % 50}"},
            "namespace": namespace,
        },
    }


def geonames_doc(i):
    city = random.choice(["Shanghai", "Beijing", "New York", "London", "Tokyo", "Berlin", "Sydney", "Singapore"])
    return {
        "name": f"{city} {i}",
        "asciiname": city.lower().replace(" ", "-"),
        "country_code": random.choice(["CN", "US", "GB", "JP", "DE", "AU", "SG"]),
        "feature_class": random.choice(["P", "A", "H", "S"]),
        "feature_code": random.choice(["PPL", "PPLA", "ADM1", "LK", "AIRP"]),
        "population": random.randint(0, 25000000),
        "elevation": random.randint(-100, 4500),
        "location": {"lat": round(random.uniform(-85, 85), 6), "lon": round(random.uniform(-180, 180), 6)},
    }


def nyc_taxis_doc(i):
    pickup = base_time + datetime.timedelta(minutes=i)
    dropoff = pickup + datetime.timedelta(minutes=random.randint(3, 75))
    return {
        "pickup_datetime": timestamp(datetime.timedelta(minutes=i)),
        "dropoff_datetime": dropoff.isoformat().replace("+00:00", "Z"),
        "pickup_location": {"lat": round(random.uniform(40.55, 40.90), 6), "lon": round(random.uniform(-74.05, -73.75), 6)},
        "dropoff_location": {"lat": round(random.uniform(40.55, 40.90), 6), "lon": round(random.uniform(-74.05, -73.75), 6)},
        "passenger_count": random.randint(1, 6),
        "trip_distance": round(random.uniform(0.2, 35.0), 2),
        "fare_amount": round(random.uniform(3.5, 120.0), 2),
        "payment_type": random.choice(["cash", "card", "no_charge", "dispute"]),
        "vendor_id": random.choice(["CMT", "VTS", "DDS"]),
    }


def noaa_doc(i):
    return {
        "@timestamp": timestamp(datetime.timedelta(hours=i)),
        "station_id": f"STN{i % 10000:05d}",
        "station_name": random.choice(["Central Observatory", "Harbor Station", "Airport Station", "Mountain Station"]),
        "location": {"lat": round(random.uniform(-75, 75), 6), "lon": round(random.uniform(-180, 180), 6)},
        "temperature_c": round(random.uniform(-35, 45), 2),
        "precipitation_mm": round(max(0, random.gauss(2.0, 8.0)), 2),
        "wind_speed_mps": round(random.uniform(0, 45), 2),
        "weather_type": random.choice(["clear", "cloudy", "rain", "snow", "storm", "fog"]),
    }


def nested_doc(i):
    items = []
    for n in range(random.randint(1, 5)):
        items.append({
            "sku": f"sku-{(i + n) % 2000}",
            "category": random.choice(["books", "electronics", "grocery", "apparel", "tools"]),
            "quantity": random.randint(1, 8),
            "price": round(random.uniform(1.0, 500.0), 2),
        })
    return {
        "@timestamp": timestamp(datetime.timedelta(seconds=i)),
        "order_id": f"order-{i}",
        "customer_id": f"customer-{i % 1000}",
        "status": random.choice(["created", "paid", "shipped", "returned"]),
        "items": items,
    }


def pmc_doc(i):
    topic = random.choice(["search relevance", "distributed systems", "genomics", "immunology", "observability"])
    return {
        "title": f"Study of {topic} cohort {i}",
        "abstract": f"This paper evaluates {topic} using controlled experiments and statistical analysis.",
        "body": " ".join([f"{topic} measurement shows repeatable behavior under benchmark condition {i % 17}." for _ in range(8)]),
        "journal": random.choice(["PMC Systems", "Journal of Search", "Computational Biology", "Medical Informatics"]),
        "year": random.randint(1995, 2026),
        "authors": [f"author-{(i + n) % 500}" for n in range(random.randint(1, 5))],
    }


def so_doc(i):
    tag_pool = ["elasticsearch", "kubernetes", "golang", "python", "performance", "query-dsl", "indexing"]
    tags = random.sample(tag_pool, random.randint(2, 4))
    return {
        "creation_date": timestamp(datetime.timedelta(days=i % 3650)),
        "title": f"How to tune {' '.join(tags[:2])} workload {i}?",
        "body": f"I am benchmarking {' and '.join(tags)} and need predictable latency for query {i}.",
        "tags": tags,
        "score": random.randint(-2, 250),
        "answer_count": random.randint(0, 12),
        "accepted_answer": [{
            "body": "Use representative mappings, warmup, and stable benchmark data before comparing results.",
            "score": random.randint(0, 500),
            "user": f"user-{i % 1000}",
        }],
    }


def dense_vector_doc(i):
    return {
        "@timestamp": timestamp(datetime.timedelta(seconds=i)),
        "title": f"Vector benchmark document {i}",
        "text": f"Generated semantic document about {random.choice(['search', 'databases', 'cloud', 'observability'])} number {i}",
        "category": random.choice(["search", "database", "cloud", "observability"]),
        "embedding": [round(random.uniform(-1, 1), 6) for _ in range(8)],
    }


profile_docs = {
    "logs": log_doc,
    "metrics": metrics_doc,
    "http_logs": http_logs_doc,
    "metricbeat": metricbeat_doc,
    "geonames": geonames_doc,
    "nyc_taxis": nyc_taxis_doc,
    "noaa": noaa_doc,
    "nested": nested_doc,
    "pmc": pmc_doc,
    "so": so_doc,
    "dense_vector": dense_vector_doc,
}


def write_documents():
    track_path.mkdir(parents=True, exist_ok=True)
    documents_file.parent.mkdir(parents=True, exist_ok=True)
    make_doc = profile_docs[profile]
    total_bytes = 0
    with documents_file.open("w", encoding="utf-8") as handle:
        for i in range(document_count):
            line = json.dumps(make_doc(i), separators=(",", ":")) + "\n"
            total_bytes += len(line.encode("utf-8"))
            handle.write(line)
    return total_bytes


def query_templates():
    common = {
        "logs": {
            "term": {"term": {"service": "api"}},
            "range": {"range": {"latency_ms": {"gte": 100}}},
            "aggregation": {"aggs": {"by_status": {"terms": {"field": "status"}}}, "size": 0},
        },
        "metrics": {
            "term": {"term": {"namespace": "default"}},
            "range": {"range": {"cpu_pct": {"gte": 75}}},
            "aggregation": {"aggs": {"by_host": {"terms": {"field": "host"}}}, "size": 0},
        },
        "http_logs": {
            "term": {"term": {"method": "GET"}},
            "range": {"range": {"response_time_ms": {"gte": 250}}},
            "aggregation": {"aggs": {"by_status": {"terms": {"field": "status"}}}, "size": 0},
        },
        "metricbeat": {
            "term": {"term": {"event.dataset": "system.cpu"}},
            "range": {"range": {"system.cpu.total.pct": {"gte": 0.75}}},
            "aggregation": {"aggs": {"by_namespace": {"terms": {"field": "kubernetes.namespace"}}}, "size": 0},
        },
        "geonames": {
            "term": {"term": {"country_code": "CN"}},
            "range": {"range": {"population": {"gte": 1000000}}},
            "aggregation": {"aggs": {"by_country": {"terms": {"field": "country_code"}}}, "size": 0},
        },
        "nyc_taxis": {
            "term": {"term": {"payment_type": "card"}},
            "range": {"range": {"fare_amount": {"gte": 20}}},
            "aggregation": {"aggs": {"by_vendor": {"terms": {"field": "vendor_id"}}}, "size": 0},
        },
        "noaa": {
            "term": {"term": {"weather_type": "clear"}},
            "range": {"range": {"temperature_c": {"gte": 25}}},
            "aggregation": {"aggs": {"by_weather": {"terms": {"field": "weather_type"}}}, "size": 0},
        },
        "nested": {
            "term": {"term": {"status": "paid"}},
            "range": {"range": {"@timestamp": {"gte": "2024-01-01T00:00:00Z"}}},
            "aggregation": {"aggs": {"by_status": {"terms": {"field": "status"}}}, "size": 0},
        },
        "pmc": {
            "term": {"term": {"journal": "Journal of Search"}},
            "range": {"range": {"year": {"gte": 2015}}},
            "aggregation": {"aggs": {"by_year": {"terms": {"field": "year"}}}, "size": 0},
        },
        "so": {
            "term": {"term": {"tags": "elasticsearch"}},
            "range": {"range": {"score": {"gte": 10}}},
            "aggregation": {"aggs": {"by_tag": {"terms": {"field": "tags"}}}, "size": 0},
        },
        "dense_vector": {
            "term": {"term": {"category": "search"}},
            "range": {"range": {"@timestamp": {"gte": "2024-01-01T00:00:00Z"}}},
            "aggregation": {"aggs": {"by_category": {"terms": {"field": "category"}}}, "size": 0},
        },
    }
    return common[profile]


def search_operation(name, body):
    return {
        "name": name,
        "operation-type": "search",
        "index": "{{target_index}}",
        "body": body,
    }


def task(operation, **kwargs):
    item = {"operation": operation}
    item.update(kwargs)
    return item


def build_track(uncompressed_bytes):
    document_entry = {
        "source-file": documents_file.name,
        "document-count": document_count,
        "uncompressed-bytes": uncompressed_bytes,
        "target-index": "{{target_index}}",
    }
    if target_major_version == 6:
        document_entry["target-type"] = "_doc"

    queries = query_templates()
    operations = [
        {"name": "delete-index", "operation-type": "delete-index", "only-if-exists": True},
        {"name": "create-index", "operation-type": "create-index"},
        {"name": "bulk-index", "operation-type": "bulk", "bulk-size": 500},
        {"name": "refresh", "operation-type": "refresh", "index": "{{target_index}}"},
        search_operation("match-all-search", {"query": {"match_all": {}}}),
        search_operation("term-or-keyword-search", {"query": queries["term"]}),
        search_operation("range-search", {"query": queries["range"]}),
        search_operation("aggregation", queries["aggregation"]),
        search_operation("mixed-bulk-search", {"query": {"match_all": {}}}),
    ]

    index_tasks = [
        task("delete-index"),
        task("create-index"),
        task("bulk-index", clients=1),
        task("refresh"),
    ]
    search_tasks = index_tasks + [
        task("match-all-search", iterations=20),
        task("term-or-keyword-search", iterations=20),
        task("range-search", iterations=20),
        task("aggregation", iterations=10),
    ]
    mixed_tasks = [
        task("delete-index"),
        task("create-index"),
        task("bulk-index", clients=1),
        task("refresh"),
        task("mixed-bulk-search", iterations=10),
        task("term-or-keyword-search", iterations=10),
        task("bulk-index", clients=1),
        task("refresh"),
        task("range-search", iterations=10),
        task("aggregation", iterations=5),
    ]
    all_tasks = index_tasks + [
        task("match-all-search", iterations=20),
        task("term-or-keyword-search", iterations=20),
        task("range-search", iterations=20),
        task("aggregation", iterations=10),
        task("bulk-index", clients=1),
        task("refresh"),
        task("mixed-bulk-search", iterations=10),
    ]

    return {
        "version": 2,
        "description": "Kubebench generated ESRally workload",
        "indices": [{"name": "{{target_index}}", "body": index_body_for_profile()}],
        "corpora": [{"name": "kubebench-generated", "documents": [document_entry]}],
        "operations": operations,
        "challenges": [
            {"name": "index", "schedule": index_tasks},
            {"name": "search", "schedule": search_tasks},
            {"name": "mixed", "schedule": mixed_tasks},
            {"name": "all", "default": True, "schedule": all_tasks},
        ],
    }


print(
    "Generating ESRally track: "
    f"profile={profile} workload={workload} documentCount={document_count} "
    f"targetVersion={target_version or 'unspecified'} index={index_name}"
)
uncompressed_bytes = write_documents()
track = build_track(uncompressed_bytes)
track_file = track_path / "track.json"
track_file.write_text(json.dumps(track, indent=2, sort_keys=True) + "\n", encoding="utf-8")
print(f"Generated Rally track {track_file} and corpus {documents_file}")
