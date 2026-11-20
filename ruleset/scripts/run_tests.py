"""Run the ruleset's unittest files side by side, one process per file (OPS_S00_06).

    python ruleset/scripts/run_tests.py [PATTERN] [--jobs N]

PATTERN is a glob over ruleset/scripts (default test_*.py). Each file runs as
`python -m unittest discover -s ruleset/scripts -p <file>`, exactly as CI ran the whole set before,
so a file sees the same imports and working directory. The files are independent (each reads the
repo, none writes to it), so they run in parallel, by default one per CPU. A failing file prints
its full output; the exit code is 1 if any file failed.
"""

import argparse
import os
import subprocess
import sys
import time
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path

SCRIPTS = Path(__file__).resolve().parent


def run_file(path):
    started = time.perf_counter()
    done = subprocess.run(
        [sys.executable, "-m", "unittest", "discover", "-s", str(SCRIPTS), "-p", path.name],
        capture_output=True,
        text=True,
    )
    return path.name, done.returncode, done.stdout + done.stderr, time.perf_counter() - started


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("pattern", nargs="?", default="test_*.py", help="glob over ruleset/scripts (default test_*.py)")
    parser.add_argument("--jobs", type=int, default=os.cpu_count() or 1, help="files at once (default: one per CPU)")
    args = parser.parse_args(argv)

    files = sorted(SCRIPTS.glob(args.pattern))
    if not files:
        print(f"no test files match {args.pattern}", file=sys.stderr)
        return 1
    started = time.perf_counter()
    failed = []
    with ThreadPoolExecutor(max_workers=max(1, args.jobs)) as pool:
        for name, code, output, seconds in pool.map(run_file, files):
            print(f"{'ok  ' if code == 0 else 'FAIL'} {seconds:5.1f}s  {name}", flush=True)
            if code != 0:
                failed.append((name, output))
    for name, output in failed:
        print(f"\n===== {name} =====\n{output}", flush=True)
    print(f"\n{len(files) - len(failed)} of {len(files)} test files passed in {time.perf_counter() - started:.1f}s")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
