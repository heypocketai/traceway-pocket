#!/usr/bin/env bash
# Run the hardware benchmark from your laptop, end-to-end, against real Hetzner.
# The GitHub Action is a thin wrapper around run-matrix-entry.sh; this script
# is the developer-facing equivalent — same orchestration, run from anywhere.
#
# Required env:
#   HCLOUD_TOKEN          Hetzner Cloud API token
#   BENCHMARK_SSH_KEY     Path to the private key matching the Hetzner-side
#                         SSH key named 'benchmark-key'.
#
# Common usage:
#   run-local.sh                        # full matrix (4 tiers x 2 modes x 3 signals), throughput
#   run-local.sh --scenario read-probe  # same matrix but ingest-and-probe-reads
#   run-local.sh --smoke                # 1 tier, 1 mode, 1 signal, short steps (~5 min)
#   run-local.sh --tier ccx13 --mode sqlite --signal spans
#   run-local.sh --async                # set CH_ASYNC_INSERT=1; -async suffix on output files
#   run-local.sh --dry-run              # validate env + print plan, no provisioning
#
# Output:
#   --scenario throughput  ->  benchmarks/results-throughput/
#   --scenario read-probe  ->  benchmarks/results-probe/
#   Each scenario folder is wiped at the start of its run so output reflects
#   only the current dispatch. The two folders are never touched by each
#   other's runs — running read-probe will not clobber throughput results.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

TIERS="ccx13,ccx23,ccx33,ccx43"
MODES="sqlite,pgch"
SIGNALS="spans,metrics,logs"
SCENARIO="throughput"
DURATION="30m"
SMOKE=0
DRY_RUN=0
ASYNC=0

usage() {
    sed -n '1,23p' "$0"; exit 2
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --tier)     TIERS="$2"; shift 2 ;;
        --mode)     MODES="$2"; shift 2 ;;
        --signal)   SIGNALS="$2"; shift 2 ;;
        --scenario) SCENARIO="$2"; shift 2 ;;
        --duration) DURATION="$2"; shift 2 ;;
        --smoke)    SMOKE=1; shift ;;
        --async)    ASYNC=1; shift ;;
        --dry-run)  DRY_RUN=1; shift ;;
        -h|--help)  usage ;;
        *) echo "unknown flag: $1" >&2; usage ;;
    esac
done

case "${SCENARIO}" in
    throughput|read-probe) ;;
    *) echo "invalid --scenario '${SCENARIO}' (expected throughput|read-probe)" >&2; exit 2 ;;
esac
export BENCH_SCENARIO="${SCENARIO}"

if [[ "${SMOKE}" -eq 1 ]]; then
    TIERS="ccx13"
    MODES="sqlite"
    SIGNALS="spans"
    DURATION="3m"
    echo "smoke mode: tier=ccx13 mode=sqlite signal=spans scenario=${SCENARIO} duration=${DURATION} (one short run)" >&2
fi

# Preflight first — fail fast on missing tooling. Pass the chosen modes so
# managed-ch can demand CH credentials before we burn a single euro.
export BENCH_MODES_PLAN="${MODES}"
"${SCRIPT_DIR}/preflight.sh"

# One canonical folder per scenario. Wiped at the start so each run's output
# reflects only that run's matrix entries — no cross-dispatch mixing (which
# was the failure mode of the date-folder layout: two dispatches landing on
# the same day silently combined into one summary.md). The throughput folder
# and the read-probe folder are siblings so each scenario stays out of the
# other's way. Cross-run comparison happens via `git log`, not adjacent
# folders.
case "${SCENARIO}" in
    throughput) OUT_DIR="${REPO_ROOT}/benchmarks/results-throughput" ;;
    read-probe) OUT_DIR="${REPO_ROOT}/benchmarks/results-probe" ;;
esac
rm -rf "${OUT_DIR}"
mkdir -p "${OUT_DIR}"
echo "results dir: ${OUT_DIR}" >&2

# Plan: explode tiers x modes x signals.
plan=()
IFS=',' read -ra TIER_ARR <<<"${TIERS}"
IFS=',' read -ra MODE_ARR <<<"${MODES}"
IFS=',' read -ra SIGNAL_ARR <<<"${SIGNALS}"
for t in "${TIER_ARR[@]}"; do
    for m in "${MODE_ARR[@]}"; do
        for s in "${SIGNAL_ARR[@]}"; do
            plan+=("${t}|${m}|${s}")
        done
    done
done

echo "plan (${#plan[@]} entries, scenario=${SCENARIO}):" >&2
for e in "${plan[@]}"; do
    echo "  - ${e//|/ x }" >&2
done

if [[ "${DRY_RUN}" -eq 1 ]]; then
    echo "dry-run: would call run-matrix-entry.sh ${#plan[@]} time(s) with duration=${DURATION}; no servers will be created." >&2
    exit 0
fi

# Sequential execution. Local runs prioritize simplicity and low concurrent
# Hetzner spend over wall-clock speed; the GH workflow parallelizes with
# strategy.matrix instead.
smoke_arg=""
[[ "${SMOKE}" -eq 1 ]] && smoke_arg="smoke"
async_arg=""
if [[ "${ASYNC}" -eq 1 ]]; then
    async_arg="async"
    export CH_ASYNC_INSERT=1
    echo "CH_ASYNC_INSERT=1 — async-insert benchmark pass; output files will get -async suffix" >&2
fi

failures=()
for e in "${plan[@]}"; do
    IFS='|' read -r tier mode signal <<<"${e}"
    if ! "${SCRIPT_DIR}/run-matrix-entry.sh" "${tier}" "${mode}" "${signal}" "${DURATION}" "${OUT_DIR}" "${smoke_arg}" "${async_arg}"; then
        echo "FAIL: ${tier}/${mode}/${signal}" >&2
        failures+=("${tier}/${mode}/${signal}")
    fi
done

# Render charts once everything (or at least something) is done.
if compgen -G "${OUT_DIR}/*.json" >/dev/null; then
    echo "rendering charts" >&2
    python3 "${SCRIPT_DIR}/chart.py" "${OUT_DIR}"
else
    echo "no JSON files were produced; skipping chart render" >&2
fi

if [[ ${#failures[@]} -gt 0 ]]; then
    echo "FINISHED with failures: ${failures[*]}" >&2
    exit 1
fi
echo "FINISHED. Results in ${OUT_DIR}" >&2
