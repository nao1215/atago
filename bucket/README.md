# Scoop bucket

This directory is atago's [Scoop](https://scoop.sh/) bucket. Scoop accepts any
git repository with a `bucket/` directory, so the manifest lives here rather
than in a second repository whose only content would be one generated file:

```shell
scoop bucket add nao1215 https://github.com/nao1215/atago
scoop install nao1215/atago
```

`atago.json` is written by GoReleaser on every tagged release — it carries the
release's Windows archive URLs and their SHA-256 hashes, so it cannot be edited
by hand without going stale at the next tag. Change
[`.goreleaser.yml`](../.goreleaser.yml) instead.

The manifest does not exist until the first release that generates it.
