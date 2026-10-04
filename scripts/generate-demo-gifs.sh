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
Usage: $(basename "$0") <command> [OPTIONS]

Commands:
  help                Show this help message and exit
  perform             Build the demo binary and record all VHS tapes into GIFs

Options for 'perform':
  -j, --jobs N        Maximum number of concurrent recordings (default: $TUICHART_MAX_CONCURRENT)
  -h, --help          Show this help message and exit

Environment:
  TUICHART_MAX_CONCURRENT  Same as --jobs

Examples:
  $(basename "$0") perform
  $(basename "$0") perform --jobs 8
  TUICHART_MAX_CONCURRENT=2 $(basename "$0") perform
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
    go build -o "$VHS_DIR/tuichartdemo" ./scripts/demo >/dev/null 2>&1
}

# Record a single tape file. Runs VHS from the tape directory so that the
# relative paths inside the tape (Source ../_shared/base.tape,
# Output ../../assets/demos/) resolve correctly.
record_tape() {
    local tape="$1"
    local name
    name="$(basename "$tape" .tape)"
    
    echo "▶ Recording $name"
    ( cd "$VHS_DIR" && vhs "$(basename "$tape")" >/dev/null 2>&1 )
    echo "✔ Finished $name"
}

record_all_tapes() {
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
    rm -f "$VHS_DIR/tuichartdemo"
}

perform() {
    parse_args "$@"
    
    cd "$PROJECT_ROOT"
    mkdir -p "$ASSETS_DIR"
    
    trap cleanup EXIT
    
    build_demo_binary
    record_all_tapes
    
    cleanup
    trap - EXIT
    
    ls -la "$ASSETS_DIR"/demo-*.gif 2>/dev/null | head -5
}

main() {
    local cmd="${1:-help}"
    shift || true
    
    case "$cmd" in
        help|--help|-h)
            usage
            ;;
        perform)
            perform "$@"
            ;;
        *)
            echo "Error: unknown command: $cmd" >&2
            usage >&2
            exit 1
            ;;
    esac
}

main "$@"