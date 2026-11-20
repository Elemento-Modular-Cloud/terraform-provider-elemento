terraform {
  required_providers {
    elemento = {
      source = "Elemento-Modular-Cloud/elemento"
    }
  }
}

# host selects the destination cloud. The resource catalog does not change.
provider "elemento" {
  target   = "azure-dev-private"
  username = var.elemento_username
  password = var.elemento_password
}

variable "elemento_username" {
  type        = string
  description = "Elemento account username"
}

variable "elemento_password" {
  type        = string
  sensitive   = true
  description = "Elemento account password"
}
