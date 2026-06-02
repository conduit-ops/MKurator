locals {
  mq_host         = "mq.localhost"
  mq_release_name = "ibm-mq"
}

resource "kubernetes_namespace_v1" "ibm_mq" {
  metadata {
    name = var.mq_namespace
  }
}

# Copy the mkcert TLS material into the MQ namespace for the web console Ingress.
resource "kubernetes_secret_v1" "mq_tls" {
  metadata {
    name      = var.tls_secret_name
    namespace = kubernetes_namespace_v1.ibm_mq.metadata[0].name
  }

  type = "kubernetes.io/tls"

  data = {
    "tls.crt" = base64decode(var.tls_cert_string)
    "tls.key" = base64decode(var.tls_key_string)
  }
}

# Passwords for the MQ "admin" (MQWebAdmin / REST admin API) and "app" users.
resource "kubernetes_secret_v1" "mq_credentials" {
  metadata {
    name      = "mq-credentials"
    namespace = kubernetes_namespace_v1.ibm_mq.metadata[0].name
  }

  type = "Opaque"

  data = {
    mqAdminPassword = var.mq_admin_password
    mqAppPassword   = var.mq_app_password
  }
}

resource "helm_release" "ibm_mq" {
  name      = local.mq_release_name
  namespace = kubernetes_namespace_v1.ibm_mq.metadata[0].name

  # Vendored chart (../charts/ibm-mq). The official IBM MQ chart.
  chart = "${path.module}/../charts/ibm-mq"

  wait    = true
  timeout = 900

  values = [
    yamlencode({
      # Accepts the IBM MQ Advanced for Developers license for local dev use.
      license = "accept"

      image = {
        repository = "icr.io/ibm-messaging/mq"
        tag        = "9.4.2.0-r1"
      }

      queueManager = {
        name = var.mq_queue_manager_name
      }

      # Enable the web server (web console + administrative REST API).
      web = {
        enable = true
      }

      credentials = {
        enable = true
        secret = kubernetes_secret_v1.mq_credentials.metadata[0].name
      }

      metrics = {
        enabled = true
      }

      persistence = {
        qmPVC = {
          enable = true
          size   = "2Gi"
        }
      }

      resources = {
        requests = {
          cpu    = "250m"
          memory = "512Mi"
        }
        limits = {
          cpu    = "1"
          memory = "1024Mi"
        }
      }

      # Expose the web console / REST API via ingress-nginx. The backend speaks
      # HTTPS (port console-https/9443), so tell nginx to re-encrypt.
      route = {
        ingress = {
          webconsole = {
            enable   = true
            hostname = local.mq_host
            path     = "/"
            tls = {
              enable = true
              secret = var.tls_secret_name
            }
          }
          annotations = {
            "nginx.ingress.kubernetes.io/backend-protocol" = "HTTPS"
          }
        }
      }
    })
  ]

  depends_on = [
    helm_release.ingress_nginx,
    kubernetes_secret_v1.mq_tls,
    kubernetes_secret_v1.mq_credentials,
  ]
}
