variable "kubeconfig" {
  description = "Path to kubeconfig used by Terraform."
  type        = string
}

variable "ingress_class_name" {
  description = "IngressClass name. Must be 'nginx' to match the IBM MQ chart's web console Ingress."
  type        = string
  default     = "nginx"
}

variable "tls_namespace" {
  description = "Namespace to store the shared TLS secret for ingresses."
  type        = string
  default     = "ingress-nginx"
}

variable "tls_secret_name" {
  description = "Secret name for the mkcert TLS material."
  type        = string
  default     = "wildcard-localhost-tls"
}

variable "tls_cert_string" {
  description = "Base64-encoded TLS cert (PEM)."
  type        = string
  default     = ""
}

variable "tls_key_string" {
  description = "Base64-encoded TLS key (PEM)."
  type        = string
  default     = ""
}

variable "grafana_admin_user" {
  description = "Grafana admin username (kube-prometheus-stack)."
  type        = string
  default     = "admin"
}

variable "grafana_admin_password" {
  description = "Grafana admin password (kube-prometheus-stack)."
  type        = string
  default     = "admin"
  sensitive   = true
}

variable "enable_monitoring" {
  description = "Whether to install kube-prometheus-stack (Prometheus + Grafana)."
  type        = bool
  default     = true
}

variable "mq_namespace" {
  description = "Namespace for the IBM MQ queue manager."
  type        = string
  default     = "ibm-mq"
}

variable "mq_queue_manager_name" {
  description = "Name of the IBM MQ queue manager."
  type        = string
  default     = "QM1"
}

variable "mq_admin_password" {
  description = "Password for the MQ 'admin' user (MQWebAdmin role / REST admin API)."
  type        = string
  default     = "passw0rd"
  sensitive   = true
}

variable "mq_app_password" {
  description = "Password for the MQ 'app' user."
  type        = string
  default     = "passw0rd"
  sensitive   = true
}
