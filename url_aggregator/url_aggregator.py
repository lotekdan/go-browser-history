#!/usr/bin/env python3
"""
url_aggregator.py — Aggregate browser history JSON into bucketed URL/domain outputs.

Usage:
    python3 url_aggregator.py --input <file.json> [--output-dir <dir>] [--prefix <prefix>]

Arguments:
    --input       Path to input JSON file (list of objects with 'timestamp' and 'url' fields)
    --output-dir  Directory to write output files (default: current directory)
    --prefix      Optional filename prefix for all outputs (default: none)

Outputs (6 JSON files):
    unique_urls_1min.json     — Unique URLs (no args) bucketed every 1 minute
    unique_urls_5min.json     — Unique URLs (no args) bucketed every 5 minutes
    unique_urls_15min.json    — Unique URLs (no args) bucketed every 15 minutes
    unique_domains_1min.json  — Unique domains/subdomains bucketed every 1 minute
    unique_domains_5min.json  — Unique domains/subdomains bucketed every 5 minutes
    unique_domains_15min.json — Unique domains/subdomains bucketed every 15 minutes

Each bucket entry format:
    {
      "bucket_start": "2026-03-25T15:00:00+00:00",
      "latest_timestamp": "2026-03-25T15:14:58+00:00",
      "count": 7,
      "unique_urls": ["https://example.com/page", ...]   // or "unique_domains"
    }

Input JSON record format (minimum required fields):
    {
      "timestamp": "2026-03-25T15:24:04-05:00",  // ISO 8601, any timezone offset
      "url": "https://example.com/page?q=foo"
    }
"""

import argparse
import json
import sys
import os
from urllib.parse import urlparse, urlunparse
from datetime import datetime, timezone

try:
    import pandas as pd
except ImportError:
    print("Error: pandas is required. Install with: pip install pandas", file=sys.stderr)
    sys.exit(1)


def strip_url_args(url: str) -> str:
    """Remove query string and fragment from a URL, keeping scheme, host, and path."""
    try:
        parsed = urlparse(url)
        return urlunparse((parsed.scheme, parsed.netloc, parsed.path, '', '', ''))
    except Exception:
        return url


def extract_domain(url: str) -> str:
    """Extract scheme + netloc (domain + subdomain) from a URL."""
    try:
        parsed = urlparse(url)
        return f"{parsed.scheme}://{parsed.netloc}"
    except Exception:
        return url


def bucket_floor(ts: pd.Timestamp, minutes: int) -> pd.Timestamp:
    """Floor a timestamp to the nearest N-minute boundary."""
    return ts.floor(f'{minutes}min')


def aggregate(df: pd.DataFrame, minutes: int, value_col: str, output_key: str) -> dict:
    """
    Group rows into time buckets and collect unique values per bucket.

    Args:
        df:         DataFrame with 'timestamp' (UTC) and value_col columns
        minutes:    Bucket width in minutes
        value_col:  Column to deduplicate within each bucket
        output_key: Key name for the list in output JSON ('unique_urls' or 'unique_domains')

    Returns:
        Dict keyed by bucket_start ISO string, sorted chronologically.
    """
    df2 = df.copy()
    df2['bucket'] = df2['timestamp'].apply(lambda t: bucket_floor(t, minutes))

    result = {}
    for bucket, group in df2.groupby('bucket'):
        values = sorted(group[value_col].unique().tolist())
        latest = group['timestamp'].max().isoformat()
        bucket_str = bucket.isoformat()
        result[bucket_str] = {
            "bucket_start": bucket_str,
            "latest_timestamp": latest,
            "count": len(values),
            output_key: values
        }

    return dict(sorted(result.items()))


def load_input(path: str) -> pd.DataFrame:
    """Load and validate the input JSON file, returning a clean DataFrame."""
    if not os.path.exists(path):
        print(f"Error: Input file not found: {path}", file=sys.stderr)
        sys.exit(1)

    with open(path, 'r', encoding='utf-8') as f:
        try:
            raw = json.load(f)
        except json.JSONDecodeError as e:
            print(f"Error: Failed to parse JSON: {e}", file=sys.stderr)
            sys.exit(1)

    if not isinstance(raw, list):
        print("Error: Input JSON must be a list of objects.", file=sys.stderr)
        sys.exit(1)

    if len(raw) == 0:
        print("Error: Input JSON list is empty.", file=sys.stderr)
        sys.exit(1)

    df = pd.DataFrame(raw)

    if 'url' not in df.columns:
        print("Error: Input records must have a 'url' field.", file=sys.stderr)
        sys.exit(1)
    if 'timestamp' not in df.columns:
        print("Error: Input records must have a 'timestamp' field.", file=sys.stderr)
        sys.exit(1)

    # Drop rows with missing url or timestamp
    before = len(df)
    df = df.dropna(subset=['url', 'timestamp'])
    dropped = before - len(df)
    if dropped:
        print(f"Warning: Dropped {dropped} rows with missing url/timestamp.", file=sys.stderr)

    # Parse timestamps, coercing bad values to NaT
    df['timestamp'] = pd.to_datetime(df['timestamp'], utc=True, errors='coerce')
    bad_ts = df['timestamp'].isna().sum()
    if bad_ts:
        print(f"Warning: Dropped {bad_ts} rows with unparseable timestamps.", file=sys.stderr)
        df = df.dropna(subset=['timestamp'])

    if len(df) == 0:
        print("Error: No valid rows remaining after cleaning.", file=sys.stderr)
        sys.exit(1)

    df = df.sort_values('timestamp').reset_index(drop=True)
    return df


def write_output(data: dict, path: str) -> None:
    """Write a dict as pretty-printed JSON to path."""
    os.makedirs(os.path.dirname(path) if os.path.dirname(path) else '.', exist_ok=True)
    with open(path, 'w', encoding='utf-8') as f:
        json.dump(data, f, indent=2)


def main():
    parser = argparse.ArgumentParser(
        description="Aggregate browser history JSON into bucketed URL/domain outputs.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__
    )
    parser.add_argument(
        '--input', '-i',
        required=True,
        metavar='FILE',
        help='Path to input JSON file (list of objects with timestamp and url fields)'
    )
    parser.add_argument(
        '--output-dir', '-o',
        default='.',
        metavar='DIR',
        help='Directory to write output JSON files (default: current directory)'
    )
    parser.add_argument(
        '--prefix', '-p',
        default='',
        metavar='PREFIX',
        help='Optional filename prefix, e.g. "march25_" → "march25_unique_urls_1min.json"'
    )

    args = parser.parse_args()

    # Load and clean data
    print(f"Loading: {args.input}")
    df = load_input(args.input)
    print(f"  {len(df):,} valid records | "
          f"{df['timestamp'].min()} → {df['timestamp'].max()}")

    # Derive stripped URL and domain columns
    df['url_stripped'] = df['url'].apply(strip_url_args)
    df['domain']       = df['url'].apply(extract_domain)

    # Define the 6 outputs: (filename, minutes, value_col, output_key)
    outputs = [
        ("unique_urls_1min.json",      1,  'url_stripped', 'unique_urls'),
        ("unique_urls_5min.json",      5,  'url_stripped', 'unique_urls'),
        ("unique_urls_15min.json",     15, 'url_stripped', 'unique_urls'),
        ("unique_domains_1min.json",   1,  'domain',       'unique_domains'),
        ("unique_domains_5min.json",   5,  'domain',       'unique_domains'),
        ("unique_domains_15min.json",  15, 'domain',       'unique_domains'),
    ]

    print(f"\nWriting outputs to: {os.path.abspath(args.output_dir)}/")
    for filename, minutes, value_col, output_key in outputs:
        prefixed_name = f"{args.prefix}{filename}"
        out_path = os.path.join(args.output_dir, prefixed_name)

        data = aggregate(df, minutes, value_col, output_key)
        write_output(data, out_path)

        # Count total unique values across all buckets for summary
        total_unique = len(df[value_col].unique())
        print(f"  {prefixed_name:<40} {len(data):>4} buckets  |  {total_unique:>5} total unique {output_key.split('_')[1]}")

    print("\nDone.")


if __name__ == '__main__':
    main()
