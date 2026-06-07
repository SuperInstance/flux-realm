# ⚡ Flux Realm — From Zero to A2A in 5 Minutes

This guide gets you from a bare machine to running a cross-language
Agent-to-Agent (A2A) pipeline using the Hybrid Manifold pattern.

## Prerequisites

| Tool   | Min. Version | Check command      | Install (if missing)                       |
|--------|-------------|--------------------|--------------------------------------------|
| Go     | 1.21+       | `go version`       | https://go.dev/dl/                         |
| Python | 3.10+       | `python3 --version`| https://www.python.org/downloads/          |
| Make   | 4.0+        | `make --version`   | `apt install make` / `brew install make`   |

That's it. No Docker, no Kubernetes, no orchestrator cluster — just files,
agents, and a shell.

---

## Step 1: Clone / Navigate

```bash
# If you just downloaded the repo:
cd flux-realm
```

If you're in the repo already (`ls` shows `Makefile`, `QUICKSTART.md`):
```bash
# You're good.
```

---

## Step 2: Run the Demo

```bash
make demo
```

That single command will:

1. ✅ **Build** the Go veto agent (`go build`)
2. ✅ **Create** a trade proposal (JSON bottle)
3. ✅ **Run** the Python TernaryL gate agent
4. ✅ **Run** the Go SAEP Veto agent
5. ✅ **Display** the final verdict

Expected output (abbreviated):

```
╔══════════════════════════════════════════════════════╗
║   Triangle Trade · A2A Pipeline                     ║
║   Python (TernaryL Gate) → Go (SAEP Veto)           ║
╚══════════════════════════════════════════════════════╝

Step 1: Creating trade proposal...
  ✔ Wrote bottle_proposal.json
  {
    "ticker": "NVDA",
    "side": "buy",
    "quantity": 50,
    "price": 124.50,
    ...
  }

Step 2: Running TernaryL Gate (Python)...
  10:00:01 │ trade_agent        │ INFO  │ ── TernaryL Trade Agent ──
  10:00:01 │ trade_agent        │ INFO  │ Received proposal: buy 50 × NVDA @ $124.50
  10:00:01 │ trade_agent        │ INFO  │ TernaryL gate → HIGH (score=0.90)

Step 3: Running SAEP Veto (Go)...
  10:00:01 │ veto_agent  │ Reading bottle_conviction.json ...
  10:00:01 │ veto_agent  │ Running SAEP Veto (4 axes)...
  10:00:01 │ veto_agent  │   • safety:pass
  10:00:01 │ veto_agent  │   • alignment:pass
  10:00:01 │ veto_agent  │   • execution:pass
  10:00:01 │ veto_agent  │   • price:pass

═══════════════════════  RESULT  ═══════════════════════

  Ticker:     NVDA
  Approved:   true
  Axis votes:
    • safety:pass
    • alignment:pass
    • execution:pass
    • price:pass
  Summary:    ✅ Trade NVDA approved — all four SAEP axes passed.

✓  Triangle trade pipeline complete.
```

---

## Step 3: Run the Tests

```bash
make test
```

This runs both the Python and Go test suites:

```
→ Running Python tests...
  test_trade.py::test_high_conviction PASSED
  test_trade.py::test_medium_conviction PASSED
  test_trade.py::test_low_conviction PASSED
  ...

→ Running Go tests...
  ok   triangle-trade  0.123s
```

---

## Step 4: Customise

Try different trades to see how the gates respond:

```bash
# Edit the proposal directly, then re-run the orchestrator
cat > examples/triangle-trade/bottle_proposal.json <<'EOF'
{
  "ticker": "SNAP",
  "side": "sell",
  "quantity": 10000,
  "price": 12.40,
  "rationale": ""
}
EOF
cd examples/triangle-trade && bash orchestrator.sh
```

This should produce a **blocked** trade (LOW conviction + empty rationale →
safety veto + alignment veto).

---

## What Just Happened?

You ran a **Hybrid Manifold** A2A pipeline:

| Component | Language | Role |
|-----------|----------|------|
| `trade_agent.py` | Python | TernaryL gate → conviction scoring |
| `veto_agent.go` | Go | SAEP Veto → 4-axis approval |
| `orchestrator.sh` | Bash | Sequence, bottles, presentation |

The agents communicated through **file-based I2I bottles** (JSON files on disk).
No HTTP, no message queue, no shared database — just a file system contract.

---

## Next Steps

1. Read `examples/triangle-trade/README.md` — deep dive on the topology
2. Read `trade_agent.py` — TernaryL gate implementation
3. Read `veto_agent.go` — SAEP Veto implementation
4. Try `make clean && make test` to verify everything from scratch
5. Fork this repo and add your own agent (Rust? Node.js? C#?)
