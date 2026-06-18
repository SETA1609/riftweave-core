#!/usr/bin/env bash
set -euo pipefail

IMAGE="${IMAGE:-riftweave-test}"
DOCKERFILE="${DOCKERFILE:-Dockerfile}"
ROOT="$(cd "$(dirname "$0")" && pwd)"

echo "==> Building Docker image: ${IMAGE}"
docker build -q -t "${IMAGE}" -f "${ROOT}/${DOCKERFILE}" "${ROOT}" > /dev/null

fail=0

echo ""
echo "==> Running schema validation..."
# Override ENTRYPOINT so we can run validate.py directly
if docker run --rm --entrypoint python "${IMAGE}" ruleset/scripts/validate.py; then
    echo "    Schema validation: PASSED"
else
    echo "    Schema validation: FAILED"
    fail=1
fi

echo ""
echo "==> Running module system tests..."
if docker run --rm --entrypoint python "${IMAGE}" -m unittest ruleset/scripts/test_module_system.py 2>&1; then
    echo "    Module tests: PASSED"
else
    echo "    Module tests: FAILED"
    fail=1
fi

echo ""
if [ "${fail}" -eq 0 ]; then
    echo "All tests passed."
else
    echo "Some tests FAILED."
fi

exit "${fail}"
