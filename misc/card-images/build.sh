#!/bin/bash

set -euo pipefail

DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)

cd "$DIR"

# This utility is intentionally isolated from the main application dependencies. Install the small
# Node-only adapter dependencies locally, without touching the repository lockfile.
npm install --no-package-lock --no-save jsdom@26.1.0 @types/jsdom@21.1.7 xmlserializer@0.6.1

npx tsc --project "$DIR/tsconfig.json" --noEmit

echo "Card image generator type-check complete."
