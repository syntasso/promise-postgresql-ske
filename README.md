# PostgreSQL

This Promise provides PostgreSQL-as-a-Service. The Promise has four fields:

* `.spec.env`
* `.spec.teamId`
* `.spec.dbName`
* `.spec.backupEnabled`

Check the CRD documentation for more information.


To install:
```
kubectl apply -f https://raw.githubusercontent.com/<owner>/<repo>/main/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/<owner>/<repo>/main/resource-request.yaml
```

## Development

For development see [README.md](./workflows/configure-pipeline/README.md)

## Releasing

Releases are automated via [release-please](https://github.com/googleapis/release-please). Version bumps are determined by [Conventional Commits](https://www.conventionalcommits.org/):

- `fix:` → patch bump
- `feat:` → minor bump
- `feat!:` or `BREAKING CHANGE:` → major bump

This repo is bootstrapped for a new import and currently tracks an unreleased baseline of `0.0.0`.

### First stable release

To force the first release in the new repository to `v1.0.0`, merge a commit to `main` with `Release-As: 1.0.0` in the commit body. Example:

```text
chore: release 1.0.0

Release-As: 1.0.0
```

Once that release PR is merged, release-please will:

1. Create the git tag (for example `v1.0.0`)
2. Publishes the pipeline image tagged as both `v1.0.0` and `latest`

### Subsequent releases

After `v1.0.0`, version bumps can be driven normally by Conventional Commits without `Release-As`.

### Import note

Before cutting the first release in the new repository, replace the remaining `syntasso/promise-postgresql` GitHub and GHCR coordinates in this directory with the new repository and image locations.

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you
run into issues or have ideas for us, please let us know. Feel free to [open an
issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or
[put time on our calendar](https://www.syntasso.io/contact-us). We'd love to
hear from you.
