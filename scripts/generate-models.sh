#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT

normalized_spec="${tmp_dir}/openapi.normalized.yaml"

cd "${repo_root}"

go run ./cmd/tools/openapi-normalize -in "${repo_root}/openapi.yaml" -out "${normalized_spec}"
go tool oapi-codegen -config "${repo_root}/pkg/models/oapi-codegen.yaml" "${normalized_spec}"
