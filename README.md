# Terraform Provider for Elemento Modular Cloud

The Elemento provider manages resources on [Elemento Modular Cloud](https://elemento.cloud) through [Atomosphere](https://elemento.cloud/atomosphere.html), Elemento's multicloud gateway. The same resource types work on every connected cloud, and the `target` argument selects which one.

- [Provider documentation](docs/index.md)
- [Examples](examples/)
- [Changelog](CHANGELOG.md)

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/install) 1.x
- An [Elemento](https://elemento.cloud) account
- [Go](https://go.dev/doc/install) 1.25 or later, only to build the provider

## Usage

```hcl
terraform {
  required_providers {
    elemento = {
      source = "Elemento-Modular-Cloud/elemento"
    }
  }
}

provider "elemento" {
  target   = "upcloud.dev.meson.elemento.cloud"
  username = var.elemento_username
  password = var.elemento_password
}

resource "elemento_dbaas" "app" {
  name              = "app-db"
  region            = "fi-hel2"
  engine            = "postgres"
  disk_size_gb      = 50
  nodes_number      = 1
  backup_time       = "03:00"
  billing_frequency = "month"
  password          = var.db_password
}
```

Provider arguments can also be set with environment variables:

| Argument   | Environment variable | Description                                                    |
| ---------- | -------------------- | -------------------------------------------------------------- |
| `target`   | `ELEMENTO_TARGET`    | Atomosphere endpoint of the cloud to manage                    |
| `username` | `ELEMENTO_USERNAME`  | Account username                                               |
| `password` | `ELEMENTO_PASSWORD`  | Account password                                               |
| `portal`   | `ELEMENTO_PORTAL`    | Portal hostname, defaults to `portal.elemento.cloud`           |

To manage several clouds in one configuration, declare one provider block per cloud with an `alias`.

## Resources

Compute, spot VMs, block, file and object storage, managed databases (DBaaS), Kubernetes (KaaS), queues, registries, serverless functions, AI serverless, load balancers, private networks, public gateways and IPs, n8n and OpenClaw. Most resources have a matching data source. Service availability depends on the target cloud, see the [Atomosphere matrix](https://elemento.cloud/atomosphere.html).

## Development

```bash
make build      # compile
make install    # install into the local Go bin
make generate   # regenerate docs
make test       # unit tests
make testacc    # acceptance tests, needs TF_ACC=1 and credentials
make lint       # golangci-lint
```

## License

[Mozilla Public License 2.0](LICENSE)
