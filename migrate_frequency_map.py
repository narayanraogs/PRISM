#!/usr/bin/env python3
"""
Standalone PRISM FrequencyMap & StageFactor Migration Tool
==========================================================
Completely independent script (no dependencies on PRISM server or client code).
Compatible with Python 3.7+.

Migrates UpDownConverter table to:
- FrequencyMap: JSON array of integer frequency pairs [{"Input": ..., "Output": ...}]
- StageFactor: INTEGER NOT NULL DEFAULT 1

Default Config Search Order:
----------------------------
1. /umacsops/prism/config/config.json
2. ./config/config.json (current working directory)
3. ./config.json (current working directory)
4. ~/prism/config/config.json (user home directory)

Usage:
------
    # 1. Automatic mode:
    python3 migrate_frequency_map.py

    # 2. Preview changes without modifying any file (DRY RUN):
    python3 migrate_frequency_map.py --dry-run

    # 3. Explicitly pass one or more config files:
    python3 migrate_frequency_map.py /umacsops/prism/config/config.json
    python3 migrate_frequency_map.py /umacsops/prism/config/*.json

    # 4. Point to a config directory:
    python3 migrate_frequency_map.py --config-dir /umacsops/prism/config
    python3 migrate_frequency_map.py --config-dir /umacsops/prism/

    # 5. Or directly migrate a database without reading any config file:
    python3 migrate_frequency_map.py --db /umacsops/prism/db/prism.db
"""

from __future__ import annotations

import argparse
import glob
import json
import os
import shutil
import sqlite3
import sys
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Tuple, Union


def resolve_path(p: Union[str, Path]) -> Path:
    """Expand user tilde ~ and resolve absolute path."""
    return Path(os.path.expanduser(str(p))).resolve()


def to_int_freq(val: Union[int, float, str]) -> int:
    """Safely convert frequency to an integer."""
    return int(round(float(val)))


def find_default_configs() -> List[Path]:
    """Look for config.json in standard PRISM locations."""
    candidates = [
        Path("/umacsops/prism/config/config.json"),
        Path.cwd() / "config" / "config.json",
        Path.cwd() / "config.json",
        Path.home() / "prism" / "config" / "config.json",
    ]
    for c in candidates:
        resolved = resolve_path(c)
        if resolved.is_file():
            return [resolved]
    return []


def extract_db_from_config(config_path: Path) -> Tuple[Optional[Path], Optional[str]]:
    """Parse PRISM config JSON and return resolved database path."""
    if not config_path.is_file():
        return None, f"Config file does not exist: {config_path}"
    try:
        with open(config_path, "r", encoding="utf-8") as f:
            cfg = json.load(f)
    except Exception as exc:
        return None, f"Failed to parse JSON in {config_path}: {exc}"

    db_entry = cfg.get("Database")
    if not isinstance(db_entry, dict):
        return None, f"No 'Database' object found in {config_path.name}"

    db_path_str = db_entry.get("DBPath")
    if not db_path_str:
        return None, f"No 'Database.DBPath' field found in {config_path.name}"

    resolved = resolve_path(db_path_str)
    return resolved, None


def backup_database(db_path: Path) -> Path:
    """Create a timestamped backup copy before modifying any database file."""
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    backup_path = db_path.with_name(f"{db_path.stem}_backup_{timestamp}{db_path.suffix}")
    shutil.copy2(db_path, backup_path)
    return backup_path


def get_column_names(cur: sqlite3.Cursor, table_name: str) -> set:
    cur.execute(f'PRAGMA table_info("{table_name}")')
    return {row[1] for row in cur.fetchall()}


def build_frequency_map_json(input_freq: Union[float, int], output_freq: Union[float, int]) -> str:
    """Format FrequencyMap JSON with capitalized 'Input' and 'Output' keys and integer values."""
    pairs = [{"Input": to_int_freq(input_freq), "Output": to_int_freq(output_freq)}]
    return json.dumps(pairs, indent=2)


def validate_migrated_table(cur: sqlite3.Cursor, original_rows: list) -> Tuple[bool, str]:
    """Ensure every row matches original data, has valid JSON structure, integer values, and StageFactor."""
    cur.execute('SELECT "Name", "FrequencyMap", "StageFactor" FROM "UpDownConverter"')
    migrated = {row[0]: (row[1], row[2]) for row in cur.fetchall()}

    for orig in original_rows:
        name = orig["Name"]
        if name not in migrated:
            return False, f"Row '{name}' missing after migration"

        freq_str, stage_factor = migrated[name]
        try:
            pairs = json.loads(freq_str)
        except json.JSONDecodeError as err:
            return False, f"Row '{name}' has invalid JSON ({err})"

        if not isinstance(pairs, list) or len(pairs) == 0:
            return False, f"Row '{name}' FrequencyMap must be a non-empty list"

        first = pairs[0]
        if "Input" not in first or "Output" not in first:
            return False, f"Row '{name}' missing 'Input' or 'Output' key: {first}"

        if not isinstance(first["Input"], int) or not isinstance(first["Output"], int):
            return False, f"Row '{name}' frequency values must be integers: {first}"

        expected_in = to_int_freq(orig["InputFrequency"])
        expected_out = to_int_freq(orig["OutputFrequency"])

        if first["Input"] != expected_in:
            return False, f"Row '{name}' Input mismatch: expected {expected_in}, got {first['Input']}"

        if first["Output"] != expected_out:
            return False, f"Row '{name}' Output mismatch: expected {expected_out}, got {first['Output']}"

        if stage_factor is None:
            return False, f"Row '{name}' StageFactor is NULL"

    return True, f"All {len(original_rows)} row(s) verified OK"


def normalize_existing_frequencymap(cur: sqlite3.Cursor, dry_run: bool) -> Tuple[bool, str]:
    """If table already exists, ensure keys use capitalized 'Input'/'Output', integer frequencies, and StageFactor column."""
    cols = get_column_names(cur, "UpDownConverter")
    added_stage_factor = False

    if "StageFactor" not in cols:
        if not dry_run:
            cur.execute('ALTER TABLE "UpDownConverter" ADD COLUMN "StageFactor" INTEGER NOT NULL DEFAULT 1')
            cur.execute('UPDATE "UpDownConverter" SET "StageFactor" = 1 WHERE "StageFactor" IS NULL')
        added_stage_factor = True

    cur.execute('SELECT "ID", "Name", "FrequencyMap" FROM "UpDownConverter"')
    rows = cur.fetchall()
    needs_update = []

    for row_id, name, freq_str in rows:
        try:
            pairs = json.loads(freq_str)
            updated = False
            new_pairs = []
            for p in pairs:
                raw_in = p.get("Input", p.get("input"))
                raw_out = p.get("Output", p.get("output"))
                in_val = to_int_freq(raw_in)
                out_val = to_int_freq(raw_out)
                new_pairs.append({"Input": in_val, "Output": out_val})

                if (
                    not isinstance(p.get("Input"), int)
                    or not isinstance(p.get("Output"), int)
                    or "input" in p
                    or "output" in p
                ):
                    updated = True

            if updated:
                needs_update.append((row_id, name, json.dumps(new_pairs, indent=2)))
        except json.JSONDecodeError:
            return False, f"Malformed FrequencyMap JSON in row '{name}'"

    msg_parts = []
    if added_stage_factor:
        msg_parts.append("Added 'StageFactor' column (default 1)")
    if needs_update:
        if not dry_run:
            for row_id, name, new_json in needs_update:
                cur.execute('UPDATE "UpDownConverter" SET "FrequencyMap" = ? WHERE "ID" = ?', (new_json, row_id))
        msg_parts.append(f"Normalized {len(needs_update)} FrequencyMap row(s)")

    if not msg_parts:
        return True, "Already up-to-date with FrequencyMap & StageFactor (no changes needed)"

    return True, "; ".join(msg_parts)


def migrate_database_file(db_path: Path, dry_run: bool) -> Tuple[str, str]:
    """
    Perform the schema and data migration on a single SQLite database.
    Returns (Status, Details).
    """
    if not db_path.exists():
        return "ERROR", f"Database file not found: {db_path}"

    conn = sqlite3.connect(db_path)
    cur = conn.cursor()

    try:
        # Check if UpDownConverter table exists
        cur.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='UpDownConverter'")
        if not cur.fetchone():
            conn.close()
            return "SKIPPED", "Table 'UpDownConverter' does not exist in this database"

        cols = get_column_names(cur, "UpDownConverter")

        # Case 1: Already has FrequencyMap (check/update to integer values and add StageFactor if missing)
        if "FrequencyMap" in cols:
            ok, msg = normalize_existing_frequencymap(cur, dry_run)
            if not dry_run and ok:
                conn.commit()
            conn.close()
            return ("SUCCESS" if ok else "ERROR"), msg

        # Case 2: Must have InputFrequency and OutputFrequency
        if "InputFrequency" not in cols or "OutputFrequency" not in cols:
            conn.close()
            return "ERROR", f"Missing InputFrequency/OutputFrequency columns (found: {cols})"

        # Read original data
        cur.execute(
            'SELECT "ID", "Name", "InputFrequency", "OutputFrequency", '
            '"MaxPowerCable", "MinPowerCable", "MaxPowerRadiated", "MinPowerRadiated" '
            'FROM "UpDownConverter"'
        )
        col_names = [d[0] for d in cur.description]
        original_rows = [dict(zip(col_names, row)) for row in cur.fetchall()]

        if dry_run:
            conn.close()
            return "DRY-RUN", f"Would migrate {len(original_rows)} record(s) to FrequencyMap + StageFactor=1"

        # Backup before live modification
        backup_path = backup_database(db_path)

        # Atomic transaction
        conn.execute("BEGIN EXCLUSIVE")

        # Step 1: Rename old table
        conn.execute('ALTER TABLE "UpDownConverter" RENAME TO "UpDownConverter_old"')

        # Step 2: Create new table with FrequencyMap & StageFactor
        conn.execute("""
            CREATE TABLE "UpDownConverter" (
                "ID"               INTEGER PRIMARY KEY AUTOINCREMENT,
                "Name"             TEXT    UNIQUE NOT NULL,
                "FrequencyMap"     TEXT    NOT NULL,
                "StageFactor"      INTEGER NOT NULL DEFAULT 1,
                "MaxPowerCable"    REAL    NOT NULL,
                "MinPowerCable"    REAL    NOT NULL,
                "MaxPowerRadiated" REAL,
                "MinPowerRadiated" REAL
            )
        """)

        # Step 3: Insert transformed rows
        for row in original_rows:
            fmap = build_frequency_map_json(row["InputFrequency"], row["OutputFrequency"])
            conn.execute(
                """
                INSERT INTO "UpDownConverter"
                    ("ID", "Name", "FrequencyMap", "StageFactor", "MaxPowerCable", "MinPowerCable",
                     "MaxPowerRadiated", "MinPowerRadiated")
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                """,
                (
                    row["ID"],
                    row["Name"],
                    fmap,
                    1,
                    row["MaxPowerCable"],
                    row["MinPowerCable"],
                    row["MaxPowerRadiated"],
                    row["MinPowerRadiated"],
                ),
            )

        # Step 4: Validate
        valid, val_msg = validate_migrated_table(cur, original_rows)
        if not valid:
            conn.rollback()
            conn.close()
            return "ERROR", f"Validation failed ({val_msg}). Rolled back."

        # Step 5: Drop temporary table & commit
        conn.execute('DROP TABLE "UpDownConverter_old"')
        conn.execute("COMMIT")
        conn.close()

        return "SUCCESS", f"Migrated {len(original_rows)} record(s) with FrequencyMap & StageFactor=1. Backup: {backup_path.name}"

    except Exception as exc:
        conn.rollback()
        conn.close()
        return "ERROR", f"Exception during migration: {exc}"


def collect_targets(args) -> Tuple[Dict[Path, List[str]], List[str]]:
    """Resolve all target databases and their sources."""
    db_to_sources: Dict[Path, List[str]] = {}
    errors: List[str] = []

    # Direct database flag
    if args.db:
        target_db = resolve_path(args.db)
        db_to_sources[target_db] = ["--db argument"]
        return db_to_sources, errors

    # Config files passed via arguments or directory
    config_files: List[Path] = []
    if args.configs:
        for item in args.configs:
            expanded = os.path.expanduser(item)
            matches = glob.glob(expanded)
            if matches:
                for m in matches:
                    p = Path(m).resolve()
                    if p.is_file():
                        config_files.append(p)
            else:
                p = resolve_path(item)
                if p.is_file():
                    config_files.append(p)
                else:
                    errors.append(f"Specified config file not found: {item}")
    elif args.config_dir:
        cdir = resolve_path(args.config_dir)
        if cdir.is_dir():
            found = list(cdir.glob("*.json"))
            if not found and (cdir / "config").is_dir():
                found = list((cdir / "config").glob("*.json"))
            config_files.extend(sorted(found))
        else:
            errors.append(f"Config directory does not exist: {cdir}")
    else:
        defaults = find_default_configs()
        if defaults:
            config_files.extend(defaults)
        else:
            errors.append(
                "No config files specified and no default config.json found in /umacsops/prism/config/, ./config/, or ~/prism/config/."
            )

    unique_configs = sorted(list(set(config_files)))

    for cfg in unique_configs:
        db_path, err = extract_db_from_config(cfg)
        if err:
            errors.append(f"[{cfg.name}] {err}")
        else:
            db_to_sources.setdefault(db_path, []).append(cfg.name)

    return db_to_sources, errors


def main():
    parser = argparse.ArgumentParser(
        description="Standalone tool to migrate UpDownConverter to FrequencyMap JSON + StageFactor in PRISM SQLite databases."
    )
    parser.add_argument(
        "configs",
        nargs="*",
        help="Optional: config.json files or glob pattern (e.g. /umacsops/prism/config/*.json)",
    )
    parser.add_argument(
        "--db",
        type=Path,
        help="Directly migrate a SQLite .db file without reading a config.json",
    )
    parser.add_argument(
        "--config-dir",
        type=Path,
        help="Directory to scan for *.json config files (e.g. /umacsops/prism/config or /umacsops/prism)",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Preview actions without modifying any database",
    )

    args = parser.parse_args()

    print("=" * 75)
    print(" Standalone PRISM FrequencyMap & StageFactor Migration Tool")
    print(f" Mode: {'DRY RUN (preview only)' if args.dry_run else 'LIVE MIGRATION'}")
    print("=" * 75 + "\n")

    db_to_sources, errors = collect_targets(args)

    if errors:
        for err in errors:
            print(f"[!] {err}")
        print()

    if not db_to_sources:
        print("[ERROR] No databases could be resolved to migrate.")
        print("Please provide a config file, --config-dir, or --db /path/to/database.db")
        sys.exit(1)

    print(f"Found {len(db_to_sources)} unique target database(s) to process:\n")
    for db_path, sources in db_to_sources.items():
        print(f"  * {db_path}")
        print(f"    Source(s): {', '.join(sources)}")
    print("\n" + "-" * 75 + "\n")

    summary = []
    for db_path, sources in db_to_sources.items():
        print(f">> Migrating: {db_path}")
        status, detail = migrate_database_file(db_path, args.dry_run)
        print(f"   Status   : [{status}] {detail}\n")
        summary.append((str(db_path), ", ".join(sources), status, detail))

    print("=" * 75)
    print(" Summary:")
    print("=" * 75)
    for db_str, src_str, status, detail in summary:
        print(f" * [{status}] {db_str}")
        print(f"   From:   {src_str}")
        print(f"   Detail: {detail}")
        print()

    if args.dry_run:
        print("[info] Dry run complete. Re-run without '--dry-run' to execute live.")


if __name__ == "__main__":
    main()
