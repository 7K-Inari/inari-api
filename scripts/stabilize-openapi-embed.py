#!/usr/bin/env python3
"""Normalize the generated OpenAPI embedded blob without changing its JSON.

Some oapi-codegen/Go toolchain combinations produce different valid deflate
streams for identical JSON. Canonical fixed-parameter recompression keeps the
repository drift check deterministic without editing generated semantics.
"""

from __future__ import annotations

import base64
import json
import re
import sys
import zlib

_BLOCK = re.compile(r'var swaggerSpec = \[\]string\{\n(?P<body>.*?)\n\}', re.DOTALL)
_PART = re.compile(r'"([A-Za-z0-9+/=]+)"')


def normalize_generated_go(source: str) -> str:
    match = _BLOCK.search(source)
    if match is None:
        raise ValueError("generated swaggerSpec block not found")
    parts = _PART.findall(match.group("body"))
    if not parts:
        raise ValueError("generated swaggerSpec block is empty")
    try:
        packed = base64.b64decode("".join(parts), validate=True)
        raw = zlib.decompress(packed, wbits=-15)
    except (ValueError, zlib.error) as exc:
        raise ValueError("invalid generated OpenAPI compression") from exc
    # Validate without re-marshaling: the exact generated JSON bytes are the
    # semantic contract and must not be reordered by this stabilizer.
    if not isinstance(json.loads(raw), dict):
        raise ValueError("embedded OpenAPI document must be an object")
    compressor = zlib.compressobj(
        level=zlib.Z_BEST_COMPRESSION,
        method=zlib.DEFLATED,
        wbits=-15,
        memLevel=8,
        strategy=zlib.Z_DEFAULT_STRATEGY,
    )
    normalized_packed = compressor.compress(raw) + compressor.flush()
    encoded = base64.b64encode(normalized_packed).decode()
    normalized_parts = [encoded[index:index + 80] for index in range(0, len(encoded), 80)]
    body = "\n".join(f'\t"{part}",' for part in normalized_parts)
    replacement = "var swaggerSpec = []string{\n" + body + "\n}"
    return source[:match.start()] + replacement + source[match.end():]


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: stabilize-openapi-embed.py GENERATED_GO", file=sys.stderr)
        return 2
    path = sys.argv[1]
    with open(path, encoding="utf-8") as handle:
        source = handle.read()
    normalized = normalize_generated_go(source)
    with open(path, "w", encoding="utf-8") as handle:
        handle.write(normalized)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
