#!/bin/sh
# Writes the inputs of the himorime suite in bench/. Deterministic: the same
# arguments always write the same bytes, so a base and a head revision read
# the same input. Only sh and awk are needed.
#
#   sh gen.sh specs N FILE   one spec of N scenarios, each running echo
#                            without a shell and asserting its exit code
#                            and standard output
set -eu

case "$1" in
specs)
	mkdir -p "$(dirname "$3")"
	awk -v n="$2" 'BEGIN {
		print "version: \"1\""
		print ""
		print "suite:"
		printf "  name: bench %d scenarios\n", n
		print ""
		print "scenarios:"
		for (i = 0; i < n; i++) {
			printf "  - name: scenario %d\n", i
			print  "    steps:"
			print  "      - run:"
			printf "          command: echo scenario-%d\n", i
			print  "      - assert:"
			print  "          exit_code: 0"
			print  "          stdout:"
			printf "            contains: scenario-%d\n", i
		}
	}' > "$3"
	;;
*)
	echo "gen.sh: unknown kind $1" >&2
	exit 2
	;;
esac
