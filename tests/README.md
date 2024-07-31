# Qernal Terraform

Before running the tests, you'll need to build and install the provider
which can be done with;

```bash
go mod download
make local-build
make local-install
```

Running the tests;

```bash
go test ./tests
```

However, if you need to supply extra environment vars such as the token or a different
build environment then this can be done like so;

```bash
QERNAL_CHAOS_ENDPOINT=https://chaos.qernal.com QERNAL_HYDRA_ENDPOINT=https://hydra.qernal.com QERNAL_TOKEN=$(cat ./qernal-token) go test ./tests
```