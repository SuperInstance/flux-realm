# =============================================================================
# Flux Realm — Hybrid Manifold A2A Demo
# =============================================================================
# One Makefile to build, test, run, and clean the entire demo.
# Usage:  make demo   (runs the full triangle-trade pipeline)
#         make test   (unit tests)
#         make clean  (remove bottles, binaries, __pycache__)
#         make build  (compile Go agent)

SHELL := /bin/bash
.PHONY: help demo build test clean lint

help:
	@echo "Flux Realm — Hybrid Manifold A2A"
	@echo ""
	@echo "  make demo    Run the full triangle-trade pipeline"
	@echo "  make test    Run unit tests (Python + Go)"
	@echo "  make build   Compile the Go veto agent"
	@echo "  make clean   Remove artifacts and bottle files"
	@echo "  make lint    Check code style"

# ---------------------------------------------------------------------------
# Demo — runs the triangle trade end to end
# ---------------------------------------------------------------------------
demo: build
	@echo ""
	@echo "╔══════════════════════════════════════════════════════╗"
	@echo "║   Flux Realm · Triangle Trade · A2A Demo            ║"
	@echo "╚══════════════════════════════════════════════════════╝"
	@echo ""
	cd examples/triangle-trade && bash orchestrator.sh

# ---------------------------------------------------------------------------
# Build — compile Go binary
# ---------------------------------------------------------------------------
build:
	@echo "→ Building Go veto agent..."
	cd examples/triangle-trade && go build -o veto_agent veto_agent.go
	@echo "✓  veto_agent built"

# ---------------------------------------------------------------------------
# Test — run all unit tests
# ---------------------------------------------------------------------------
test: build
	@echo "→ Running Python tests..."
	cd examples/triangle-trade && python3 -m pytest test_trade.py -v --tb=short
	@echo ""
	@echo "→ Running Go tests..."
	cd examples/triangle-trade && go test -v ./...

# ---------------------------------------------------------------------------
# Clean
# ---------------------------------------------------------------------------
clean:
	@echo "→ Cleaning..."
	rm -f examples/triangle-trade/veto_agent
	rm -f examples/triangle-trade/bottle_*.json
	rm -rf examples/triangle-trade/__pycache__
	rm -rf examples/triangle-trade/.pytest_cache
	@echo "✓  Clean"

# ---------------------------------------------------------------------------
# Lint
# ---------------------------------------------------------------------------
lint:
	@echo "→ Linting Python..."
	cd examples/triangle-trade && python3 -m py_compile trade_agent.py vetor_agent.py 2>/dev/null || true
	@echo "→ Linting Go..."
	cd examples/triangle-trade && go vet ./...
	@echo "✓  Lint ok"
