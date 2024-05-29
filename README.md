# Terraform Provider for Qernal

The Terraform provider for Qernal allows you to manage resources via
terraform. This provider is maintained by Qernal.

## Quick Start

<!-- TODO insert quick start docs -->

## Provider Usage

<!-- TODO insert provider usage docs -->

## Developing the provider


To test the provider locally, Terraform provides an easy way to specify dev ovverides. Begin retrieving your GOBIN using the `go env` command:

```bash
go env GOBIN
```


Next edit or create file in the following path `~/.terraformrc`

```terraform
provider_installation {

  dev_overrides {
      "qernal/qernal" = "<PATH>"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```
Be sure to replace <PATH> with the output returned from `go env GOBIN`


