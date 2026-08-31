#!/usr/bin/env python3
"""Split flux-local rendered output into one file per homelab-app-chart app.

Keeps only documents rendered from homelab-app chart templates and writes
each app's documents to <out_dir>/<app>.yaml (keyed by its
app.kubernetes.io/name label), so Trivy misconfig ignores can be scoped to
one app instead of the whole rendered chart output.
"""
import re
import sys
from pathlib import Path

NAME_RE = re.compile(r"^\s*app\.kubernetes\.io/name:\s*(\S+)", re.MULTILINE)


def main() -> None:
    in_file, out_dir = Path(sys.argv[1]), Path(sys.argv[2])
    out_dir.mkdir(parents=True, exist_ok=True)
    for existing in out_dir.glob("*.yaml"):
        existing.unlink()

    apps: dict[str, list[str]] = {}
    for doc in in_file.read_text().split("\n---\n"):
        if "# Source: homelab-app/" not in doc:
            continue
        match = NAME_RE.search(doc)
        if not match:
            continue
        apps.setdefault(match.group(1), []).append(doc.strip("\n"))

    for app, docs in apps.items():
        (out_dir / f"{app}.yaml").write_text("\n---\n".join(docs) + "\n")

    print(f"split {sum(len(d) for d in apps.values())} documents across {len(apps)} apps")


if __name__ == "__main__":
    main()
