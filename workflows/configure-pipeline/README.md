# Development

## Dependencies

In this Promise we bundle the Zalando Operator CRDs and a reference manifest
used to source defaults for the CRDs rendered by the Promise code. 

Bundling dependency manifests into a Promise is a best practice that helps
mitigate against runtime access issues to these manifests, ensuring they are
always available to the Promise. We provide a convenience script to update
these files from the canonical source.

Be aware that if you run your cluster in an [air-gapped environment](https://docs.kratix.io/ske/installing-ske/air-gapped)
or use a [private registry](https://docs.kratix.io/main/platform-concepts/private-image-registries),
you will need to update the dependency manifests to point to these images.

```shell
# Fetch dependencies
./scripts/fetch-deps
# Fetch reference manifests
./scripts/fetch-pipeline-resources
```

## Pipeline image

To build the image:

```shell
make build
```

To load the image to the local kind platform cluster:

```shell
make load
```

To push the image to ghcr.io:

```shell
make push
```

## Testing

The test suite uses [Ginkgo](https://onsi.github.io/ginkgo/). To run it, install Kratix first (see the
[quickstart](https://docs.kratix.io/main/guides/installing-kratix)),
then:

```shell
make test
```

The tests apply `promise.yaml` and `resource-request.yaml` to the platform
cluster and assert the expected state on the worker cluster. The following
environment variables can be overridden:

| Variable | Default | Description |
|---|---|---|
| `WORKER_CONTEXT` | `kind-worker` | kubeconfig context for the worker cluster |
| `PROMISE_YAML` | `../../promise.yaml` | Path to the promise manifest |
| `RESOURCE_REQUEST_YAML` | `../../resource-request.yaml` | Path to the resource request manifest |
