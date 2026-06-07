package main

import (
	"testing"
)

// ---------------------------------------------------------------------------
// SAEP Veto tests
// ---------------------------------------------------------------------------

func TestSafetyAxisPass(t *testing.T) {
	v := ConvictionVerdict{Ticker: "AAPL", Conviction: "high", Score: 0.85, Rationale: "Solid earnings growth and strong balance sheet."}
	result := safetyAxis(v)
	if !result.Pass {
		t.Errorf("expected pass, got veto: %s", result.Why)
	}
}

func TestSafetyAxisFail(t *testing.T) {
	v := ConvictionVerdict{Ticker: "PENN", Conviction: "low", Score: 0.05, Rationale: "idk"}
	result := safetyAxis(v)
	if result.Pass {
		t.Errorf("expected veto for low score + no rationale")
	}
}

func TestAlignmentHighConviction(t *testing.T) {
	v := ConvictionVerdict{Ticker: "NVDA", Conviction: "high", Score: 0.90, Rationale: ""}
	result := alignmentAxis(v)
	if !result.Pass {
		t.Errorf("HIGH conviction should pass alignment")
	}
}

func TestAlignmentLowConvictionBelowFloor(t *testing.T) {
	v := ConvictionVerdict{Ticker: "SNAP", Conviction: "low", Score: 0.08, Rationale: "meh"}
	result := alignmentAxis(v)
	if result.Pass {
		t.Errorf("LOW conviction score=0.08 should trigger veto")
	}
}

func TestAlignmentLowConvictionAboveFloor(t *testing.T) {
	v := ConvictionVerdict{Ticker: "UBER", Conviction: "low", Score: 0.15, Rationale: "Possible recovery play."}
	result := alignmentAxis(v)
	if !result.Pass {
		t.Errorf("LOW conviction score=0.15 should pass floor")
	}
}

func TestExecutionAxisValidTicker(t *testing.T) {
	v := ConvictionVerdict{Ticker: "GOOG", Conviction: "high", Score: 0.80, Rationale: ""}
	result := executionAxis(v)
	if !result.Pass {
		t.Errorf("GOOG is a valid ticker")
	}
}

func TestExecutionAxisEmptyTicker(t *testing.T) {
	v := ConvictionVerdict{Ticker: "", Conviction: "low", Score: 0.0, Rationale: ""}
	result := executionAxis(v)
	if result.Pass {
		t.Errorf("empty ticker should veto")
	}
}

func TestExecutionAxisLongTicker(t *testing.T) {
	v := ConvictionVerdict{Ticker: "ABCDEF", Conviction: "medium", Score: 0.5, Rationale: ""}
	result := executionAxis(v)
	if result.Pass {
		t.Errorf("ticker longer than 5 chars should veto")
	}
}

func TestPriceAxisNaN(t *testing.T) {
	v := ConvictionVerdict{Ticker: "HACK", Conviction: "high", Score: -1.0, Rationale: ""}
	result := priceAxis(v)
	if result.Pass {
		t.Errorf("negative score should veto")
	}
}

func TestPriceAxisHighConvictionLowScore(t *testing.T) {
	v := ConvictionVerdict{Ticker: "DOGE", Conviction: "high", Score: 0.40, Rationale: "moon"}
	result := priceAxis(v)
	if result.Pass {
		t.Errorf("HIGH conviction but score=0.40 should veto price axis")
	}
}

func TestPriceAxisPass(t *testing.T) {
	v := ConvictionVerdict{Ticker: "META", Conviction: "high", Score: 0.85, Rationale: "Strong ad revenue growth."}
	result := priceAxis(v)
	if !result.Pass {
		t.Errorf("HIGH conviction with score=0.85 should pass")
	}
}

// ---------------------------------------------------------------------------
// Full SAEP Veto integration
// ---------------------------------------------------------------------------

func TestSaepVetoAllPass(t *testing.T) {
	v := ConvictionVerdict{
		Ticker:     "MSFT",
		Conviction: "high",
		Score:      0.82,
		Rationale:  "Azure growth accelerating, Copilot driving enterprise adoption. Strong financials.",
	}
	verdict := saepVeto(v)
	if !verdict.Approved {
		t.Errorf("expected approved, got veto: %v", verdict.VetoReason)
	}
	if len(verdict.AxisVotes) != 4 {
		t.Errorf("expected 4 axis votes, got %d", len(verdict.AxisVotes))
	}
}

func TestSaepVetoBlocked(t *testing.T) {
	v := ConvictionVerdict{
		Ticker:     "",
		Conviction: "low",
		Score:      0.05,
		Rationale:  "",
	}
	verdict := saepVeto(v)
	if verdict.Approved {
		t.Errorf("expected blocked for multiple violations")
	}
}

func TestSaepVetoIntermediate(t *testing.T) {
	// Only execution axis should veto (empty ticker).
	v := ConvictionVerdict{
		Ticker:     "",
		Conviction: "high",
		Score:      0.90,
		Rationale:  "Great rationale, but ticker is empty...",
	}
	verdict := saepVeto(v)
	if verdict.Approved {
		t.Errorf("empty ticker should block")
	}
	// But should still have some passes
	hasPass := false
	for _, vote := range verdict.AxisVotes {
		if len(vote) > 5 && vote[len(vote)-4:] == "pass" {
			hasPass = true
		}
	}
	if !hasPass {
		t.Errorf("should have at least one passing axis (e.g. alignment)")
	}
}
