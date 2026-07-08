#!/usr/bin/env python3
"""Regenerate website/data/spec_keys.json: spec-key path -> first atago release.

Walks every release tag's schema/atago.schema.json and records the first tag
each key path appears in; keys only on main map to "unreleased". Run from the
repository root after tagging a release that changed the schema:

    python3 website/tools/gen-spec-keys.py

Key paths are $defs-relative ("run.command", "cdp.actions[].click") plus
top-level paths ("suite.name"), matching the lookup the site's spec-reference
shortcode performs. $ref pointers are not followed, so each definition's keys
are recorded exactly once.
"""
import json
import subprocess


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
    data = {p: first.get(p, "unreleased") for p in sorted(current)}
    with open("website/data/spec_keys.json", "w") as f:
        json.dump({"spec_version": "1", "keys": data}, f, indent=1, ensure_ascii=False)
        f.write("\n")
    print(f"{len(data)} keys written to website/data/spec_keys.json")


if __name__ == "__main__":
    main()
