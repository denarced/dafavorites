#!/usr/bin/env python3
"""Build dafavorites."""

import argparse
import os
import subprocess
import tempfile
import time
import typing

COLOR_GREEN = "\033[32m"
COLOR_END = "\033[0m"

COVERAGE_FILE = "cover.out"


def colorize(color, value):
    return "".join([color, value, COLOR_END])


def run_command(
        *args,
        capture=False
) -> tuple[typing.Literal[True] | bool, str | None]:
    if not capture:
        return subprocess.run(
            [*args],
            check=False
        ).returncode == 0, None
    filen = None
    with tempfile.NamedTemporaryFile(
            "w",
            encoding="utf-8",
            delete=False
    ) as filep:
        filen = filep.name
        res = subprocess.run(
            [*args],
            check=False,
            stdout=filep,
            stderr=filep,
            encoding="utf-8"
        )
    success = res.returncode == 0
    with open(filen, encoding="utf-8") as filep:
        text = filep.read()
    os.remove(filen)
    if not success:
        print(text, end="")
        return success, None
    return success, text


def run_test(html_coverage, fail_fast):
    """Run unit tests."""
    fail = ["-failfast"] if fail_fast else []
    alpha = ["go", "test"] + fail + ["-cover"]
    for_html = ["-coverprofile", COVERAGE_FILE]
    omega = ["-covermode", "count", "./..."]
    if html_coverage:
        success, text = run_command(*(alpha + for_html + omega), capture=True)
    else:
        success, text = run_command(*(alpha + omega), capture=True)
    if not success:
        return success
    print_table(text)
    return success


def print_table(text):
    """Print text as a table with equally wide columns."""
    lines = text.strip().split("\n")
    lengths = []
    for line in lines:
        columns = line.strip().split("\t")
        for index, each in enumerate(columns):
            if index >= len(lengths):
                lengths.append(0)
            width = len(each.strip())
            if width > lengths[index]:
                lengths[index] = width
    form_pieces = [""] * len(lengths)
    for index, each in enumerate(lengths):
        
        form_pieces[index] = f"{{:{each}s}}"
        
    form = " ".join(form_pieces)
    for line in lines:
        pieces = [e.strip() for e in line.strip().split("\t")]
        pieces += [""] * (len(lengths) - len(pieces))
        print(form.format(*pieces))


def generate_html_coverage():
    """Generage HTML coverage."""
    return run_command(
        "go",
        "tool",
        "cover",
        "-html",
        COVERAGE_FILE,
        "-o",
        "cover.html")[0]


def main(args):
    """Build dafavorites."""
    if not args.only_test:
        print(colorize(COLOR_GREEN, "-- BUILD"))
        if not run_command("go", "install", "./...")[0]:
            return
        print()

    print(colorize(COLOR_GREEN, "-- TEST"))
    run_test(args.html_coverage, args.fail_fast)
    if args.html_coverage:
        generate_html_coverage()


def find_go_files():
    """Find all .go files."""
    result = []
    for root, _, files in os.walk("./"):
        for each in (e for e in files if e.endswith(".go")):
            result.append(os.path.join(root, each))
    return result


def keep_waiting(args):
    """Run indefinitely waiting for file changes."""
    while True:
        run_command("clear")
        main(args)

        print()
        print(colorize(COLOR_GREEN, "-- WAIT"))

        files = find_go_files()
        run_command(
            "inotifywait",
            "-q",
            "-e",
            "close_write",
            *files
        )
        time.sleep(0.5)


def cli():
    """Create CLI."""
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--html-coverage",
        "-c",
        help="Generage HTML coverage.",
        action="store_true")
    parser.add_argument(
        "--only-test",
        "-t",
        help="Skip install, only run tests.",
        action="store_true")
    parser.add_argument(
        "--fail-fast",
        "-f",
        help="Stop test run on first failure.",
        action="store_true")
    return parser.parse_args()


if __name__ == "__main__":
    keep_waiting(cli())
