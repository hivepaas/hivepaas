#!/usr/bin/env python3
"""Runs a command the way a person at a terminal would: answers sudo's password
prompt and the installer's questions, and stops the command if it hangs.

  distro_e2e_drive.py LOG -- COMMAND...

COMMAND is `ssh -tt ...`: a terminal on the far side, its output here, which is
also written to LOG. The answers come from the environment: DRIVE_SUDO_PASSWORD,
DRIVE_ADMIN_EMAIL, DRIVE_ADMIN_PASSWORD and DRIVE_APP_DOMAIN. A question with a
default in brackets - [Y/n], [localhost] - gets its default.

Exits with COMMAND's status, or:
  124  it hangs: no output for DRIVE_IDLE_SECONDS (default 600), or still
       running after DRIVE_TIMEOUT_SECONDS (default 2700)
  125  it asks a question none of the answers is for
"""

import codecs
import os
import re
import selectors
import subprocess
import sys
import time

# Colours and cursor moves, and the carriage returns of a terminal's lines.
ANSI = re.compile(r"\x1b\[[0-9;?]*[ -/]*[@-~]|\x1b[()][0-9A-Za-z]|\r")

# What a question looks like: a line left open, ending as a prompt does.
PROMPT = re.compile(r"[:?\]] $")

# How long an open question may wait before it counts as one with no answer.
PROMPT_WAIT_SECONDS = 30


def answers():
    env = os.environ.get
    return [
        # sudo's "[sudo] password for tester: ", sudo-rs's "[sudo: authenticate] Password: "
        (re.compile(r"\[sudo[^\]]*\][^\n]*[Pp]assword[^\n]*: ?$"), env("DRIVE_SUDO_PASSWORD"), "sudo's password"),
        (re.compile(r"Admin password, again: $"), env("DRIVE_ADMIN_PASSWORD"), "the admin's password"),
        (re.compile(r"Admin password \([^)]*\): $"), env("DRIVE_ADMIN_PASSWORD"), "the admin's password"),
        (re.compile(r"Admin email: $"), env("DRIVE_ADMIN_EMAIL"), "the admin's email"),
        (re.compile(r"App domain\b.*: $"), env("DRIVE_APP_DOMAIN"), "the app domain"),
        (re.compile(r"\[[Yy]/[Nn]\] $"), "", "a question with a default"),
        (re.compile(r"\[[^\]\n]*\]: $"), "", "a question with a default"),
    ]


def stop(proc, log, status, why):
    message = f"\n[distro_e2e_drive] {why}\n"
    log.write(message.encode())
    log.flush()
    sys.stderr.write(message)
    proc.terminate()
    try:
        proc.wait(timeout=5)
    except subprocess.TimeoutExpired:
        proc.kill()
        proc.wait()
    return status


def main():
    if len(sys.argv) < 4 or sys.argv[2] != "--":
        sys.stderr.write(__doc__)
        return 2
    log_path, command = sys.argv[1], sys.argv[3:]
    timeout = int(os.environ.get("DRIVE_TIMEOUT_SECONDS", "2700"))
    idle = int(os.environ.get("DRIVE_IDLE_SECONDS", "600"))
    rules = answers()

    proc = subprocess.Popen(command, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    out = proc.stdout.fileno()
    os.set_blocking(out, False)
    selector = selectors.DefaultSelector()
    selector.register(out, selectors.EVENT_READ)
    decoder = codecs.getincrementaldecoder("utf-8")("replace")
    # What was printed since the last answer, without colours.
    text = ""
    start = last = time.monotonic()

    with open(log_path, "ab") as log:
        while True:
            now = time.monotonic()
            line = text.rsplit("\n", 1)[-1]
            if now - start > timeout:
                return stop(proc, log, 124, f"still running after {timeout}s: it hangs")
            if PROMPT.search(line) and now - last > PROMPT_WAIT_SECONDS:
                return stop(proc, log, 125, f"a question with no answer: {line.strip()!r}")
            if now - last > idle:
                return stop(proc, log, 124, f"no output for {idle}s: it hangs")
            if not selector.select(timeout=1):
                continue
            try:
                chunk = os.read(out, 65536)
            except BlockingIOError:
                continue
            if not chunk:
                break
            log.write(chunk)
            log.flush()
            sys.stdout.buffer.write(chunk)
            sys.stdout.flush()
            last = time.monotonic()
            text = (text + ANSI.sub("", decoder.decode(chunk)))[-4000:]
            line = text.rsplit("\n", 1)[-1]
            for pattern, answer, name in rules:
                if pattern.search(line):
                    if answer is None:
                        return stop(proc, log, 125, f"nothing to answer {name} with: {line.strip()!r}")
                    proc.stdin.write(answer.encode() + b"\r")
                    proc.stdin.flush()
                    text = ""
                    break
    return proc.wait()


if __name__ == "__main__":
    sys.exit(main())
