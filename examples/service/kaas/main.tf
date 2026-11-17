# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    elemento = {
      source = "hashicorp/elemento"
    }
  }
}

provider "elemento" {
  target   = "upcloud-dev-private"
  username = "user@elemento.cloud"
  password = "password"
}

resource "elemento_kaas" "my_cluster" {
  name              = "my-cluster"
  region            = "fi-hel2"
  billing_frequency = "day"
  node_pools = [
    {
      name              = "my-node-pool"
      auto_scale        = false
      min_nodes         = 3
      max_nodes         = 3
      shared_storage_gb = 0
      specs = {
        cpu_cores    = 2
        memory_gb    = 4
        disk_size_gb = 64
        gpus         = 0
      }
    }
  ]
  nodes_subnet = {
    dhcp_enabled = true
    cidr         = "192.170.2.0/24"
  }
  update_policy            = "ALWAYS_UPDATE"
  kubernetes_version       = "1.35"
  control_plane_ha_enabled = true
}

output "payment_url" {
  description = "Payment URL for the created resource"
  value       = elemento_kaas.my_cluster.payment_url
}
