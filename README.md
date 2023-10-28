## Terraform Provider for Qernal cloud

## This system includes 3 main component: 

1. Qernal terraform provider: This is the custom terraform provider for Qernal cloud, also act a OAuth client and will interact with the internal Qernal API 
2. Ory Hydra: OAuth server that will handle OAuth flows from OAuth client(Terraform provider)
3. Qernal API Server: The internal API server for Qernal cloud. Our terraform provider will interact with this server.
