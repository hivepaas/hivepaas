#!/bin/bash
#
# The full test run: race detector plus a coverage profile.
#
# This is what CI runs. `make test` is the one to use while working - it skips
# the coverage instrumentation and comes back in about a second.

set -euo pipefail

TEST_RESULT_DIR="${TEST_RESULTS:-./test-results}"
mkdir -p "$TEST_RESULT_DIR"

PROFILE="$TEST_RESULT_DIR/coverage.out"

echo "---------------------------------------------------------------"
echo "Test:"
echo "---------------------------------------------------------------"

# No -coverpkg, deliberately.
#
# It attributes a package's lines to every test binary that could reach them,
# which on this repo produced a 1.2GB profile - 12.9M lines, of which 40k were
# distinct - and moved the total from 13.0% to 13.5%. Half a point is not worth a
# file that size, least of all one meant to be uploaded on every CI run.
#
# What that flag's package list was really doing was excluding generated code and
# entry points from the report. That belongs in codecov.yml's `ignore`, where it
# costs nothing and can be read.
#
# -covermode=atomic is what -race selects anyway. Stating it means the mode does
# not quietly fall back to `set` if someone takes -race off.
go test -race -covermode=atomic -coverprofile="$PROFILE" ./...

echo "---------------------------------------------------------------"
echo "Result:"
echo "---------------------------------------------------------------"

# The total only. The per-function table is hundreds of lines of CI log, and it
# is already in the profile for anyone who wants to look.
go tool cover -func="$PROFILE" | tail -1

# Nobody reads HTML on CI, and Codecov does not want it.
if [ -z "${CI:-}" ]; then
  go tool cover -html="$PROFILE" -o "$TEST_RESULT_DIR/coverage.html"
  echo "HTML report: $TEST_RESULT_DIR/coverage.html"
fi

echo "---------------------------------------------------------------"
echo "DONE."
echo "---------------------------------------------------------------"
