"""
TOON Formatter for Code Cartographer

Token-Oriented Object Notation (TOON) encoder/decoder for efficient
code map serialization. Provides ~40% token savings vs JSON.

Based on TOON spec v3.0: https://toonformat.dev/
"""

import re
import json
from typing import Any, Dict, List, Optional, Tuple, Union
from dataclasses import dataclass, field, asdict
from datetime import datetime


@dataclass
class ToonOptions:
    indent: int = 2
    delimiter: str = ","
    key_folding: str = "off"
    flatten_depth: Optional[int] = None
    strict: bool = True


class ToonEncodeError(Exception):
    pass


class ToonDecodeError(Exception):
    pass


def _needs_quoting(value: str, delimiter: str = ",", is_row: bool = False) -> bool:
    if not isinstance(value, str):
        return False
    if value == "":
        return True
    if value.strip() != value:
        return True
    if value in ("true", "false", "null"):
        return True
    if re.match(r"^-?\d+(?:\.\d+)?(?:e[+-]?\d+)?$", value, re.IGNORECASE):
        return True
    if re.match(r"^0\d+$", value):
        return True
    if ":" in value:
        return True
    if '"' in value or "\\" in value:
        return True
    if "[" in value or "]" in value or "{" in value or "}" in value:
        return True
    if "\n" in value or "\r" in value or "\t" in value:
        return True
    if delimiter in value:
        return True
    if value == "-" or value.startswith("-"):
        return True
    return False


def _escape_string(value: str) -> str:
    return (
        value.replace("\\", "\\\\")
        .replace('"', '\\"')
        .replace("\n", "\\n")
        .replace("\r", "\\r")
        .replace("\t", "\\t")
    )


def _unescape_string(value: str) -> str:
    replacements = {
        "\\\\": "\\",
        '\\"': '"',
        "\\n": "\n",
        "\\r": "\r",
        "\\t": "\t",
    }
    result = value
    for old, new in replacements.items():
        result = result.replace(old, new)
    return result


def _normalize_number(value: Union[int, float]) -> str:
    if isinstance(value, bool):
        return "true" if value else "false"
    if value is None:
        return "null"
    if isinstance(value, (int, float)):
        if isinstance(value, float):
            if value != value or value == float("inf") or value == float("-inf"):
                return "null"
            normalized = int(value) if value == int(value) else float(value)
            if normalized == 0:
                return "0"
            if isinstance(normalized, float):
                return f"{normalized:.10f}".rstrip("0").rstrip(".")
            return str(normalized)
        return str(value)
    return str(value)


class ToonEncoder:
    def __init__(self, options: Optional[ToonOptions] = None):
        self.options = options or ToonOptions()
        self._delimiter = self.options.delimiter
        self._indent_str = " " * self.options.indent

    def encode(self, data: Any) -> str:
        if isinstance(data, dict):
            return self._encode_object(data, depth=0)
        elif isinstance(data, list):
            return self._encode_array(data, depth=0, key=None)
        else:
            return self._encode_primitive(data)

    def _encode_object(self, obj: Dict, depth: int) -> str:
        lines = []
        indent = self._indent_str * depth
        for key, value in obj.items():
            if isinstance(value, dict):
                lines.append(f"{indent}{key}:")
                lines.append(self._encode_object(value, depth + 1))
            elif isinstance(value, list):
                lines.append(f"{indent}{key}:")
                lines.append(self._encode_array(value, depth + 1, key=key))
            else:
                encoded_value = self._encode_primitive(value)
                lines.append(f"{indent}{key}: {encoded_value}")
        return "\n".join(lines)

    def _encode_array(self, arr: List, depth: int, key: Optional[str]) -> str:
        if not arr:
            header = f"[0]" if key is None else f"{key}[0]"
            return header + ":"
        first_item = arr[0] if arr else None
        indent = self._indent_str * depth
        header_indent = self._indent_str * (depth - 1) if depth > 0 else ""
        if key is not None:
            header_indent = indent

        if all(isinstance(item, dict) for item in arr):
            if all(set(item.keys()) == set(arr[0].keys()) for item in arr):
                fields = list(arr[0].keys())
                field_str = "{" + ",".join(fields) + "}"
                header = f"{header_indent}{key}[{len(arr)}]{field_str}:"
                lines = [header]
                for item in arr:
                    values = []
                    for field_name in fields:
                        val = item.get(field_name)
                        if val is None:
                            values.append("")
                        elif isinstance(val, (dict, list)):
                            values.append(self._encode_primitive(val))
                        else:
                            values.append(self._encode_primitive(val))
                    row_values = []
                    for v in values:
                        if _needs_quoting(v, self._delimiter, True):
                            row_values.append(f'"{_escape_string(v)}"')
                        else:
                            row_values.append(v)
                    lines.append(f"{indent}{self._delimiter.join(row_values)}")
                return "\n".join(lines)

        header = f"{header_indent}[{len(arr)}]:"
        if all(not isinstance(item, (dict, list)) for item in arr):
            values = []
            for item in arr:
                encoded = self._encode_primitive(item)
                if _needs_quoting(encoded, self._delimiter, True):
                    values.append(f'"{_escape_string(encoded)}"')
                else:
                    values.append(encoded)
            return header + " " + self._delimiter.join(values)

        lines = [header]
        for item in arr:
            if isinstance(item, dict):
                lines.append(f"{indent}-")
                lines.append(self._encode_object(item, depth + 2))
            elif isinstance(item, list):
                lines.append(f"{indent}- [0]:")
            else:
                encoded = self._encode_primitive(item)
                if _needs_quoting(encoded, self._delimiter, False):
                    encoded = f'"{_escape_string(encoded)}"'
                lines.append(f"{indent}- {encoded}")
        return "\n".join(lines)

    def _encode_primitive(self, value: Any) -> str:
        if value is None:
            return "null"
        if isinstance(value, bool):
            return "true" if value else "false"
        if isinstance(value, (int, float)):
            return _normalize_number(value)
        if isinstance(value, str):
            if _needs_quoting(value, self._delimiter, False):
                return f'"{_escape_string(value)}"'
            return value
        return str(value)


class ToonDecoder:
    def __init__(self, strict: bool = True):
        self.strict = strict
        self._lines: List[str] = []
        self._pos = 0
        self._indent_size = 2

    def decode(self, toon_str: str) -> Any:
        self._lines = toon_str.split("\n")
        self._pos = 0
        return self._parse()

    def _parse(self) -> Any:
        if not self._lines:
            return {}

        first_line = self._lines[0].strip()
        if first_line.startswith("["):
            return self._parse_array_at_root()
        return self._parse_object_at_root()

    def _parse_object_at_root(self) -> Dict:
        result = {}
        current_depth = 0
        stack = [(result, 0)]

        while self._pos < len(self._lines):
            line = self._lines[self._pos]
            stripped = line.strip()

            if stripped == "":
                self._pos += 1
                continue

            depth = len(line) - len(line.lstrip())
            depth //= self._indent_size

            while stack and depth <= stack[-1][1]:
                stack.pop()

            if ":" not in stripped:
                self._pos += 1
                continue

            if depth == 0 and "{" in stripped and "[" in stripped:
                self._pos += 1
                continue

            colon_idx = stripped.index(":")
            key = stripped[:colon_idx].strip()
            value_str = stripped[colon_idx + 1 :].strip()

            if not value_str:
                obj = {}
                parent_obj, parent_depth = stack[-1] if stack else (result, -1)
                if parent_depth < depth - 1:
                    pass
                if key in parent_obj:
                    current_obj = parent_obj[key]
                else:
                    parent_obj[key] = obj
                    current_obj = obj
                stack.append((current_obj, depth))
            elif value_str.startswith("["):
                arr = self._parse_array(depth, key)
                parent_obj = stack[-1][0] if stack else result
                parent_obj[key] = arr
            else:
                parent_obj = stack[-1][0] if stack else result
                parent_obj[key] = self._parse_primitive(value_str)

            self._pos += 1

        return result

    def _parse_array_at_root(self) -> List:
        result = []
        while self._pos < len(self._lines):
            line = self._lines[self._pos].strip()
            if not line:
                self._pos += 1
                continue
            if line.startswith("["):
                result.append(self._parse_array(0, None))
            else:
                break
        return result

    def _parse_array(self, parent_depth: int, key: Optional[str]) -> List:
        line = self._lines[self._pos].strip()
        self._pos += 1

        header_match = re.match(r"(\w+)?\[(\d+)\](?:\{([^}]+)\})?:$", line)
        if header_match:
            arr_key = header_match.group(1)
            count = int(header_match.group(2))
            fields = header_match.group(3).split(",") if header_match.group(3) else None
            result = []

            if fields:
                rows = []
                while self._pos < len(self._lines):
                    row_line = self._lines[self._pos]
                    stripped = row_line.strip()
                    depth = len(row_line) - len(row_line.lstrip())
                    depth //= self._indent_size
                    if depth <= parent_depth:
                        break
                    if stripped:
                        values = self._split_values(stripped, len(fields))
                        obj = {}
                        for i, field_name in enumerate(fields):
                            obj[field_name] = self._parse_primitive(values[i]) if i < len(values) else ""
                        rows.append(obj)
                    self._pos += 1
                return rows

            values_match = re.search(r":\s*(.+)$", line)
            if values_match:
                inline_values = self._split_values(values_match.group(1), count)
                return [self._parse_primitive(v) for v in inline_values]

            items = []
            while self._pos < len(self._lines):
                next_line = self._lines[self._pos]
                stripped = next_line.strip()
                depth = len(next_line) - len(next_line.lstrip())
                depth //= self._indent_size
                if depth <= parent_depth:
                    break
                if stripped.startswith("- "):
                    item_str = stripped[2:]
                    if item_str.startswith("["):
                        arr = self._parse_array(depth, None)
                        items.append(arr)
                    elif item_str.startswith("{") or (":" in item_str and len(item_str.split(":")[0].strip()) > 0):
                        if ":" in item_str:
                            parts = item_str.split(":", 1)
                            obj = {parts[0].strip(): self._parse_primitive(parts[1].strip())}
                            self._pos += 1
                            while self._pos < len(self._lines):
                                inner_line = self._lines[self._pos]
                                inner_depth = len(inner_line) - len(inner_line.lstrip())
                                inner_depth //= self._indent_size
                                if inner_depth <= depth:
                                    break
                                if ":" in inner_line:
                                    k, v = inner_line.split(":", 1)
                                    obj[k.strip()] = self._parse_primitive(v.strip())
                                self._pos += 1
                            items.append(obj)
                        else:
                            items.append(self._parse_primitive(item_str))
                    else:
                        items.append(self._parse_primitive(item_str))
                elif stripped and not stripped.startswith("-"):
                    break
                else:
                    self._pos += 1
                    continue
                self._pos += 1
            return items

        if ":" not in line:
            return []

        arr_match = re.match(r"(\w+)?\[(\d+)\]:\s*(.+)$", line)
        if arr_match:
            count = int(arr_match.group(2))
            values_str = arr_match.group(3)
            values = self._split_values(values_str, count)
            return [self._parse_primitive(v) for v in values]

        return []

    def _split_values(self, line: str, expected_count: int) -> List[str]:
        delimiter = self._detect_delimiter(line)
        values = []
        current = ""
        in_quotes = False
        i = 0
        while i < len(line):
            char = line[i]
            if char == '"':
                in_quotes = not in_quotes
                current += char
            elif char == delimiter and not in_quotes:
                values.append(current.strip())
                current = ""
            else:
                current += char
            i += 1
        if current:
            values.append(current.strip())
        while len(values) < expected_count:
            values.append("")
        return values[:expected_count]

    def _detect_delimiter(self, line: str) -> str:
        if "\t" in line:
            return "\t"
        if "|" in line:
            return "|"
        return ","

    def _parse_primitive(self, value_str: str) -> Any:
        value_str = value_str.strip()

        if value_str == "":
            return ""

        if value_str.startswith('"') and value_str.endswith('"'):
            return _unescape_string(value_str[1:-1])

        if value_str == "null":
            return None
        if value_str == "true":
            return True
        if value_str == "false":
            return False

        if re.match(r"^-?\d+(?:\.\d+)?(?:e[+-]?\d+)?$", value_str, re.IGNORECASE):
            if "." in value_str or "e" in value_str.lower():
                return float(value_str)
            return int(value_str)

        return value_str


def to_toon(data: Any, options: Optional[ToonOptions] = None) -> str:
    encoder = ToonEncoder(options)
    return encoder.encode(data)


def from_toon(toon_str: str, strict: bool = True) -> Any:
    decoder = ToonDecoder(strict)
    return decoder.decode(toon_str)


def toon_to_json(toon_str: str) -> str:
    data = from_toon(toon_str)
    return json.dumps(data, separators=(",", ":"), default=str)


def json_to_toon(json_str: str, options: Optional[ToonOptions] = None) -> str:
    data = json.loads(json_str)
    return to_toon(data, options)
