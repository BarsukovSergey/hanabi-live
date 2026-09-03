#!/bin/bash

set -euo pipefail

DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
REPO_ROOT="$DIR/../.."

bash "$DIR/build.sh"
cd "$REPO_ROOT"

npx tsx "$DIR/src/createAllCards.ts" "$@"
