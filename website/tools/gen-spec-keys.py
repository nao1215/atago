#!/usr/bin/env python3
"""Regenerate website/data/spec_keys.json: spec-key path -> first atago release.

Walks every release tag's schema/atago.schema.json and records the first tag
each key path appears in; keys only on main map to "unreleased", except the
ones a release PR has already stamped with the version it is about to tag.

The website workflow runs this before every Hugo build. The release PR stamps
its new keys with the upcoming version, and that stamp is kept while the
version is newer than every tag: the site is deployed when the release PR
merges, before the tag exists, and GitHub Pages does not replace a deployment
of the same commit when the tag's own build runs, so a stamp rewritten to
"unreleased" there stayed on the site until an unrelated push. A stamp that is
not newer than the latest tag names a release the key is not in, so it goes
back to "unreleased". The committed copy is also what a local `hugo server`
reads and what schema_parity_test.go checks the key inventory against. Refresh
it from the repository root:

    python3 website/tools/gen-spec-keys.py

Key paths are $defs-relative ("run.command", "cdp.actions[].click") plus
top-level paths ("suite.name"), matching the lookup the site's spec-reference
shortcode performs. $ref pointers are not followed, so each definition's keys
are recorded exactly once.
"""
import json
import re
import subprocess

KEYS_FILE = "website/data/spec_keys.json"
VERSION = re.compile(r"^v(\d+)\.(\d+)\.(\d+)$")


def version_tuple(v):
    """(major, minor, patch) for a vX.Y.Z tag, or None for anything else."""
    m = VERSION.match(v or "")
    return tuple(int(x) for x in m.groups()) if m else None


def since(path, first, stamped, latest_tag):
    """The release a key path first appeared in.

    first maps a path to the first tag carrying it. For a path no tag carries,
    the committed stamp is kept when it names a version newer than latest_tag
    (a release PR stamped it for the tag about to be cut); anything else is
    "unreleased".
    """
    if path in first:
        return first[path]
    stamp = version_tuple(stamped.get(path))
    latest = version_tuple(latest_tag)
    if stamp is not None and (latest is None or stamp > latest):
        return stamped[path]
    return "unreleased"


def extract_paths(schema):
    paths = set()

    def walk(o, prefix):
        if not isinstance(o, dict):
            return
        props = o.get("properties")
        if isinstance(props, dict):
            for k, v in props.items():
                p = f"{prefix}.{k}" if prefix else k
                paths.add(p)
                walk(v, p)
        if isinstance(o.get("items"), dict):
            walk(o["items"], prefix + "[]")
        if isinstance(o.get("additionalProperties"), dict):
            walk(o["additionalProperties"], prefix + ".*")
        for key in ("oneOf", "anyOf"):
            for x in o.get(key) or []:
                walk(x, prefix)

    walk(schema, "")
    for name, d in (schema.get("$defs") or {}).items():
        walk(d, name)
    return paths


def main():
    tags = subprocess.run(
        ["git", "tag", "--sort=version:refname"], capture_output=True, text=True, check=True
    ).stdout.split()
    first = {}
    for tag in tags:
        r = subprocess.run(
            ["git", "show", f"{tag}:schema/atago.schema.json"], capture_output=True, text=True
        )
        if r.returncode != 0:
            continue
        for p in extract_paths(json.loads(r.stdout)):
            first.setdefault(p, tag)
    with open("schema/atago.schema.json") as f:
        current = extract_paths(json.load(f))
    try:
        with open(KEYS_FILE) as f:
            stamped = json.load(f).get("keys", {})
    except FileNotFoundError:
        stamped = {}
    releases = [t for t in tags if version_tuple(t)]
    latest_tag = max(releases, key=version_tuple) if releases else None
    data = {p: since(p, first, stamped, latest_tag) for p in sorted(current)}
    with open(KEYS_FILE, "w") as f:
        json.dump({"spec_version": "1", "keys": data}, f, indent=1, ensure_ascii=False)
        f.write("\n")
    print(f"{len(data)} keys written to {KEYS_FILE}")


if __name__ == "__main__":
    main()
