#!/usr/bin/env python3
"""
trade_agent.py — A Python agent that routes trade proposals through a
TernaryL gate.

What is a TernaryL gate?
------------------------
A three-way decision gate that classifies inputs into Low, Medium, or High
conviction buckets.  Unlike a simple binary filter, TernaryL preserves
gradient information so downstream agents (like the Veto) can apply
different thresholds.

Topology (the "triangle"):
  1.  Trade proposal enters via bottle (JSON file on disk).
  2.  TernaryL gate evaluates conviction.
  3.  Conviction + metadata written to an output bottle.
  4.  Baton passed to the Go veto agent via orchestrator.sh.

Usage:
    python3 trade_agent.py
"""

from __future__ import annotations

import json
import os
import sys
import logging
from dataclasses import dataclass, asdict, field
from enum import Enum
from typing import Optional

# ---------------------------------------------------------------------------
# Logging — human-readable pipeline progress
# ---------------------------------------------------------------------------
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s │ %(name)-18s │ %(levelname)-5s │ %(message)s",
    datefmt="%H:%M:%S",
)
log = logging.getLogger("trade_agent")


# ---------------------------------------------------------------------------
# Constants — file-based I2I bottle paths
# ---------------------------------------------------------------------------
BOTTLE_IN  = "bottle_proposal.json"      # input: trade proposal
BOTTLE_OUT = "bottle_conviction.json"    # output: TernaryL result


# ---------------------------------------------------------------------------
# Data models
# ---------------------------------------------------------------------------

class Conviction(Enum):
    """TernaryL conviction levels supported by the gate."""
    LOW    = "low"
    MEDIUM = "medium"
    HIGH   = "high"

    @property
    def score(self) -> float:
        return {"low": 0.15, "medium": 0.55, "high": 0.90}[self.value]


@dataclass
class TradeProposal:
    """What the user wants to trade."""
    ticker: str
    side: str                  # "buy" | "sell"
    quantity: int
    price: float
    rationale: str = ""        # why the agent thinks this is a good trade

    @classmethod
    def from_bottle(cls, path: str) -> "TradeProposal":
        with open(path) as f:
            raw = json.load(f)
        return cls(**raw)


@dataclass
class ConvictionVerdict:
    """Output of the TernaryL gate."""
    ticker: str
    conviction: str            # low | medium | high
    score: float
    rationale: str
    agent_version: str = "ternaryl-py-1.0"

    def to_bottle(self, path: str) -> None:
        with open(path, "w") as f:
            json.dump(asdict(self), f, indent=2)


# ---------------------------------------------------------------------------
# TernaryL gate — the core decision function
# ---------------------------------------------------------------------------

def ternaryl_gate(proposal: TradeProposal) -> ConvictionVerdict:
    """
    Evaluate a trade proposal and return a conviction verdict.

    The gate considers three factors:
      1.  **Rationale quality** — longer rationales hint at more research.
      2.  **Side alignment** — buys of cheap-ish stocks get a bonus.
      3.  **Quantity sanity** — very large quantities lower conviction
          (slippage / market impact risk).

    Each factor contributes to a score in [0, 1]; the final score
    selects the TernaryL bucket.
    """
    score = 0.0
    parts: list[str] = []

    # --- Factor 1: Rationale quality (0 – 0.4) ---
    rlen = len(proposal.rationale.strip())
    if rlen > 80:
        score += 0.35
        parts.append("rationale: strong (+0.35)")
    elif rlen > 30:
        score += 0.20
        parts.append("rationale: adequate (+0.20)")
    else:
        score += 0.05
        parts.append("rationale: weak (+0.05)")

    # --- Factor 2: Side & price (0 – 0.3) ---
    if proposal.side == "buy" and proposal.price < 200:
        score += 0.25
        parts.append("buy-below-200 (+0.25)")
    elif proposal.side == "buy":
        score += 0.10
        parts.append("buy-premium (+0.10)")
    else:  # sell
        score += 0.20
        parts.append("sell (+0.20)")

    # --- Factor 3: Quantity sanity (0 – 0.3) ---
    # Very large quantities introduce slippage risk.
    if proposal.quantity <= 100:
        score += 0.30
        parts.append("qty: retail (+0.30)")
    elif proposal.quantity <= 1000:
        score += 0.15
        parts.append("qty: moderate (+0.15)")
    else:
        score += 0.05
        parts.append("qty: large (+0.05)")

    # --- Bucket into TernaryL conviction ---
    if score >= 0.70:
        conviction = Conviction.HIGH
    elif score >= 0.35:
        conviction = Conviction.MEDIUM
    else:
        conviction = Conviction.LOW

    log.info(
        "TernaryL gate → %s (score=%.2f)  │ %s",
        conviction.value.upper(),
        score,
        ", ".join(parts),
    )

    return ConvictionVerdict(
        ticker=proposal.ticker,
        conviction=conviction.value,
        score=round(score, 3),
        rationale=proposal.rationale,
    )


# ---------------------------------------------------------------------------
# Main — read from bottle, evaluate, write to bottle, exit
# ---------------------------------------------------------------------------

def main() -> None:
    log.info("── TernaryL Trade Agent ──────────────────────────")

    # 1. Wait for / read the input bottle
    if not os.path.exists(BOTTLE_IN):
        log.error("No proposal bottle found at %s", BOTTLE_IN)
        log.error("Run orchestrator.sh to create the bottle chain.")
        sys.exit(1)

    proposal = TradeProposal.from_bottle(BOTTLE_IN)
    log.info("Received proposal: %s %d × %s @ $%.2f",
             proposal.side, proposal.quantity, proposal.ticker, proposal.price)

    # 2. Run the TernaryL gate
    verdict = ternaryl_gate(proposal)

    # 3. Write conviction to output bottle
    verdict.to_bottle(BOTTLE_OUT)
    log.info("Wrote conviction to %s", BOTTLE_OUT)

    # 4. Signal readiness by writing to stderr (parsed by orchestrator)
    print(f"BATON_READY:{BOTTLE_OUT}", flush=True)
    log.info("Baton ready → %s", BOTTLE_OUT)


if __name__ == "__main__":
    main()
