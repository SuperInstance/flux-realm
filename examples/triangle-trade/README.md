# 🔺 Triangle Trade — Hybrid Manifold A2A Example

## Overview

The **Triangle Trade** is a complete, runnable example of the Hybrid Manifold
paradigm: two agents written in different languages, communicating through
file-based **I2I (Inter-Identity Interface) bottles**, orchestrated by a shell
script.

```
                  ┌──────────────────┐
                  │ Trade Proposal   │
                  │ (bottle_proposal │
                  │  .json)          │
                  └────────┬─────────┘
                           │
                           ▼
                  ┌──────────────────┐
                  │ TernaryL Gate    │
                  │  (Python)        │
                  │                  │
                  │ Low | Med | High │
                  └────────┬─────────┘
                           │ conviction
                           ▼
                  ┌──────────────────┐
                  │ SAEP Veto        │
                  │  (Go)            │
                  │                  │
                  │ 4-axis check     │
                  └────────┬─────────┘
                           │ ✅ / ❌
                           ▼
                  ┌──────────────────┐
                  │ Orchestrator.sh  │
                  │ - creates bottle │
                  │ - runs agents    │
                  │ - shows result   │
                  └──────────────────┘
```

## Why a Triangle?

The name comes from **three legs** — three artifacts — forming the complete
decision pipeline:

| Leg | Component | Language | Role |
|-----|-----------|----------|------|
| 1 | Trade proposal (input) | JSON | Raw data entering the manifold |
| 2 | TernaryL Gate | Python | Conviction scoring (Low/Medium/High) |
| 3 | SAEP Veto | Go | Four-axis safety/alignment check |

Three agents, three files, one pipeline. **That's the triangle.**

## The Gates

### TernaryL Gate (Python)

A **three-way decision gate** that classifies trade proposals by conviction
level. Unlike a binary yes/no filter, TernaryL preserves gradient information:

- **Low** → score < 0.35 — needs more research
- **Medium** → score 0.35–0.70 — plausible but not slam-dunk
- **High** → score ≥ 0.70 — strong conviction

The gate evaluates three factors:
1. **Rationale quality** — how thorough is the reasoning?
2. **Side/price alignment** — buys below $200 get a bonus
3. **Quantity sanity** — retail quantities vs institutional blocks

### SAEP Veto (Go)

A **four-axis veto gate** that checks each dimension independently:

| Axis | What it checks |
|------|---------------|
| **S**afety | Is this trade fundamentally safe? |
| **A**lignment | Does it align with agent goals? |
| **E**xecution | Can it be executed in practice? |
| **P**rice | Does price match conviction? |

If *any* axis vetoes, the trade is blocked. This prevents single-axis failures
from slipping through (e.g., great rationale but terrible price).

## Running It

```bash
# From the repo root
make demo

# Or step-by-step
cd examples/triangle-trade
go build -o veto_agent veto_agent.go
bash orchestrator.sh
```

## Customising

Edit `bottle_proposal.json` (created by `orchestrator.sh`) to try different
trades:

```json
{
  "ticker": "HOOD",
  "side": "sell",
  "quantity": 5000,
  "price": 58.20,
  "rationale": "High retail concentration, P/E of 150+, earnings next week."
}
```

Re-run `bash orchestrator.sh` — the pipeline will evaluate your new proposal.

## Learning Goals

After studying this example, you should understand:

1. ✅ How file-based I2I bottles work for cross-language A2A
2. ✅ The TernaryL gate pattern (three-way conviction routing)
3. ✅ The SAEP Veto pattern (multi-axis safety checks)
4. ✅ How a shell orchestrator can sequence heterogeneous agents
5. ✅ The "triangle" topology as a minimal multi-agent system
