"""
Tests for the TernaryL gate trade agent.

These tests verify the gate logic without touching I/O:
- Each conviction bucket (low/medium/high) is reachable.
- The TradeProposal deserializer works from JSON.
- Edge cases (empty rationale, huge quantity, premium price).
"""

import json
import pytest
from trade_agent import (
    TradeProposal,
    ConvictionVerdict,
    ternaryl_gate,
    Conviction,
    BOTTLE_IN,
    BOTTLE_OUT,
)


# ==============================================================================
# TradeProposal deserialization
# ==============================================================================

def test_proposal_from_dict():
    """TradeProposal can be constructed from a dict (simulating JSON load)."""
    p = TradeProposal(
        ticker="AAPL",
        side="buy",
        quantity=100,
        price=180.0,
        rationale="Decent fundamentals."
    )
    assert p.ticker == "AAPL"
    assert p.side == "buy"


def test_proposal_from_bottle(tmp_path):
    """TradeProposal.from_bottle reads a JSON file correctly."""
    path = tmp_path / "test_proposal.json"
    data = {
        "ticker": "TSLA",
        "side": "sell",
        "quantity": 50,
        "price": 250.0,
        "rationale": "Test rationale here."
    }
    path.write_text(json.dumps(data))

    p = TradeProposal.from_bottle(str(path))
    assert p.ticker == "TSLA"
    assert p.quantity == 50


# ==============================================================================
# TernaryL gate — conviction buckets
# ==============================================================================

def test_high_conviction():
    """A well-researched buy under $200 with retail qty → HIGH conviction."""
    p = TradeProposal(
        ticker="NVDA",
        side="buy",
        quantity=50,
        price=124.50,
        rationale="Strong AI tailwinds, Blackwell GPU launch, data center "
                  "revenue growing 200% YoY. Q3 earnings beat consensus.",
    )
    v = ternaryl_gate(p)
    assert v.conviction == "high"
    assert v.score >= 0.70


def test_medium_conviction():
    """A moderate proposal with adequate rationale → MEDIUM conviction."""
    p = TradeProposal(
        ticker="MSFT",
        side="buy",
        quantity=500,
        price=350.0,
        rationale="Decent cloud growth and expanding margins.",
    )
    v = ternaryl_gate(p)
    assert v.conviction == "medium"
    assert 0.35 <= v.score < 0.70


def test_low_conviction():
    """A sell with huge qty and no rationale → LOW conviction."""
    p = TradeProposal(
        ticker="GME",
        side="sell",
        quantity=9999,
        price=25.0,
        rationale="",
    )
    v = ternaryl_gate(p)
    assert v.conviction == "low"
    assert v.score < 0.35


def test_edge_short_ticker():
    """Single-char tickers work fine."""
    p = TradeProposal(ticker="X", side="buy", quantity=10, price=5.0, rationale="")
    v = ternaryl_gate(p)
    assert isinstance(v, ConvictionVerdict)
    assert v.ticker == "X"


# ==============================================================================
# ConvictionVerdict serialization
# ==============================================================================

def test_verdict_roundtrip(tmp_path):
    """ConvictionVerdict → JSON → dict preserves all fields."""
    v = ConvictionVerdict(
        ticker="AMD",
        conviction="medium",
        score=0.55,
        rationale="Zen 5 looks promising.",
    )
    path = tmp_path / "test_verdict.json"
    v.to_bottle(str(path))

    loaded = json.loads(path.read_text())
    assert loaded["ticker"] == "AMD"
    assert loaded["conviction"] == "medium"
    assert loaded["score"] == 0.55


# ==============================================================================
# Conviction enum
# ==============================================================================

def test_conviction_scores():
    """Each Conviction level has an increasing score."""
    assert Conviction.LOW.score < Conviction.MEDIUM.score
    assert Conviction.MEDIUM.score < Conviction.HIGH.score
