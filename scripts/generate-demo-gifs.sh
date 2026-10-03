#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR/.."
VHS_DIR="$PROJECT_ROOT/scripts/vhs"
ASSETS_DIR="$PROJECT_ROOT/assets/demos"

# Maximum number of VHS recordings to run concurrently.
# Can be overridden via the TUICHART_MAX_CONCURRENT env var.
TUICHART_MAX_CONCURRENT="${TUICHART_MAX_CONCURRENT:-4}"

usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Build the demo binary and record all VHS tape files into GIFs, running
up to \$TUICHART_MAX_CONCURRENT recordings in parallel.

Options:
  -j, --jobs N    Maximum number of concurrent recordings (default: $TUICHART_MAX_CONCURRENT)
  -h, --help      Show this help message and exit

Environment:
  TUICHART_MAX_CONCURRENT  Same as --jobs

Examples:
  $(basename "$0")
  $(basename "$0") --jobs 8
  TUICHART_MAX_CONCURRENT=2 $(basename "$0")
EOF
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -j|--jobs)
                if [[ -z "${2:-}" ]]; then
                    echo "Error: --jobs requires a value" >&2
                    usage >&2
                    exit 1
                fi
                TUICHART_MAX_CONCURRENT="$2"
                shift 2
            ;;
            -h|--help)
                usage
                exit 0
            ;;
            *)
                echo "Error: unknown argument: $1" >&2
                usage >&2
                exit 1
            ;;
        esac
    done
    
    if ! [[ "$TUICHART_MAX_CONCURRENT" =~ ^[1-9][0-9]*$ ]]; then
        echo "Error: TUICHART_MAX_CONCURRENT must be a positive integer (got: $TUICHART_MAX_CONCURRENT)" >&2
        exit 1
    fi
}

build_demo_binary() {
    echo "Building demo binary..."
    go build -o "$VHS_DIR/tuichartdemo" ./scripts/demo
}

# Record a single tape file. Runs VHS from the tape directory so that the
# relative paths inside the tape (Source ../_shared/base.tape,
# Output ../../assets/demos/) resolve correctly.
record_tape() {
    local tape="$1"
    local name
    name="$(basename "$tape" .tape)"
    
    echo "▶ Recording $name"
    ( cd "$VHS_DIR" && vhs "$(basename "$tape")" )
    echo "✔ Finished $name"
}

record_all_tapes() {
    echo "Recording demo GIFs (max $TUICHART_MAX_CONCURRENT concurrent)..."
    
    local -a pids=()
    local -a names=()
    local failed=0
    
    # Block until the number of running jobs drops below TUICHART_MAX_CONCURRENT.
    throttle() {
        while (( $(jobs -rp | wc -l) >= TUICHART_MAX_CONCURRENT )); do
            wait -n 2>/dev/null || failed=1
        done
    }
    
    local tape
    for tape in "$VHS_DIR"/*.tape; do
        [[ -e "$tape" ]] || continue
        throttle
        record_tape "$tape" &
        pids+=("$!")
        names+=("$(basename "$tape" .tape)")
    done
    
    # Wait for remaining jobs and collect status.
    local i
    for i in "${!pids[@]}"; do
        if ! wait "${pids[$i]}"; then
            echo "✖ Failed: ${names[$i]}" >&2
            failed=1
        fi
    done
    
    return "$failed"
}

cleanup() {
    echo "Cleaning up..."
    rm -f "$VHS_DIR/tuichartdemo"
}

main() {
    parse_args "$@"
    
    cd "$PROJECT_ROOT"
    mkdir -p "$ASSETS_DIR"
    
    # Ensure cleanup runs even if recording fails.
    trap cleanup EXIT
    
    build_demo_binary
    record_all_tapes
    
    cleanup
    trap - EXIT
    
    echo "Done! GIFs are in $ASSETS_DIR"
    ls -la "$ASSETS_DIR"/demo-*.gif 2>/dev/null || echo "No demo GIFs found"
}

main "$@"