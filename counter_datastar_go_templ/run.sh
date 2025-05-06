#!/bin/sh

## This script uses `air` tool to run the project
## and restart it on detected file changes.
## go install github.com/air-verse/air@latest

# firstly find `.air.toml` in current directory, if not found, use defaults
air -c .air.toml