# atago documentation

A browsable, repo-local index of atago's documentation, generated from repository sources by `make site` (see `internal/sitegen`). Every link below points at a file committed in this repository; it is rendered by GitHub and is not a hosted website.

> Regenerate with `make site`. A drift test (`TestSite_InSync`) keeps this in sync with the sources.

## Start here

- [Project README](../README.md)
## Behavior docs (generated from executable specs)

- [README.md](../doc/e2e/README.md)
- [actionlint.md](../doc/e2e/actionlint.md)
- [age.md](../doc/e2e/age.md)
- [atago.md](../doc/e2e/atago.md)
- [awscli.md](../doc/e2e/awscli.md)
- [caddy.md](../doc/e2e/caddy.md)
- [career.md](../doc/e2e/career.md)
- [coredns.md](../doc/e2e/coredns.md)
- [ffmpeg.md](../doc/e2e/ffmpeg.md)
- [fzf.md](../doc/e2e/fzf.md)
- [git.md](../doc/e2e/git.md)
- [gitea.md](../doc/e2e/gitea.md)
- [gotify.md](../doc/e2e/gotify.md)
- [grafana.md](../doc/e2e/grafana.md)
- [gup.md](../doc/e2e/gup.md)
- [htop.md](../doc/e2e/htop.md)
- [hugo.md](../doc/e2e/hugo.md)
- [iso8583tool.md](../doc/e2e/iso8583tool.md)
- [jose.md](../doc/e2e/jose.md)
- [jq.md](../doc/e2e/jq.md)
- [kustomize.md](../doc/e2e/kustomize.md)
- [mailpit.md](../doc/e2e/mailpit.md)
- [mimixbox.md](../doc/e2e/mimixbox.md)
- [minio.md](../doc/e2e/minio.md)
- [mobilepkg.md](../doc/e2e/mobilepkg.md)
- [nats.md](../doc/e2e/nats.md)
- [ntfy.md](../doc/e2e/ntfy.md)
- [openssl.md](../doc/e2e/openssl.md)
- [pandoc.md](../doc/e2e/pandoc.md)
- [prometheus.md](../doc/e2e/prometheus.md)
- [pushgateway.md](../doc/e2e/pushgateway.md)
- [python.md](../doc/e2e/python.md)
- [rclone.md](../doc/e2e/rclone.md)
- [redis.md](../doc/e2e/redis.md)
- [restic.md](../doc/e2e/restic.md)
- [sops.md](../doc/e2e/sops.md)
- [sqlite3.md](../doc/e2e/sqlite3.md)
- [sqly.md](../doc/e2e/sqly.md)
- [ssh-keygen.md](../doc/e2e/ssh-keygen.md)
- [terraform.md](../doc/e2e/terraform.md)
- [transfersh.md](../doc/e2e/transfersh.md)
- [truss.md](../doc/e2e/truss.md)
- [webhook.md](../doc/e2e/webhook.md)

## Schemas

- [Spec file schema](../schema/atago.schema.json)
- [Manifest output schema](../schema/manifest.schema.json)
- [Report output schema](../schema/report.schema.json)
- [Manifest example](../schema/examples/manifest.example.json)
- [Report example](../schema/examples/report.example.json)

## Samples gallery

Deterministic artifacts generated from a fixture run (see [samples/README.md](samples/README.md)):

- Reports: [JSON](samples/report.json) · [JUnit XML](samples/report.junit.xml) · [TAP](samples/report.tap)
- Generated PDF: [sample.pdf](samples/sample.pdf)
- Image diff: [baseline](samples/imagediff/baseline.png) · [actual](samples/imagediff/actual.png) · [diff](samples/imagediff/diff.png)

## Demos

![Run demo](../doc/img/demo.gif)

![Review demo](../doc/img/review.gif)

![Snapshot demo](../doc/img/snapshot.gif)

