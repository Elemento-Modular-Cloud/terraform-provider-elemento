# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    elemento = {
      source = "hashicorp/elemento"
    }
  }
}

provider "elemento" {
  target   = "aws-dev-private"
  username = "user@elemento.cloud"
  password = "password"
}

data "elemento_spotvm" "n8n_cloudinit_test" {
  uuid = "00000000-0000-0000-0000-000000000000"
}
