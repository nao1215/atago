# atago documentation

A browsable, repo-local index of atago's documentation, generated from repository sources by `make site` (see `internal/sitegen`). Every link below points at a file committed in this repository; it is rendered by GitHub and is not a hosted website.

> Regenerate with `make site`. A drift test (`TestSite_InSync`) keeps this in sync with the sources.

## Start here

- [Project README](../README.md)
## Behavior docs (generated from executable specs)

- [README.md](../doc/e2e/README.md)
- [actionlint.md](../doc/e2e/actionlint.md)
- [age.md](../doc/e2e/age.md)
- [aqua.md](../doc/e2e/aqua.md)
- [atago.md](../doc/e2e/atago.md)
- [awscli.md](../doc/e2e/awscli.md)
- [bats.md](../doc/e2e/bats.md)
- [caddy.md](../doc/e2e/caddy.md)
- [career.md](../doc/e2e/career.md)
- [coredns.md](../doc/e2e/coredns.md)
- [ecspresso.md](../doc/e2e/ecspresso.md)
- [ffmpeg.md](../doc/e2e/ffmpeg.md)
- [fx.md](../doc/e2e/fx.md)
- [fzf.md](../doc/e2e/fzf.md)
- [getoptions.md](../doc/e2e/getoptions.md)
- [ghostscript.md](../doc/e2e/ghostscript.md)
- [git-extras.md](../doc/e2e/git-extras.md)
- [git-open.md](../doc/e2e/git-open.md)
- [git-secrets.md](../doc/e2e/git-secrets.md)
- [git.md](../doc/e2e/git.md)
- [gitea.md](../doc/e2e/gitea.md)
- [gotify.md](../doc/e2e/gotify.md)
- [gpg.md](../doc/e2e/gpg.md)
- [grafana.md](../doc/e2e/grafana.md)
- [gum.md](../doc/e2e/gum.md)
- [gup.md](../doc/e2e/gup.md)
- [helix.md](../doc/e2e/helix.md)
- [htop.md](../doc/e2e/htop.md)
- [hugo.md](../doc/e2e/hugo.md)
- [imagemagick.md](../doc/e2e/imagemagick.md)
- [iso8583tool.md](../doc/e2e/iso8583tool.md)
- [jose.md](../doc/e2e/jose.md)
- [jq.md](../doc/e2e/jq.md)
- [kubectx.md](../doc/e2e/kubectx.md)
- [kustomize.md](../doc/e2e/kustomize.md)
- [lazygit.md](../doc/e2e/lazygit.md)
- [mailpit.md](../doc/e2e/mailpit.md)
- [mimixbox.md](../doc/e2e/mimixbox.md)
- [minio.md](../doc/e2e/minio.md)
- [mobilepkg.md](../doc/e2e/mobilepkg.md)
- [nats.md](../doc/e2e/nats.md)
- [ntfy.md](../doc/e2e/ntfy.md)
- [openfga.md](../doc/e2e/openfga.md)
- [openssl.md](../doc/e2e/openssl.md)
- [pandoc.md](../doc/e2e/pandoc.md)
- [prometheus.md](../doc/e2e/prometheus.md)
- [pushgateway.md](../doc/e2e/pushgateway.md)
- [python.md](../doc/e2e/python.md)
- [rbenv.md](../doc/e2e/rbenv.md)
- [rclone.md](../doc/e2e/rclone.md)
- [redis.md](../doc/e2e/redis.md)
- [restic.md](../doc/e2e/restic.md)
- [shdotenv.md](../doc/e2e/shdotenv.md)
- [shellspec.md](../doc/e2e/shellspec.md)
- [sops.md](../doc/e2e/sops.md)
- [sqlite3.md](../doc/e2e/sqlite3.md)
- [sqly.md](../doc/e2e/sqly.md)
- [ssh-keygen.md](../doc/e2e/ssh-keygen.md)
- [terraform.md](../doc/e2e/terraform.md)
- [transcrypt.md](../doc/e2e/transcrypt.md)
- [transfersh.md](../doc/e2e/transfersh.md)
- [truss.md](../doc/e2e/truss.md)
- [unzip.md](../doc/e2e/unzip.md)
- [webhook.md](../doc/e2e/webhook.md)
- [yazi.md](../doc/e2e/yazi.md)
- [zstd.md](../doc/e2e/zstd.md)

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

