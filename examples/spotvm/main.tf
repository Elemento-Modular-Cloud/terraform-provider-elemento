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

resource "elemento_spotvm" "n8n_cloudinit_test" {
  vm_name = "n8n-cloudinit-test"
  cpu = {
    slots        = 2
    shared_cores = true
    arch         = ["X86_64"]
  }
  mem = {
    capacity_mb = 4096
  }
  os = {
    family  = "linux"
    flavour = "ubuntu"
    version = "24.04"
  }
  storage = {
    boot = {
      name    = "root"
      size_gb = 32
    }
    data = []
  }
  auth = {
    username = "elemento"
    password = "elemento123"
    ssh_key  = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBS/qkkoKOEmkCANzrvPki5t5crf0mwRItAXWnw45R3v"
  }
  tags = ""
  region        = "eu-south-1"
  cloudinit_b64 = "I2Nsb3VkLWNvbmZpZwpwYWNrYWdlczoKICAtIG5naW54CnJ1bmNtZDoKICAtIFsgc3lzdGVtY3RsLCBlbmFibGUsIG5naW54IF0K"
}
