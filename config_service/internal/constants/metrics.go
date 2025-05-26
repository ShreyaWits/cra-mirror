package constants

// Admin Handler Metrics Constants
const AdminHandlerRequestCounterMetric = "admin_request_total"
const AdminHandlerRequestLatencyMetric = "admin_request_duration_seconds"
const AdminHandlerErrorCounterMetric = "admin_error_total"

// Config Handler Metrics Constants
const ConfigHandlerRequestCounterMetric = "config_request_total"
const ConfigHandlerRequestLatencyMetric = "config_request_duration_seconds"
const ConfigHandlerErrorCounterMetric = "config_error_total"

const WebhookHandlerRequestCounterMetric = "webhook_request_total"
const WebhookHandlerRequestLatencyMetric = "webhook_request_duration_seconds"
const WebhookHandlerErrorCounterMetric = "webhook_error_total"

const ConfigRepoOperationLatencyMetric = "config_repo.operation_latency"
const ConfigRepoOperationCountMetric = "config_repo.operation_count"
const ConfigRepoErrorCountMetric = "config_repo.error_count"
