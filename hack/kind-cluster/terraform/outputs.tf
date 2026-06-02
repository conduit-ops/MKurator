output "ingress_class_name" {
  value       = var.ingress_class_name
  description = "IngressClass name used by ingress objects."
}

output "tls_secret_ref" {
  value       = "${var.tls_namespace}/${var.tls_secret_name}"
  description = "Namespace/name of the TLS secret intended for ingresses."
}

output "grafana_url" {
  value       = var.enable_monitoring ? "https://${local.grafana_host}:30443/" : "monitoring disabled"
  description = "Grafana UI URL (via ingress-nginx NodePort)."
}

output "grafana_admin_user" {
  value       = var.enable_monitoring ? var.grafana_admin_user : "monitoring disabled"
  description = "Grafana admin username."
}

output "grafana_admin_password" {
  value       = var.enable_monitoring ? nonsensitive(var.grafana_admin_password) : "monitoring disabled"
  description = "Grafana admin password."
}

output "mq_web_console_url" {
  value       = "https://${local.mq_host}:30443/ibm/mq/console/"
  description = "IBM MQ web console URL (via ingress-nginx NodePort)."
}

output "mq_rest_admin_url" {
  value       = "https://${local.mq_host}:30443/ibm/mq/rest/v2/admin/qmgr"
  description = "IBM MQ administrative REST API base (via ingress-nginx NodePort)."
}

output "mq_in_cluster_endpoint" {
  value       = "https://${local.mq_release_name}.${var.mq_namespace}.svc:9443"
  description = "In-cluster mqweb endpoint for the operator's QueueManagerConnection.endpoint."
}

output "mq_admin_user" {
  value       = "admin"
  description = "IBM MQ admin user (REST admin API / MQWebAdmin role)."
}

output "mq_admin_password" {
  value       = nonsensitive(var.mq_admin_password)
  description = "IBM MQ admin user password."
}
