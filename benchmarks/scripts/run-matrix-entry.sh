#!/usr/bin/env bash
# Run ONE (tier, mode) cycle end-to-end: provision Hetzner -> bootstrap SUT ->
# run loadgen -> fetch JSON -> tear down. Called by both run-local.sh and the
# GitHub workflow; treat this as the single source of truth for matrix-entry
# orchestration.
#
# Usage: run-matrix-entry.sh <tier> <mode> <signal> <duration> <out-dir> [smoke]
#   <tier>      ccx13 | ccx23 | ccx33 | ccx43
#   <mode>      sqlite | pgch
#   <signal>    spans | metrics | logs
#   <duration>  Loadgen total runtime, e.g. 30m, 3m
#   <out-dir>   Directory to write <tier>-<mode>-<signal>.json into
#   [smoke]     "smoke" to enable short-step overrides (--phase1-batch-sizes
#               256,1024 --phase2-request-rates 1,5 --step-duration 15s).
#               Optional.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

if [[ $# -lt 5 ]]; then
    echo "usage: $0 <tier> <mode> <signal> <duration> <out-dir> [smoke]" >&2
    exit 2
fi
TIER="$1"; MODE="$2"; SIGNAL="$3"; DURATION="$4"; OUT_DIR="$5"; SMOKE="${6:-}"
LOCATION="${BENCH_LOCATION:-nbg1}"
SCENARIO="${BENCH_SCENARIO:-throughput}"

case "${SIGNAL}" in
    spans|metrics|logs) ;;
    *) echo "invalid signal '${SIGNAL}' (expected spans|metrics|logs)" >&2; exit 2 ;;
esac

case "${SCENARIO}" in
    throughput|read-probe) ;;
    *) echo "invalid BENCH_SCENARIO '${SCENARIO}' (expected throughput|read-probe)" >&2; exit 2 ;;
esac

# Hetzner caps server names at 63 chars; the prefix `bench-loadgen-` eats 14,
# so the RUN_ID must stay <= 49 chars. Abbreviate the scenario to keep margin
# for the worst combo (e.g. managed-ch + metrics + read-probe).
case "${SCENARIO}" in
    throughput) SCEN_SHORT="tp" ;;
    read-probe) SCEN_SHORT="rp" ;;
    *)          SCEN_SHORT="${SCENARIO}" ;;
esac
RUN_ID="$(date -u +%Y%m%d-%H%M%S)-${TIER}-${MODE}-${SIGNAL}-${SCEN_SHORT}-$RANDOM"
echo "=== run-matrix-entry tier=${TIER} mode=${MODE} signal=${SIGNAL} scenario=${SCENARIO} duration=${DURATION} run_id=${RUN_ID} ===" >&2

mkdir -p "${OUT_DIR}"

# Always tear down — even on failure, even on Ctrl-C. The trap is set BEFORE
# any hcloud create call so a failure mid-provision still cleans up.
cleanup() {
    local rc=$?
    echo "--- teardown for ${RUN_ID} (exit=${rc}) ---" >&2
    "${SCRIPT_DIR}/hetzner-down.sh" "${RUN_ID}" || true
    exit "${rc}"
}
trap cleanup EXIT INT TERM

# 1. Provision.
INFRA_JSON=$("${SCRIPT_DIR}/hetzner-up.sh" "${TIER}" "${RUN_ID}" "${LOCATION}")
echo "infra: ${INFRA_JSON}" >&2
SUT_PUBLIC_IP=$(printf '%s' "${INFRA_JSON}" | jq -r '.sutPublicIp')
SUT_PRIVATE_IP=$(printf '%s' "${INFRA_JSON}" | jq -r '.sutPrivateIp')
LOADGEN_PUBLIC_IP=$(printf '%s' "${INFRA_JSON}" | jq -r '.loadgenPublicIp')

# 2. Bring up the backend on the SUT.
"${SCRIPT_DIR}/sut-bootstrap.sh" "${SUT_PUBLIC_IP}" "${MODE}"

# 3. Run the loadgen, pulling JSON back into OUT_DIR.
extra_args=( --scenario "${SCENARIO}" )
if [[ "${SMOKE}" == "smoke" ]]; then
    if [[ "${SCENARIO}" == "read-probe" ]]; then
        extra_args+=( --fill-levels 100000,1000000 --settle-seconds 5s )
    else
        extra_args+=( --phase1-batch-sizes 256,1024 --phase2-request-rates 1,5 --step-duration 15s )
    fi
fi

OUT_PATH="${OUT_DIR}/${TIER}-${MODE}-${SIGNAL}-${SCENARIO}.json"
"${SCRIPT_DIR}/loadgen-bootstrap.sh" \
    "${LOADGEN_PUBLIC_IP}" \
    "${SUT_PRIVATE_IP}" \
    "${SUT_PUBLIC_IP}" \
    "${DURATION}" \
    "${TIER}" \
    "${MODE}" \
    "${SIGNAL}" \
    "${OUT_PATH}" \
    "${extra_args[@]}"

# Trap handles teardown — no explicit call needed.
echo "matrix entry ${TIER}-${MODE}-${SIGNAL}-${SCENARIO} complete -> ${OUT_PATH}" >&2
