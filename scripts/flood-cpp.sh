#!/usr/bin/env bash
set -u

BASE="${BASE:?set BASE, e.g. export BASE=http://127.0.0.1:3000}"
N="${1:-50}"
P="${P:-40}"

# Busy loop so each job holds a worker until the 10s timeout (queue can grow).
BODY='{"language":"cpp","code":"int main(){ while(true); }\n"}'

echo "Firing $N jobs in parallel (up to $P at once) → $BASE/api/execute"

seq 1 "$N" | xargs -n1 -P "$P" -I{} \
  curl -sS -X POST "$BASE/api/execute" \
    -H "Content-Type: application/json" \
    -d "$BODY" \
    -o /dev/null -w "ok %{http_code}\n"

echo
echo "Queue: kubectl exec deploy/redis -- redis-cli LLEN jobs"
echo "Pods:  kubectl get pods -l app=worker-app"
