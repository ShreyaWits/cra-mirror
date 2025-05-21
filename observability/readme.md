*2.1. OpenTelemetry Collector (Agent & Gateway)*
Two main deployment modes for the OpenTelemetry Collector:

* **Agent (DaemonSet on each Kubernetes Node):**
    * **Responsibilities:**
        * Collect file-based logs from containers on its node using the `filelog` receiver (configured with include/exclude paths and parsing rules).
        * Receive OTLP telemetry (logs, traces, metrics) directly from applications running on the same node via `localhost:4317` (gRPC) and `localhost:4318` (HTTP).
        * Perform initial processing like batching, resource attribute enrichment (e.g., adding `k8s.pod.name`, `k8s.node.name`).
        * Forward processed data to the Central OTEL Collector Gateway.
    * **Example Agent Pipelines:**
        * **Logs:**
            * Receivers: `otlp` (for app direct OTLP logs), `filelog` (for container stdout/stderr logs).
            * Processors: `memory_limiter`, `batch`, `resource` (for node/pod attributes), `attributes` (for custom tags), `filter` (to drop noisy/unwanted logs).
            * Exporters: `otlphttp` (to Central Gateway).
        * **Traces:**
            * Receivers: `otlp`.
            * Processors: `memory_limiter`, `batch`, `resource`, `attributes` (e.g., for sampling decisions if not done at SDK).
            * Exporters: `otlphttp` (to Central Gateway).
        * **Metrics:**
            * Receivers: `otlp` (for app custom metrics), `hostmetrics` (for node-level OS metrics), `kubeletstats` (for pod/container resource usage from kubelet).
            * Processors: `memory_limiter`, `batch`, `resource`.
            * Exporters: `otlphttp` (to Central Gateway).

* **Central OTEL Collector Gateway (Deployment with HPA):**
    * **Responsibilities:**
        * Receive aggregated telemetry data from all node agents.
        * Perform centralized processing:
            * Final batching for efficiency.
            * Potentially global sampling for traces.
            * Attribute manipulation/addition (e.g., cluster name).
            * Routing data to appropriate backend storage systems.
        * Provide a stable endpoint for agents, abstracting backend details.
    * **Example Gateway Pipelines:**
        * **Logs:**
            * Receivers: `otlphttp` (from Agents).
            * Processors: `memory_limiter`, `batch`.
            * Exporters: `loki` (configured with Loki endpoint, tenant ID if multi-tenancy is used).
        * **Traces:**
            * Receivers: `otlphttp` (from Agents).
            * Processors: `memory_limiter`, `batch`, `spanmetrics` (to generate metrics from spans like RED).
            * Exporters: `otlphttp` (to Tempo, using gRPC: `tempo:4317`).
        * **Metrics:**
            * Receivers: `otlphttp` (from Agents).
            * Processors: `memory_limiter`, `batch`.
            * Exporters: `prometheusremotewrite` (to Mimir endpoint).




*2.2. Loki (Logs)*
* **Deployment:** Deployed via its official Helm chart on Kubernetes, configured for microservices mode (distributor, ingester, querier, query-frontend components).
* **Ingestion:** Receives logs from the Central OTEL Collector Gateway via the Loki exporter (HTTP push).
* **Storage Backend (Chunks):** MinIO. Logs are batched into chunks and stored as objects.
    * Schema: Configured for a time-series based schema (e.g., v11, v12) for efficient querying.
* **Index Backend:** Cassandra. Stores an index of log labels (metadata like service name, pod name, trace ID) to quickly find relevant log chunks in MinIO.
* **Query Language:** LogQL.
* **Retention:**
    * **Primary (Hot Storage - MinIO/Cassandra):** 90 days. Data is queryable directly via Grafana.
    * **Archival (Cold Storage - MinIO Archive Tier):** Logs older than 90 days are moved to a separate, lower-cost archive tier in MinIO using MinIO's lifecycle policies. Access to archived logs might require a restoration process or a specialized Grafana plugin/datasource that can query the archive tier (potentially slower).
* **Multi-tenancy:** Can be enabled via `X-Scope-OrgID` header, managed by the OTEL Collector Gateway.

*2.3. Tempo (Traces)*
* **Deployment Mode:** Can be deployed in monolithic mode for simplicity or microservices mode for larger scale, using its official Helm chart.
* **Ingestion:** Receives traces via OTLP (gRPC/HTTP) from the Central OTEL Collector Gateway.
* **Storage Backend:** MinIO. Traces are stored as objects.
* **Indexing:** Primarily relies on search by Trace ID. For broader search capabilities, it can be integrated with a separate search backend (e.g., Elasticsearch, or relies on metrics generated from spans by `spanmetrics` processor for discovering traces). The LLD does not specify Cassandra for Tempo indexing; Tempo's design often minimizes indexing dependencies for cost and operational simplicity, relying on Trace ID lookups.
* **Query Language:** TraceQL.
* **Retention:** 90+ days in MinIO. Archival can be managed by MinIO lifecycle policies similar to Loki.
* **Span Correlation:** Achieved via `trace_id` linking all spans in a request, and `span_id` / `parent_span_id` defining the causal relationships within a trace.
* ***`x-track-` headers:** This refers to custom HTTP headers (e.g., `x-track-operation-id`, `x-track-user-id`) that applications might propagate. If these are included as attributes on spans by the OTEL SDKs, they become searchable tags in Tempo, enhancing traceability for specific business operations or users.

*2.4. Mimir (Metrics)*
* **Deployment:** Deployed via its official Helm chart, leveraging its Cortex-style microservices architecture (distributor, ingester, querier, store-gateway, compactor, ruler, alertmanager).
* **Ingestion:** Receives metrics via Prometheus Remote Write protocol from the Central OTEL Collector Gateway. OTEL Collector can also scrape Prometheus exporters and convert to OTLP if needed, then the gateway converts back to remote write for Mimir.
* **Storage Backend:** MinIO for long-term storage of metric blocks.
* **Query Language:** PromQL.
* **Scaling:** Horizontally scalable components. Ingesters are stateful but can be scaled; other components are mostly stateless.
* **Retention:** 90 days for queryable metrics in primary storage. Older data can be archived or downsampled. (Aligned with PRD).
* **High Availability:** Achieved through replication of data blocks and multiple replicas of each component.
* **Multi-tenancy:** Native support via `X-Scope-OrgID` header.

*2.5. Grafana (Visualization & Alerting)*
* **Deployment:** Deployed as a Kubernetes Deployment with persistent storage for its configuration database (e.g., PostgreSQL).
* **Dashboards:**
    * **Unified View:** Pre-built and customizable dashboards showing key RED metrics (Rate, Errors, Duration), resource usage (CPU/memory/network) per service.
    * **Log Exploration:** Connects to Loki data source for searching, filtering, and live-tailing logs.
    * **Trace Visualization:** Connects to Tempo data source for searching traces by ID, visualizing trace flame graphs/Gantt charts, and examining span details.
    * **Metric Analysis:** Connects to Mimir data source for PromQL queries, creating graphs, heatmaps, and stat panels.
* **Features:**
    * Dark/light mode toggle for user preference.
    * **Drill-down Capabilities:**
        * From a metric anomaly on a dashboard (e.g., high latency in Mimir) → link to relevant traces in Tempo for that service and time range.
        * From a span in Tempo → link to logs in Loki that share the same `trace_id` and approximate timestamp.
        * Log lines with `trace_id` → link to the full trace in Tempo.
    * **Alerting:** Utilizes Grafana Alerting (unified alerting system).
        * Alert rules defined using PromQL (for Mimir metrics) or LogQL (for Loki logs).
        * Notification channels: Email, Slack, PagerDuty, Webhooks.
        * Integration with Mimir's ruler component is also possible for distributed alert evaluation.

*2.6. MinIO (Blob Storage)*
* **Role:** Primary and archive storage for Loki (log chunks), Tempo (trace objects), and Mimir (metric blocks).
* **Deployment:** Deployed as a distributed, S3-compatible object storage system on Kubernetes.
* **Tiering:** Configured with lifecycle policies to automatically transition older data from "hot" (frequent access) storage classes/buckets to "cold" (archive, infrequent access) storage classes/buckets to optimize costs.
* **Access:** Accessed by Loki, Tempo, and Mimir components using S3 SDKs.

*2.7. Cassandra (Blob Index Storage)*
* **Role:** Stores the index for Loki, mapping log stream labels to MinIO object locations. This enables fast lookups of logs based on metadata.
* **Deployment:** Deployed as a stateful cluster on Kubernetes.
* **Data Model:** Loki uses Cassandra to store tables for index entries (e.g., label sets to chunk IDs).
* **Scalability & Availability:** Cassandra's distributed nature provides horizontal scalability and fault tolerance for the index.


*5.1. Trace Emitter (Go)*
```go
package otelutils

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0" // Use latest stable
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // For demo; use secure in prod
)

// InitTracer initializes and registers a global TracerProvider.
// collectorEndpoint is like "otel-collector-agent.observability.svc.cluster.local:4317" or "localhost:4317" for local agent.
func InitTracer(ctx context.Context, serviceName string, serviceVersion string, environment string, collectorEndpoint string) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			semconv.DeploymentEnvironment(environment),
		),
		resource.WithSchemaURL(semconv.SchemaURL),
	)
	if err != nil {
		return nil, log.Fatalf("failed to create resource: %v", err)
	}

	// For demo purposes, use insecure. In production, configure TLS.
	conn, err := grpc.DialContext(ctx, collectorEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, log.Fatalf("failed to create gRPC connection to collector: %v", err)
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, log.Fatalf("failed to create trace exporter: %v", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch span processor to aggregate spans before exporting.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Or ParentBased(AlwaysSample())
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator to tracecontext and baggage.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// Shutdown function to ensure telemetry is flushed.
	shutdown := func(ctx context.Context) error {
		// Cleanly shutdown and flush telemetry when the application exits.
		if err := tracerProvider.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
			return err
		}
		if err := traceExporter.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down trace exporter: %v", err)
			return err
		}
		if err := conn.Close(); err != nil {
			log.Printf("Error closing gRPC connection: %v", err)
			return err
		}
		return nil
	}

	log.Printf("Tracer initialized for service: %s, sending to %s", serviceName, collectorEndpoint)
	return shutdown, nil
}

// GetTracer returns a tracer for the given service name.
// Call this after InitTracer.
func GetTracer(serviceName string) trace.Tracer {
	return otel.Tracer(serviceName)
}
```

*5.2. Log Emitter (Go) - Using OpenTelemetry Logging SDK (Experimental)*
*Note: The OpenTelemetry Logging SDK is still experimental. For production, consider structured logging libraries (e.g., zap, logrus) that can be configured to output to files tailed by the OTEL Collector's `filelog` receiver, or directly format logs with trace/span IDs for easier correlation if not using direct OTLP log export.*
```go
package otelutils

import (
	"context"
	"log" // Standard log for internal logging of this setup

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc" // Experimental
	"go.opentelemetry.io/otel/log/global"                         // Experimental
	sdklog "go.opentelemetry.io/otel/sdk/log"                     // Experimental
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InitLoggerProvider initializes and registers a global LoggerProvider for OTLP log export.
// collectorEndpoint is like "otel-collector-agent.observability.svc.cluster.local:4317" or "localhost:4317".
func InitLoggerProvider(ctx context.Context, serviceName string, serviceVersion string, environment string, collectorEndpoint string) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			semconv.DeploymentEnvironment(environment),
		),
		resource.WithSchemaURL(semconv.SchemaURL),
	)
	if err != nil {
		return nil, log.Fatalf("failed to create resource for logger: %v", err)
	}

	conn, err := grpc.DialContext(ctx, collectorEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, log.Fatalf("failed to create gRPC connection to collector for logger: %v", err)
	}

	logExporter, err := otlploggrpc.New(ctx, otlploggrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, log.Fatalf("failed to create OTLP log exporter: %v", err)
	}

	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter,
			sdklog.WithBatchTimeout(5*time.Second), // Example: customize batching
		)),
		sdklog.WithResource(res),
	)
	global.SetLoggerProvider(loggerProvider) // Set the global logger provider

	shutdown := func(ctx context.Context) error {
		if err := loggerProvider.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down logger provider: %v", err)
			return err
		}
		if err := logExporter.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down log exporter: %v", err)
			return err
		}
		if err := conn.Close(); err != nil {
			log.Printf("Error closing gRPC connection for logger: %v", err)
			return err
		}
		return nil
	}
	log.Printf("LoggerProvider initialized for service: %s, sending OTLP logs to %s", serviceName, collectorEndpoint)
	return shutdown, nil
}

// GetLogger returns a logger. Call this after InitLoggerProvider.
// The instrumentationName is typically the package or library name.
func GetLogger(instrumentationName string) sdklog.Logger { // Using sdklog.Logger directly
    return global.LoggerProvider().Logger(instrumentationName)
}
```

*5.3. Metrics Emitter (Go)*
```go
package otelutils

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InitMeterProvider initializes and registers a global MeterProvider.
// collectorEndpoint is like "otel-collector-agent.observability.svc.cluster.local:4317" or "localhost:4317".
func InitMeterProvider(ctx context.Context, serviceName string, serviceVersion string, environment string, collectorEndpoint string) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			semconv.DeploymentEnvironment(environment),
		),
		resource.WithSchemaURL(semconv.SchemaURL),
	)
	if err != nil {
		return nil, log.Fatalf("failed to create resource for meter: %v", err)
	}

	conn, err := grpc.DialContext(ctx, collectorEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, log.Fatalf("failed to create gRPC connection to collector for meter: %v", err)
	}

	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, log.Fatalf("failed to create OTLP metric exporter: %v", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(15*time.Second))), // Export interval
	)
	otel.SetMeterProvider(meterProvider)

	shutdown := func(ctx context.Context) error {
		if err := meterProvider.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down meter provider: %v", err)
			return err
		}
		if err := metricExporter.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down metric exporter: %v", err)
			return err
		}
		if err := conn.Close(); err != nil {
			log.Printf("Error closing gRPC connection for meter: %v", err)
			return err
		}
		return nil
	}

	log.Printf("MeterProvider initialized for service: %s, sending to %s", serviceName, collectorEndpoint)
	return shutdown, nil
}

// GetMeter returns a meter for the given service name.
// Call this after InitMeterProvider.
func GetMeter(instrumentationName string) metric.Meter {
	return otel.Meter(instrumentationName)
}
```
6. 🛡️ Security & Access Control
*6.1. Authentication & Authorization*
* **Grafana:**
    * **Authentication:** OAuth (GitHub/Okta preferred for centralized user management), SAML, or built-in Grafana user management.
    * **Authorization:** Role-Based Access Control (RBAC) with predefined roles (Admin, Editor, Viewer) and custom roles. Permissions assigned at Organization, Team, Folder, and Dashboard levels.
* **Backend APIs (Loki, Tempo, Mimir, OTEL Collector Gateway):**
    * Typically secured via mTLS for inter-component communication within the cluster.
    * If exposed externally (e.g., for direct agent push from outside Kubernetes), API gateways (like Istio Ingress Gateway, Nginx) with OIDC/API Key authentication should be used.
    * Tenant access (for Mimir/Loki) controlled by `X-Scope-OrgID` header, which should be injected by an authenticated gateway or the OTEL Collector based on the source of telemetry.
* **MinIO & Cassandra:**
    * Access controlled by credentials (access key/secret key for MinIO, username/password for Cassandra). These secrets are managed via Kubernetes Secrets and mounted into respective backend components.
    * Network policies restrict access to these databases only from necessary components (Loki, Tempo, Mimir).

*6.2. Network Security*
* **TLS Encryption:**
    * All gRPC and HTTP communication between OTEL SDKs, Collectors, Backends, and Grafana UI access should be over TLS.
    * Use Kubernetes `cert-manager` for automated certificate provisioning and lifecycle management for internal endpoints.
* **Kubernetes Network Policies:**
    * Strict network policies applied to limit pod-to-pod communication. E.g., only OTEL Collector Gateway can talk to Loki/Tempo/Mimir distributors; only Grafana can talk to backend queriers.
    * Default deny policies with explicit allow rules for required flows.
* **Bastion Host / VPN:** Access to Cassandra or MinIO for administrative purposes should be through a bastion host or VPN.

*6.3. Secrets Management*
* All sensitive information (API keys, database passwords, TLS certificates/keys, Grafana admin password) stored as Kubernetes Secrets.
* Use tools like HashiCorp Vault or Sealed Secrets for more advanced secrets management and encryption at rest for secrets in Git.
* RBAC for Kubernetes Secrets to restrict access.


7. 📊 Monitoring, Alerts & SLOs
*7.1. Key Metrics to Monitor (for the Observability Service itself)*
* **OTEL Collector:** `otelcol_receiver_accepted_spans/log_records/metric_points`, `otelcol_exporter_sent_spans/etc.`, queue lengths, CPU/memory usage.
* **Loki:** Ingestion rate, query latency, error rates (distributor, ingester, querier), cache hit rates, Cassandra/MinIO operation latencies.
* **Tempo:** Ingestion rate, query latency (TraceID lookup), error rates, MinIO operation latencies.
* **Mimir:** Ingestion rate (samples/sec), active series count, query latency, error rates for all components, rule evaluation latency, alertmanager notification success/failure.
* **Grafana:** Dashboard load times, query proxy latency, login rates, internal error rates.
* **Cassandra:** Read/write latencies, pending compactions, disk usage, node health.
* **MinIO:** Throughput, latency, storage capacity, error rates.

*7.2. Alerting Strategy*
* Alerts defined in Grafana Alerting using PromQL (for Mimir/Collector metrics) or LogQL (for Loki logs).
* **Example Alert Types:**
    * **No data from service:** `absent(up{job="my-service"} == 1)` for a certain period.
    * **High Error Rate:** `sum(rate(http_server_requests_seconds_count{status_code=~"5.."}[5m])) / sum(rate(http_server_requests_seconds_count[5m])) > 0.05`
    * **High Latency:** `histogram_quantile(0.99, sum(rate(http_server_requests_seconds_bucket[5m])) by (le, job)) > 1` (1 second p99 latency).
    * **Observability Pipeline Issues:**
        * OTEL Collector queue overflowing: `otelcol_exporter_queue_capacity > 0 and otelcol_exporter_queue_size / otelcol_exporter_queue_capacity > 0.8`
        * Loki/Mimir ingestion falling behind: Monitor specific lag metrics if available, or alert on persistent high error rates from exporters.
        * Backend component down: `up{job="loki-ingester"} == 0`
    * **Resource Saturation:** High CPU/memory on collector/backend nodes.
* **Notification Channels:** Webhook (Slack, PagerDuty), Email.
* **Alert Deduplication/Grouping:** Handled by Grafana Alerting or an external Alertmanager if Mimir's ruler is used with one.

*7.3. Service Level Objectives (SLOs)*
* **Data Ingestion Latency (P95):** < 5 seconds from OTEL SDK export to queryable in backend.
* **Query Latency (P95):**
    * Loki (typical log query): < 3 seconds.
    * Tempo (TraceID lookup): < 2 seconds.
    * Mimir (typical dashboard panel query): < 2 seconds.
* **Platform Uptime:** ≥ 99.9% for critical path (collection, storage, querying via Grafana).
* **Data Completeness (Logs/Traces/Metrics):** > 99.95% of successfully generated telemetry should be ingested.


8. 📈 Scalability & Performance
*8.1. Component Scaling Strategies*
* **OTEL Collector Agent (DaemonSet):** Scales with the number of nodes. Resource requests/limits tuned per node capacity.
* **OTEL Collector Gateway (Deployment):** Horizontal Pod Autoscaler (HPA) based on CPU/memory usage or custom metrics (e.g., queue length, incoming OTLP requests/sec).
* **Loki:**
    * Stateless components (distributor, querier, query-frontend) scaled via HPA.
    * Stateful components (ingester) scaled by increasing replicas; Loki handles sharding and replication.
* **Tempo:**
    * Monolithic mode: Vertical scaling or increase replicas if stateless parts allow.
    * Microservices mode: Individual components (distributor, ingester, querier, compactor) scaled via HPA or replica count. Sharding by Trace ID is inherent.
* **Mimir:** All components are designed for horizontal scaling. Ingesters are stateful but sharded. Distributors, queriers, store-gateways, rulers, alertmanagers are typically stateless and scaled with HPA.
* **Grafana:** Stateless; scale by increasing replicas behind a load balancer.
* **Cassandra:** Horizontal scaling by adding more nodes to the ring. Data is rebalanced automatically.
* **MinIO:** Distributed mode allows scaling by adding more servers/disks. Performance scales with the number of drives and nodes.

*8.2. Performance Considerations*
* **Batching:** Crucial at SDK, Agent, and Gateway levels to reduce network overhead and improve ingestion throughput.
* **Sampling (Traces):** Implement head-based or tail-based sampling strategies to manage trace volume without losing critical insights, especially for high-traffic services.
* **Cardinality Management (Metrics):** Be mindful of high cardinality labels in Mimir, as they significantly impact storage and query performance. Use OTEL processors to remove or aggregate unnecessary labels.
* **Efficient Queries:** Optimize LogQL, PromQL, and TraceQL queries. Use selectors effectively, avoid querying unbounded time ranges.
* **Caching:** Leverage caching mechanisms in Grafana, Loki, Mimir (query frontends, store gateways) to speed up repeated queries.
* **Resource Allocation:** Properly size CPU, memory, and disk I/O for all components based on expected load.

10. 🗄️ Data Retention, Archival & Backup
*10.1. Data Retention Policies*
* **Loki (Logs):** 90 days in primary, queryable storage (MinIO hot tier + Cassandra index).
* **Tempo (Traces):** 90 days in primary, queryable storage (MinIO hot tier).
* **Mimir (Metrics):** 90 days in primary, queryable storage (MinIO hot tier).
* These retention periods are enforced by the respective backend systems' configuration (e.g., Loki's `retention_period`, Mimir's compactor settings) and MinIO lifecycle policies.

*10.2. Data Archival Process*
* **Mechanism:** MinIO server-side object lifecycle policies.
* **Process:** Objects (log chunks, trace data, metric blocks) in primary storage buckets are automatically transitioned to a designated "archive" bucket or storage class after their primary retention period (e.g., 90 days).
* **Archive Storage:** The archive bucket may use a lower-cost, higher-latency storage class within MinIO.
* **Accessing Archived Data:**
    * Loki: May require specific configuration or a "search-shipper" like tool to re-ingest or query from archive. Direct querying of archived data in Grafana might be limited or slower.
    * Tempo/Mimir: Similar challenges; rehydration might be necessary for deep analysis.
    * "Archive logs not instantly queryable in Grafana" is a known limitation. Solutions might involve specific Grafana plugins for tiered storage or a manual rehydration process.

*10.3. Backup and Restore*
* **Grafana Configuration DB:** Regular backups (e.g., daily `pg_dump` if using PostgreSQL) to a separate location. Test restore procedures.
* **OTEL Collector Configurations:** Stored in Git as code (Config-as-Code).
* **MinIO Data:** MinIO supports server-side bucket replication (`mc mirror`) to another MinIO instance or cloud storage for DR. Snapshots can also be considered.
* **Cassandra Data (Loki Index):** Regular snapshots of Cassandra keyspaces. Test restore procedures. Given that the index can often be rebuilt from data in MinIO (though time-consuming for Loki), the primary focus for DR might be on MinIO data.
* **Disaster Recovery (DR) Strategy:**
    * Multi-AZ deployment for critical stateful components (MinIO, Cassandra, Mimir/Loki storage gateways/ingesters where applicable).
    * Defined RPO (Recovery Point Objective) and RTO (Recovery Time Objective) for the observability platform.


16. 📝 Data Formatting and Schemas
The Observability Service relies heavily on OpenTelemetry standards for data formats and schemas.
*16.1. Logs*
* **Protocol:** OTLP (OpenTelemetry Protocol) for transport.
* **Data Model:** Adheres to the [OpenTelemetry Log Data Model](https://opentelemetry.io/docs/specs/otel/logs/data-model/).
* **Key Fields (within OTLP `LogRecord`):**
    * `Timestamp`: Time the event occurred.
    * `ObservedTimestamp`: Time the event was observed by the collector.
    * `SeverityText`: e.g., "INFO", "WARN", "ERROR".
    * `SeverityNumber`: e.g., 9 for INFO, 13 for WARN, 17 for ERROR.
    * `Body`: The log message payload (string, JSON, etc.).
    * `Attributes`: Key-value pairs for structured metadata (e.g., `service.name`, `http.method`, custom tags).
    * `TraceId`, `SpanId`: For correlation with traces.
    * `Resource Attributes`: Describe the source of the log (e.g., `service.name`, `k8s.pod.name`).
* **Storage Format (Loki):** Loki stores logs as streams of text. If logs are structured (e.g., JSON) before reaching Loki, Loki can parse and index fields from the JSON payload. Otherwise, it treats the body as a string. Labels for streams are derived from `Resource Attributes` and selected `Attributes`.

*16.2. Traces*
* **Protocol:** OTLP for transport.
* **Data Model:** Adheres to the [OpenTelemetry Trace Data Model](https://opentelemetry.io/docs/specs/otel/trace/api/#span).
* **Key Components:**
    * **Trace:** A collection of Spans representing a single request or workflow. Identified by a `TraceId`.
    * **Span:** Represents a single operation within a trace. Key fields:
        * `TraceId`: Links it to the parent trace.
        * `SpanId`: Unique identifier for the span.
        * `ParentSpanId`: Links to the parent span (if any).
        * `Name`: Human-readable name for the operation (e.g., "HTTP GET /users", "DatabaseQuery").
        * `Kind`: e.g., `SPAN_KIND_SERVER`, `SPAN_KIND_CLIENT`, `SPAN_KIND_INTERNAL`.
        * `StartTimeUnixNano`, `EndTimeUnixNano`.
        * `Attributes`: Key-value pairs (e.g., `http.method`, `db.statement`, `error`).
        * `Events`: Timestamped events with attributes within a span.
        * `Links`: To causally link spans that are not parent/child.
        * `Status`: (Ok, Error, Unset) with an optional error message.
    * `Resource Attributes`: Describe the source of the trace.

*16.3. Metrics*
* **Protocol:** OTLP for transport. Exported to Mimir via Prometheus Remote Write.
* **Data Model:** Adheres to the [OpenTelemetry Metric Data Model](https://opentelemetry.io/docs/specs/otel/metrics/data-model/).
* **Key Instrument Types:**
    * **Counter:** A value that only increases (e.g., `http.server.request_count`). OTLP type: Sum (monotonic).
    * **UpDownCounter:** A value that can increase or decrease (e.g., `active_users`). OTLP type: Sum (non-monotonic).
    * **Histogram:** Records a distribution of values (e.g., `http.server.duration`). OTLP type: Histogram.
    * **Gauge:** Represents a value at a point in time (e.g., `cpu_utilization`). OTLP type: Gauge.
    * **(Observable Instruments):** Asynchronous versions of the above, where a callback provides the current value.
* **Key Fields (within OTLP `Metric` and `DataPoint`):**
    * `Name`: e.g., `http.server.duration`.
    * `Description`, `Unit`.
    * `DataPoints`: Contain the actual values, timestamps, and attributes (labels).
    * `Attributes`: Key-value pairs for dimensions (e.g., `http.method`, `service.name`).
    * `Resource Attributes`: Describe the source of the metric.
* **Prometheus Exposition Format Compatibility:** Mimir ingests data in Prometheus format. OTEL Collector's `prometheusremotewrite` exporter handles the conversion from OTLP metric model to this format.

*16.4. Semantic Conventions*
* Adherence to OpenTelemetry Semantic Conventions (e.g., `semconv.v1.24.0` or latest) is crucial for standardized attribute keys and meanings across all signals (logs, traces, metrics). This ensures:
    * Consistent tagging (e.g., `service.name`, `http.method`, `http.status_code`, `error.type`).
    * Improved correlation and interoperability between signals.
    * Better out-of-the-box experiences with UIs like Grafana that understand these conventions.
* Examples:
    * `semconv.ServiceNameKey` for `service.name`
    * `semconv.ServiceVersionKey` for `service.version`
    * `semconv.DeploymentEnvironmentKey` for `deployment.environment`
    * HTTP: `semconv.HTTPMethodKey`, `semconv.HTTPRouteKey`, `semconv.HTTPStatusCodeKey`.
    * Database: `semconv.DBSystemKey`, `semconv.DBStatementKey`.
    * Messaging: `semconv.MessagingSystemKey`, `semconv.MessagingDestinationNameKey`.
* These conventions should be applied consistently within the OTEL SDKs and potentially enriched/enforced by OTEL Collector processors.


