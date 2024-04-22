## Terraform Provider for Qernal cloud

## This system includes 3 main component: 

1. Qernal terraform provider: This is the custom terraform provider for Qernal cloud, also act a OAuth client and will interact with the internal Qernal API 
2. Ory Hydra: OAuth server that will handle OAuth flows from OAuth client(Terraform provider)
3. Qernal API Server: The internal API server for Qernal cloud. Our terraform provider will interact with this server.




### Developing Locally 


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


