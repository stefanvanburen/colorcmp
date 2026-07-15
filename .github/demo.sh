#!/bin/sh
# Drives the README demo GIFs. Recorded with asciinema and rendered to GIF with
# agg; see the Makefile `demo` target. TestReporterDemo logs a colored diff.
set -e

printf '$ go test -v -run TestReporterDemo .\n'
sleep 1
go test -v -run TestReporterDemo .
sleep 2
