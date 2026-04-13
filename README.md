# PostgreSQL with Syntasso Kratix Enterprise

This Promise is an enhanced version of the [Open Source PostgreSQL
Promise](https://github.com/syntasso/promise-postgresql/tree/main/internal)
using extended functionality from Syntasso Kratix Enterprise (SKE).

This Promise shows:

- How to log and what to log in Promise workflows
- How to use the [SKE Health Agent](https://docs.kratix.io/ske/installing-ske/ske-health-agent) to report Promise status back to SKE
- How to enable the Promise to be used by the Backstage and Cortex integrations
- How to use the [Promise Testing Framework](https://docs.kratix.io/ske/promise-testing-framework/intro) to test
  your Promise code without the need of a live Kubernetes cluster

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/promise-postgresql-ske/refs/heads/main/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/promise-postgresql-ske/refs/heads/main/resource-request.yaml
```

## Development

For development see [README.md](./workflows/configure-pipeline/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you
run into issues or have ideas for us, please let us know. Feel free to [open an
issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or
[put time on our calendar](https://www.syntasso.io/contact-us). We'd love to
hear from you.
