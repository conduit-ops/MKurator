resource "helm_release" "ingress_nginx" {
  name             = "ingress-nginx"
  namespace        = "ingress-nginx"
  create_namespace = true

  repository = "https://kubernetes.github.io/ingress-nginx"
  chart      = "ingress-nginx"
  version    = "4.11.3"

  wait    = true
  timeout = 600

  # NodePort 30080/30443 are mapped to the host by kind/cluster.yaml, so
  # ingresses are reachable at http(s)://*.localhost:3008x from the host.
  values = [
    yamlencode({
      controller = {
        service = {
          type = "NodePort"
          nodePorts = {
            http  = 30080
            https = 30443
          }
        }
        ingressClassResource = {
          name    = var.ingress_class_name
          default = true
        }
        # kind has a single node; let the admission webhook settle quickly.
        admissionWebhooks = {
          enabled = true
        }
      }
    })
  ]
}

# TLS secrets are namespace-scoped; place the mkcert material where ingresses
# that reference it live (the ingress controller namespace).
resource "kubernetes_secret_v1" "ingress_tls" {
  metadata {
    name      = var.tls_secret_name
    namespace = "ingress-nginx"
  }

  type = "kubernetes.io/tls"

  data = {
    "tls.crt" = base64decode(var.tls_cert_string)
    "tls.key" = base64decode(var.tls_key_string)
  }

  depends_on = [helm_release.ingress_nginx]
}
