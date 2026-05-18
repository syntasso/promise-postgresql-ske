# Airgapped environments

Use `scripts/update-images.sh` to list all external images referenced by this Promise and optionally rewrite them to point to a private registry.

List images:

```sh
./scripts/update-images.sh
```

Remap to a private registry (prompts for confirmation before modifying files):

```sh
./scripts/update-images.sh my-registry.example.com
```