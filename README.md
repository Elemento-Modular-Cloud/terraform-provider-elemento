# Terraform Provider for Elemento Modular Cloud

**One provider. Every connected cloud. The destination is an argument.**

This is the official Terraform provider for [Elemento Modular Cloud](https://elemento.cloud). It gives infrastructure teams a single Terraform language for compute, storage, databases, Kubernetes, and packaged services — and lets the target cloud be selected in configuration, not by installing a second provider.

The industry default is *many clouds, many providers*: a new plugin, a new schema, and a new mental model for every destination. This repository inverts that. You keep one resource catalog. You change the cloud.

```hcl
provider "elemento" {
  target   = "azure-private" # the target cloud
  username = var.elemento_username
  password = var.elemento_password
}
```

That `target` is the argument. It addresses the [Atomosphere](https://elemento.cloud/atomosphere.html) (Meson) endpoint for the backend you intend to use. The same `elemento_compute`, `elemento_dbaas`, or `elemento_kaas` resource is valid on Azure, UpCloud, AWS, and every other connected cloud. Capability coverage follows Atomosphere; the Terraform schema does not fork.

---

## Why this exists

Terraform already knows how to talk to many clouds. The cost is that each cloud is a different product:

| Traditional multi-cloud | Elemento |
| --- | --- |
| One provider plugin per cloud | One provider for all connected clouds |
| Different resource types and arguments per vendor | One resource language, reused everywhere |
| Porting a stack means rewriting HCL | Porting a stack means changing `host` (or an alias) |
| Teams learn AWS, then Azure, then GCP | Teams learn Elemento once |

Vendor lock-in is often described as a billing or API problem. In Terraform it is also a **schema** problem. Once a module is written against `aws_instance`, `azurerm_linux_virtual_machine`, or `google_compute_instance`, moving it is a rewrite. Elemento exists so that the rewrite is unnecessary.

The provider does not embed AWS, Azure, or GCP SDKs. It is a Terraform Plugin Framework client for Elemento's unified API. Translation to the destination cloud is performed by Atomosphere, Elemento's federated multicloud gateway. That is a deliberate design: new clouds appear as new `host` values, not as new providers and not as new resource types.

---

## The approach

### Many clouds, one provider

```
Traditional Terraform                         Elemento

hashicorp/aws      -->  AWS                   elemento  --host=aws...      -->  AWS
hashicorp/azurerm  -->  Azure                 elemento  --host=azure...    -->  Azure
hashicorp/google   -->  GCP                   elemento  --host=google...   -->  GCP
upcloud/upcloud    -->  UpCloud               elemento  --host=upcloud...  -->  UpCloud
```

A practitioner does not choose a provider to choose a cloud. They choose a **host**. The host is the Meson / Atomosphere instance wired to that backend:

| Target cloud | Typical `host` |
| --- | --- |
| Azure | `azure.dev.meson.elemento.cloud` |
| UpCloud | `upcloud.dev.meson.elemento.cloud` |

The hostname prefix is the cloud. Region remains a resource argument (`region = "italynorth"`, `region = "fi-hel2"`), scoped to the backend selected by `host`.

To use more than one cloud in the same root module, declare **aliases**. You do not add another provider requirement:

```hcl
provider "elemento" {
  alias    = "azure"
  host     = "azure.dev.meson.elemento.cloud"
  username = var.elemento_username
  password = var.elemento_password
}

provider "elemento" {
  alias    = "upcloud"
  host     = "upcloud.dev.meson.elemento.cloud"
  username = var.elemento_username
  password = var.elemento_password
}

resource "elemento_compute" "app_azure" {
  provider = elemento.azure
  region   = "italynorth"
  # same schema as every other cloud
}

resource "elemento_compute" "app_upcloud" {
  provider = elemento.upcloud
  region   = "fi-hel2"
  # same schema as every other cloud
}
```

### Where the abstraction lives

```
Terraform CLI
    │  plugin protocol v6
    ▼
elemento provider          Terraform Plugin Framework
    │  HTTPS
    ▼
Elemento Portal            identity, account, billing status
    │
    ▼
Atomosphere / Meson        unified API → destination CSP
    │
    ├── Azure
    ├── UpCloud
    ├── AWS
    ├── Google Cloud
    └── other connected clouds
```

Two endpoints, one workflow:

- **Portal** (`portal`, default `portal.elemento.cloud`) authenticates the account and reports billing status.
- **Host** (`host`) is the Atomosphere / Meson control plane for the chosen cloud. Resource create, read, and delete go here.

Because the provider speaks only Elemento's API, this repository has no per-cloud adapter packages and no vendor SDKs. Adding a cloud is an Atomosphere integration. The Terraform surface stays stable.

### What stays constant

Across backends, practitioners keep the same:

- Provider configuration shape (`host`, `username`, `password`, `portal`)
- Resource types and argument names
- Billing model (`billing_frequency`, `payment_url`, `billing_uuid`)
- Apply lifecycle (preflight, create, status, credentials)

What changes with the cloud is the `host` value and the region identifier that cloud understands.

---

## What you can provision

| Resource | Type | Purpose |
| --- | --- | --- |
| [Compute](docs/resources/compute.md) | `elemento_compute` | Full-fidelity virtual machines: CPU, memory, PCI, OS family, cloud-init, volumes |
| [Block storage](docs/resources/storage.md) | `elemento_storage` | Standalone volumes (bootable, shareable, and related flags) |
| [Object storage](docs/resources/objectstorage.md) | `elemento_objectstorage` | Managed object / blob buckets |
| [DBaaS](docs/resources/dbaas.md) | `elemento_dbaas` | Managed databases (for example PostgreSQL and MySQL), with connection details on read |
| [Kubernetes](docs/resources/kaas.md) | `elemento_kaas` | Managed clusters, node pools, and kubeconfig |
| [n8n](docs/resources/n8n.md) | `elemento_n8n` | Opinionated n8n service on compute |
| [OpenClaw](docs/resources/openclaw.md) | `elemento_openclaw` | Opinionated OpenClaw service on compute |

### Data sources

Each managed type has a matching singular data source for infrastructure that already exists on the Meson host selected by `host`. Set exactly one of `uuid` (or `volume_uuid` for storage), `billing_uuid`, or `name` / `vm_name`. Names are not unique: multiple matches return an error.

| Data source | Looks up |
| --- | --- |
| [Compute](docs/data-sources/compute.md) | `data.elemento_compute` |
| [Block storage](docs/data-sources/storage.md) | `data.elemento_storage` |
| [Object storage](docs/data-sources/objectstorage.md) | `data.elemento_objectstorage` |
| [DBaaS](docs/data-sources/dbaas.md) | `data.elemento_dbaas` |
| [Kubernetes](docs/data-sources/kaas.md) | `data.elemento_kaas` |

```hcl
data "elemento_dbaas" "app" {
  uuid = var.existing_db_uuid
}

output "db_host" {
  value = data.elemento_dbaas.app.database_host
}
```

Schema reference for every type lives under [`docs/`](docs/). Working configurations live under [`examples/`](examples/).

Not every Atomosphere-connected cloud exposes every service. Terraform resource coverage follows the [Atomosphere integration matrix](https://elemento.cloud/atomosphere.html). The resource *types* do not change when a capability is added on a new backend.

---

## Supported clouds

The Terraform provider supports every backend exposed through Atomosphere/Meson. Service availability varies by backend.

See the live [Atomosphere compatibility matrix](https://elemento.cloud/atomosphere.html).

The provider is cloud-agnostic by construction: a new backend is a new `host`, not a new Terraform provider. Examples in this repository currently demonstrate Azure (`azure.dev.meson.elemento.cloud`) and UpCloud (`upcloud.dev.meson.elemento.cloud`).

---

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/install) 1.x
- An [Elemento](https://elemento.cloud) account
- Network access to the Portal and to the Meson host for the target cloud

---

## Quick start

```hcl
terraform {
  required_providers {
    elemento = {
      source = "Elemento-Modular-Cloud/elemento"
    }
  }
}

provider "elemento" {
  host     = "upcloud.dev.meson.elemento.cloud"
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

output "database_host" {
  value = elemento_dbaas.app.database_host
}

output "payment_url" {
  value = elemento_dbaas.app.payment_url
}
```

Credentials should come from variables or the environment, never from committed files.

```bash
export ELEMENTO_HOST="upcloud.dev.meson.elemento.cloud"
export ELEMENTO_USERNAME="you@example.com"
export ELEMENTO_PASSWORD="..."
terraform init
terraform plan
terraform apply
```

---

## Authentication and configuration

All provider arguments are optional in the schema and may be supplied by environment variables. Configuration values override the environment. After that fallback, `host`, `username`, and `password` are required.

| Argument | Environment variable | Required | Description |
| --- | --- | --- | --- |
| `host` | `ELEMENTO_HOST` | yes | Meson / Atomosphere hostname for the **target cloud** |
| `username` | `ELEMENTO_USERNAME` | yes | Elemento account username |
| `password` | `ELEMENTO_PASSWORD` | yes (sensitive) | Elemento account password |
| `portal` | `ELEMENTO_PORTAL` | no | Portal hostname for auth and billing. Defaults to `portal.elemento.cloud` |

```hcl
provider "elemento" {
  host     = "azure.dev.meson.elemento.cloud"
  username = var.elemento_username
  password = var.elemento_password
  portal   = "portal.elemento.cloud"
}
```

---

## Resource lifecycle

Every create follows the same control-plane path, independent of the destination cloud:

1. **Preflight** — the provider asks Atomosphere whether the request can be fulfilled. Field-level reasons are returned when it cannot.
2. **Register** — a successful create returns a `billing_uuid` and a `payment_url`.
3. **Settle** — Portal billing status is observed (`paid`, `provisioning`, `running`, or a terminal error / cancel / delete).
4. **Materialize** — once the resource exists, Read attaches runtime data: connection strings, kubeconfig, service URL, or object-storage credentials.

`billing_uuid` is the durable handle used to correlate Terraform state with Portal and Meson. If billing reports the resource as deleted, canceled, or in error and the control plane no longer lists it, the provider removes it from state.

---

## Development

This provider is written in Go against [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) (protocol v6).

```bash
make build      # compile
make install    # install into the local Go bin
make generate   # regenerate docs from schema and examples
make test       # unit tests
make testacc    # acceptance tests (requires TF_ACC=1 and live credentials)
make lint       # golangci-lint
```

For a local Terraform CLI override, follow HashiCorp's [plugin-framework local install guide](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#prepare-terraform-for-local-provider-install).

Documentation under `docs/` is generated by `tfplugindocs`. After changing schema descriptions or examples, run `make generate` and commit the result.

---

## License

Mozilla Public License 2.0. See [LICENSE](LICENSE).

## Links

- [Elemento](https://elemento.cloud)
- [Atomosphere — federated multicloud gateway](https://elemento.cloud/atomosphere.html)
- [Provider schema](docs/index.md)
- [Examples](examples/)
- [Changelog](CHANGELOG.md)
- [GitHub](https://github.com/Elemento-Modular-Cloud/terraform-provider-elemento)
