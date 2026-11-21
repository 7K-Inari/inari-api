import base64
import importlib.util
import json
import unittest
import zlib
from pathlib import Path

spec = importlib.util.spec_from_file_location("stabilize", Path(__file__).parents[1] / "scripts" / "stabilize-openapi-embed.py")
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


def generated_blob(raw: bytes, level: int) -> str:
    compressor = zlib.compressobj(level=level, wbits=-15)
    packed = compressor.compress(raw) + compressor.flush()
    encoded = base64.b64encode(packed).decode()
    parts = [encoded[index:index + 80] for index in range(0, len(encoded), 80)]
    body = "\n".join(f'\t"{part}",' for part in parts)
    return f'var swaggerSpec = []string{{\n{body}\n}}'


class StabilizeOpenAPIEmbedTest(unittest.TestCase):
    def test_normalizes_compression_without_changing_json(self):
        raw = json.dumps({"paths": {f"/item/{index}": {"description": "alpha-beta-" * 20 + str(index)} for index in range(100)}, "openapi": "3.0.3"}, separators=(",", ":")).encode()
        first = generated_blob(raw, 1)
        second = generated_blob(raw, 9)
        self.assertNotEqual(first, second)
        source = "package oas\n\n" + first + "\n"
        replacement = module.normalize_generated_go(source)
        # The test remains red until the implementation actually canonicalizes.
        self.assertNotEqual(replacement, source)
        self.assertEqual(replacement, module.normalize_generated_go("package oas\n\n" + second + "\n"))

    def test_rejects_missing_or_invalid_blob(self):
        with self.assertRaises(ValueError):
            module.normalize_generated_go("var swaggerSpec = []string{\n\t\"not-base64\",\n}")


if __name__ == "__main__":
    unittest.main()
