#!/bin/bash

set -e

cd "$(dirname "$0")/../apps/finicky/src"
go test ./...
