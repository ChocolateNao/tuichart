# Default: show available recipes
default:
    @just --list

# Build all packages
build:
    go build ./...

# Run all tests
test:
    go test ./...

# Run tests with race detector
test-race:
    go test -race ./...

# Static analysis
vet:
    go vet ./...

# Run golangci-lint
lint:
    golangci-lint run ./...

# Run golangci-lint and fix all fixable errors
lint-fix:
    golangci-lint run ./... --fix

# Format check (must be clean before commit)
fmt-check:
    @sh -c 'test -z "$(gofmt -l .)" || (gofmt -l . && exit 1)'

# Auto-format all source files
fmt:
    gofumpt -w .

# Run all checks: build, vet, lint, tests
check: build vet lint test fmt-check

# Clean build artifacts
clean:
    rm -rf tmp/ bin/

# Output directory for compiled examples
bin := "bin"
examples_main := "01_sparkline 02_plot 03_bar 04_multi 05_styled 06_live 07_static 08_custom"
examples_mods := "09_bubbletea 10_tview 11_server 12_file 13_gocui"
examples_all := "{{examples_main}} {{examples_mods}}"

# Compile all examples into bin/
examples:
    mkdir -p {{bin}}
    @for ex in {{examples_main}}; do go build -o {{bin}}/$ex ./examples/$ex || exit 1; done
    @for ex in {{examples_mods}}; do (cd examples/$ex && go build -o ../../{{bin}}/$ex .) || exit 1; done
    @echo "built {{bin}}: {{examples_all}}"

# Run a root-module example: just run 07_static
run name="07_static":
    go run ./examples/{{name}}

# Run a nested-module example: just run-mod 09_bubbletea
run-mod name:
    cd examples/{{name}} && go run .

# Run the static demo (all diagram types)
demo:
    go run ./examples/07_static

# Run the live real-time demo
demo-live:
    go run ./examples/06_live

# Run demo without colors
demo-no-color:
    NO_COLOR=1 go run ./examples/07_static

# Run demo in pure ASCII mode
demo-ascii:
    LC_ALL=C NO_COLOR=1 go run ./examples/07_static

# Install golangci-lint (if not present)
setup-lint:
    @which golangci-lint > /dev/null 2>&1 || \
        go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# Install gofumpt (if not present)
setup-fmt:
    @which gofumpt > /dev/null 2>&1 || \
        go install mvdan.cc/gofumpt@latest

# Install all dev tools
setup: setup-lint setup-fmt setup-lefthook

# Set up git hooks via lefthook (run once after clone)
setup-lefthook:
    @which lefthook > /dev/null 2>&1 || \
        go install github.com/evilmartians/lefthook/v2@latest

# Apply lefthook configuration
sync-hooks:
    lefthook install

# Run everything a CI would run
ci: check test-race
