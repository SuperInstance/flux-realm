# 🔧 Troubleshooting

Common issues and their fixes when running the Flux Realm examples.

## `go: command not found`

**Cause:** Go is not installed or not on `$PATH`.

**Fix:** Install Go from https://go.dev/dl/ or via package manager:

```bash
# Ubuntu / Debian
sudo apt update && sudo apt install golang-go

# macOS
brew install go

# Verify
go version
```

---

## `python3: command not found`

**Cause:** Python 3 is not installed.

**Fix:**

```bash
# Ubuntu / Debian
sudo apt update && sudo apt install python3 python3-pip

# macOS
brew install python@3.10

# Verify
python3 --version
```

---

## `make: command not found`

**Cause:** Make is not installed.

**Fix:**

```bash
# Ubuntu / Debian
sudo apt update && sudo apt install make

# macOS (Xcode command line tools)
xcode-select --install
# or
brew install make
```

---

## `veto_agent: No such file or directory` during `make demo`

**Cause:** The Go binary wasn't built. This can happen if you skipped `make build`.

**Fix:**

```bash
# Build explicitly, then run
make build
make demo
```

Or just run `make demo` which includes the build step.

---

## `test_trade.py` tests failing

**Cause:** Usually a Python dependency issue or stale `__pycache__`.

**Fix:**

```bash
make clean
python3 -m pytest examples/triangle-trade/test_trade.py -v
```

The tests only use `pytest` and Python stdlib.
Install pytest if missing:

```bash
pip install pytest
```

---

## `go test` fails with "package is not in GOROOT"

**Cause:** Running `go test` from the wrong directory.

**Fix:** Run from inside `examples/triangle-trade/`:

```bash
cd examples/triangle-trade && go test -v ./...
```

Or from the repo root with `make test`.

---

## JSON decode error in `trade_agent.py`

**Cause:** Corrupted or empty bottle file. This can happen if you killed the
orchestrator mid-run.

**Fix:** Clean up and restart:

```bash
cd examples/triangle-trade
rm -f bottle_*.json
bash orchestrator.sh
```

---

## `BATON_READY` not detected (orchestrator hangs)

**Current orchestrator:** The current orchestrator uses direct sequential
execution (no baton polling). If you see agents hang, check:

1. Does `bottle_proposal.json` exist? → Create it manually.
2. Does `trade_agent.py` produce `bottle_conviction.json`? → Run it standalone:
   ```bash
   cd examples/triangle-trade
   python3 trade_agent.py
   ```
3. Does `veto_agent` (Go binary) exist? → `make build`
4. Run each step manually:
   ```bash
   cd examples/triangle-trade
   # Step 1: proposal (already created by orchestrator, or create manually)
   cat > bottle_proposal.json <<EOF
   { "ticker": "TEST", "side": "buy", "quantity": 10, "price": 50.0, "rationale": "Test run" }
   EOF
   # Step 2: Python agent
   python3 trade_agent.py
   # Step 3: Go agent
   ./veto_agent
   # Step 4: Check result
   cat bottle_final.json
   ```

---

## `Permission denied` when running `veto_agent`

**Cause:** The Go binary needs execute permission.

**Fix:**

```bash
chmod +x examples/triangle-trade/veto_agent
```

Or rebuild:

```bash
make build
```

---

## Still stuck?

Open an issue or ask in the project's discussion forum.
Include the output of:

```bash
make clean
make demo 2>&1 | tail -50
```

And your versions:

```bash
echo "Go: $(go version 2>/dev/null || echo 'N/A')"
echo "Python: $(python3 --version 2>/dev/null || echo 'N/A')"
echo "Make: $(make --version 2>/dev/null | head -1 || echo 'N/A')"
```
