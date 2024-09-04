#!/usr/bin/env python3
"""Build dafavorites."""

import argparse
import os
import subprocess
import tempfile
import time
import typing

import libtmux

COLOR_GREEN = "\033[32m"
COLOR_RED = "\033[31m"
COLOR_END = "\033[0m"

COVERAGE_FILE = "cover.out"


class Output:
    def __init__(self):
        self._output = ""

    def append(self, value: str, add_newline: bool = True):
        self._output += value
        if add_newline:
            self._output += "\n"

    def print(self) -> int:
        lines = len(self._output.split("\n")) + 1
        print(self._output)
        return lines

    def add_newline(self):
        if self._output.endswith("\n\n"):
            return
        self._output += "\n"


def colorize(color, value):
    return "".join([color, value, COLOR_END])


def run_command(
    *args,
    output: Output,
    append=True,
) -> tuple[typing.Literal[True] | bool, str | None]:
    with tempfile.NamedTemporaryFile("w", encoding="utf-8", delete=False) as filep:
        filen = filep.name
        res = subprocess.run(
            [*args], check=False, stdout=filep, stderr=filep, encoding="utf-8"
        )
    with open(filen, encoding="utf-8") as filep:
        text = filep.read()
    os.remove(filen)
    success = res.returncode == 0
    if append is False:
        return success, text
    output.append(text)
    return success, None


def run_test(html_coverage, silent: bool, fail_fast: bool, output: Output):
    """Run unit tests."""
    alpha = ["go", "test", "-cover"]
    if fail_fast:
        alpha.append("-failfast")
    for_html = ["-coverprofile", COVERAGE_FILE]
    omega = ["-covermode", "count", "./..."]
    if html_coverage:
        success, text = run_command(
            *(alpha + for_html + omega), output=output, append=False
        )
    else:
        success, text = run_command(*(alpha + omega), output=output, append=False)
    if not success:
        output.append(colorize_error(text))
        return success
    if silent:
        return success
    table = format_table(text)
    output.append(table)
    return success


def filter_testless(lines):
    fixed = []
    for each in lines:
        if each.startswith("\t"):
            continue
        fixed.append(each)
    return fixed


def format_table(text):
    """Format lines in text into a table."""
    lines = filter_testless(text.rstrip().split("\n"))
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
    table = ""
    for line in lines:
        pieces = [e.strip() for e in line.strip().split("\t")]
        pieces += [""] * (len(lengths) - len(pieces))
        table += form.format(*pieces) + "\n"
    return table


def colorize_error(text: str) -> str:
    pieces = text.split("Error:")
    if len(pieces) == 1:
        return text
    return (colorize(COLOR_RED, "Error") + ":").join(pieces)


def generate_html_coverage(output: Output):
    """Generage HTML coverage."""
    return run_command(
        "go", "tool", "cover", "-html", COVERAGE_FILE, "-o", "cover.html", output=output
    )[0]


def main(args, output: Output):
    """Build dafavorites."""
    if not args.only_test:
        output.append(colorize(COLOR_GREEN, "-- BUILD"))
        if not run_command("go", "install", "./...", output=output)[0]:
            return
        output.add_newline()

    output.append(colorize(COLOR_GREEN, "-- TEST"))
    run_test(args.html_coverage, args.silent_test, args.ff, output)
    if args.html_coverage:
        generate_html_coverage(output)


def find_go_files():
    """Find all .go files."""
    result = []
    for root, _, files in os.walk("./"):
        for each in (e for e in files if e.endswith(".go")):
            result.append(os.path.join(root, each))
    return result


def derive_max_window_height(window: libtmux.Window) -> int:
    height = window.height
    assert height.isdigit(), height
    return int(height) // 2


def keep_waiting(args):
    """Run indefinitely waiting for file changes."""
    tmux_server = libtmux.Server()
    session = tmux_server.sessions[0]
    window = session.active_window
    pane = session.active_pane
    while True:
        subprocess.run(["clear"], check=False)
        output = Output()
        main(args, output)

        output.add_newline()
        output.append(colorize(COLOR_GREEN, "-- WAIT"))

        files = find_go_files()
        lines = output.print()
        pane.resize(height=min(derive_max_window_height(window), lines))
        subprocess.run(["inotifywait", "-q", "-e", "close_write"] + files, check=False)
        time.sleep(0.5)


def cli():
    """Create CLI."""
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--html-coverage", "-c", help="Generage HTML coverage.", action="store_true"
    )
    parser.add_argument(
        "--only-test", "-t", help="Skip install, only run tests.", action="store_true"
    )
    parser.add_argument(
        "--silent-test",
        "-s",
        action="store_true",
        help="When tests pass, print nothing.",
    )
    parser.add_argument(
        "--ff", action="store_true", help="Stop test run on first failure."
    )
    return parser.parse_args()


if __name__ == "__main__":
    keep_waiting(cli())
