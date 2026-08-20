#!/usr/bin/env bash
set -u

BASE="${BASE:?set BASE, e.g. export BASE=http://127.0.0.1:3000}"
N="${1:-50}"
P="${P:-40}"

# Sleep under the 10s worker timeout so jobs finish with output, not SIGKILL.
# Still holds a worker long enough for the Redis list to grow.
BODY='{"language":"cpp","code":"#include <unistd.h>\n#include <iostream>\nint main(){ sleep(8); std::cout << 2 << std::endl; }\n"}'

echo "Firing $N jobs in parallel (up to $P at once) → $BASE/api/execute"

seq 1 "$N" | xargs -n1 -P "$P" -I{} \
  curl -sS -X POST "$BASE/api/execute" \
    -H "Content-Type: application/json" \
    -d "$BODY" \
    -o /dev/null -w "ok %{http_code}\n"

echo
echo "Queue: kubectl exec deploy/redis -- redis-cli LLEN jobs"
echo "Pods:  kubectl get pods -l app=worker-app"
