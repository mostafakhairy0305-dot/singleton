#!/usr/bin/env bash
#
# The per-package coverage gate. This is the single source of truth for the
# floor: CI runs it, and so should you before pushing.
#
#   ./scripts/cover.sh          run the gate
#   ./scripts/cover.sh --html   also write cover.html for local drill-down
#
# The floor is per package, not per module, because a module-wide average lets
# a well-tested package pay for an untested one.
#
# The floor is 100 because every non-excluded package is at 100 today, and a
# lower bar would let that erode a statement at a time without anyone noticing.
# Relax it for a single run with COVERAGE_THRESHOLD=90 ./scripts/cover.sh.

set -euo pipefail

threshold=${COVERAGE_THRESHOLD:-100}
html=0
for arg in "$@"; do
	case "$arg" in
	--html) html=1 ;;
	*)
		echo "usage: $0 [--html]" >&2
		exit 2
		;;
	esac
done

cd "$(dirname "$0")/.."

raw=cover.raw.out
profile=cover.out
totals=cover.totals

go test ./... -coverprofile="$raw" -covermode=atomic

# The three native adapters are the operating system itself. corelocation's run
# loop cannot execute without granted location permission and a live run loop;
# geoclue and winrt do not even compile on every host. They still carry every
# hermetic test they admit — see their _test.go files — but a coverage bar is
# not a thing CI can hold them to. examples/ is a main package demonstrating
# the API, not library code.
grep -v -E 'examples/|adapter/corelocation/|adapter/geoclue/|adapter/winrt/' "$raw" >"$profile" || true

if [[ ! -s "$profile" ]]; then
	echo "cover.sh: the filtered profile is empty; did go test produce anything?" >&2
	exit 1
fi

if [[ "$html" -eq 1 ]]; then
	go tool cover -html="$profile" -o cover.html
	echo "wrote cover.html"
fi

# Profile lines are "import/path/file.go:startLine.col,endLine.col numStmt count".
# The package is everything before the last slash; a statement counts as
# covered when it was reached at least once.
awk '
	NR == 1 { next }                       # the "mode:" header
	{
		split($1, location, ":")
		path = location[1]
		segments = split(path, part, "/")
		pkg = part[1]
		for (i = 2; i < segments; i++) pkg = pkg "/" part[i]
		total[pkg] += $2
		if ($3 > 0) covered[pkg] += $2
		seen[pkg] = 1
	}
	END {
		for (pkg in seen) print pkg, covered[pkg] + 0, total[pkg] + 0
	}
' "$profile" | sort >"$totals"

# Reported paths are relative to the module, so the table stays readable when
# the module path is long.
module=$(go list -m)

awk -v threshold="$threshold" -v root="$module" '
	BEGIN {
		prefix = root "/"
		printf "%-56s %10s %8s\n", "PACKAGE", "COVERED", "PERCENT"
	}
	{
		pkg = $1; covered = $2; total = $3
		if (root != "" && pkg == root) pkg = "."
		else if (root != "" && index(pkg, prefix) == 1) pkg = substr(pkg, length(prefix) + 1)
		# A package of pure interface declarations has no statements to cover.
		# It passes; reporting it as 0% would be a lie about nothing.
		percent = (total == 0) ? 100 : covered * 100 / total
		printf "%-56s %10s %7.1f%%\n", pkg, covered "/" total, percent
		if (total > 0 && percent < threshold) failures[++n] = sprintf("%s (%.1f%%, %d/%d)", pkg, percent, covered, total)
	}
	END {
		if (n == 0) {
			printf "\nevery package is at or above %d%%\n", threshold
			exit 0
		}
		printf "\n%d package(s) below the %d%% floor:\n", n, threshold
		for (i = 1; i <= n; i++) printf "  %s\n", failures[i]
		exit 1
	}
' "$totals"
