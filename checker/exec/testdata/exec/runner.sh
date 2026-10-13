#!/bin/sh -eu

cat <<STDOUT
this is the first line of stdout
this is the second line of stdout
this is the third line of stdout
this is the fourth line of stdout
STDOUT

cat <<STDERR >&2
this is the first line of stderr
this is the second line of stderr
this is the third line of stderr
this is the fourth line of stderr
STDERR

exit "${1:-0}"