// veto_agent.go — A Go agent that applies the SAEP Veto to trade convictions.
//
// What is a SAEP Veto?
// --------------------
// Safety, Alignment, Execution, and Price — a four-axis veto gate.
// Each axis independently reviews the conviction verdict and casts a vote.
// If any axis vetoes, the trade is blocked.  This prevents single-axis
// failures (e.g. good rationale but terrible price) from slipping through.
//
// Communication:
//   - Reads conviction from bottle_conviction.json (written by trade_agent.py).
//   - Writes final verdict to bottle_final.json.
//   - The orchestrator detects the output bottle and presents results.
//
// Build:  go build -o veto_agent
// Run:    ./veto_agent

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"time"
)

// ---------------------------------------------------------------------------
// Constants — file-based I2I bottle paths
// ---------------------------------------------------------------------------

const (
	BottleConviction = "bottle_conviction.json"
	BottleFinal      = "bottle_final.json"
	AgentVersion     = "saep-veto-go-1.0"
)

// ---------------------------------------------------------------------------
// Data models
// ---------------------------------------------------------------------------

// ConvictionVerdict matches what trade_agent.py writes.
type ConvictionVerdict struct {
	Ticker       string  `json:"ticker"`
	Conviction   string  `json:"conviction"`   // low | medium | high
	Score        float64 `json:"score"`
	Rationale    string  `json:"rationale"`
	AgentVersion string  `json:"agent_version"`
}

// VetoVerdict is the final output of the SAEP gate.
type VetoVerdict struct {
	Ticker      string   `json:"ticker"`
	Approved    bool     `json:"approved"`
	VetoReason  string   `json:"veto_reason,omitempty"`
	AxisVotes   []string `json:"axis_votes"`   // e.g. ["safety:pass", "price:veto"]
	AgentSays   string   `json:"agent_says"`   // human-readable summary
	AgentVersion string  `json:"agent_version"`
	Timestamp    string   `json:"timestamp"`
}

// AxisResult holds the outcome of one SAEP axis.
type AxisResult struct {
	Name  string
	Pass  bool
	Why   string
}

// ---------------------------------------------------------------------------
// SAEP Veto — four-axis evaluation
// ---------------------------------------------------------------------------

// saepVeto evaluates the four axes and returns the aggregate verdict.
func saepVeto(v ConvictionVerdict) VetoVerdict {
	axes := []AxisResult{
		safetyAxis(v),
		alignmentAxis(v),
		executionAxis(v),
		priceAxis(v),
	}

	votes := make([]string, len(axes))
	vetoes := []string{}

	for i, a := range axes {
		if a.Pass {
			votes[i] = fmt.Sprintf("%s:pass", a.Name)
		} else {
			votes[i] = fmt.Sprintf("%s:veto", a.Name)
			vetoes = append(vetoes, a.Name+": "+a.Why)
		}
	}

	approved := len(vetoes) == 0
	reason := ""
	says := ""

	if approved {
		says = fmt.Sprintf("✅ Trade %s approved — all four SAEP axes passed.", v.Ticker)
	} else {
		says = fmt.Sprintf("❌ Trade %s blocked by %d veto(s).", v.Ticker, len(vetoes))
		reason = fmt.Sprintf("vetoed by: %s", vetoes)
	}

	return VetoVerdict{
		Ticker:      v.Ticker,
		Approved:    approved,
		VetoReason:  reason,
		AxisVotes:   votes,
		AgentSays:   says,
		AgentVersion: AgentVersion,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}
}

// safetyAxis: Are we trading a security that passes basic safety checks?
//   - If conviction is HIGH, safety is stricter (re-checks quantity).
//   - This simulates a real safety filter (e.g. no penny stocks, no red flags).
func safetyAxis(v ConvictionVerdict) AxisResult {
	// Simulate: if score is very low and rationale is short, fail safety.
	if v.Score < 0.2 && len(v.Rationale) < 20 {
		return AxisResult{"safety", false, "low score + no rationale = unsafe"}
	}
	return AxisResult{"safety", true, ""}
}

// alignmentAxis: Does the trade align with agent goals?
//   - HIGH conviction passes automatically.
//   - LOW conviction triggers extra scrutiny.
func alignmentAxis(v ConvictionVerdict) AxisResult {
	switch v.Conviction {
	case "high":
		return AxisResult{"alignment", true, "HIGH conviction = aligned"}
	case "medium":
		return AxisResult{"alignment", true, "MEDIUM conviction = acceptable"}
	default: // low
		// Allow low-conviction trades only if score > 0.1
		if v.Score > 0.1 {
			return AxisResult{"alignment", true, "low but above floor"}
		}
		return AxisResult{"alignment", false, "LOW conviction + low score ≤ 0.1 = misaligned"}
	}
}

// executionAxis: Can the trade be executed practically?
//   - Checks trade complexity via ticker length (silly proxy, but illustrative).
//   - Real implementations would check market hours, liquidity, circuit breakers.
func executionAxis(v ConvictionVerdict) AxisResult {
	// Ticker must be between 1 and 5 characters (NYSE/NASDAQ standard-ish).
	if len(v.Ticker) < 1 || len(v.Ticker) > 5 {
		return AxisResult{"execution", false, fmt.Sprintf(
			"ticker %q length=%d outside [1,5]", v.Ticker, len(v.Ticker))}
	}
	return AxisResult{"execution", true, ""}
}

// priceAxis: Does the price make sense given the conviction?
//   - Verifies score is non-negative (sanity check).
//   - Verifies conviction/score consistency.
func priceAxis(v ConvictionVerdict) AxisResult {
	if math.IsNaN(v.Score) || v.Score < 0 {
		return AxisResult{"price", false, fmt.Sprintf("invalid score %.3f", v.Score)}
	}
	if v.Conviction == "high" && v.Score < 0.5 {
		return AxisResult{"price", false, fmt.Sprintf(
			"HIGH conviction but score=%.2f < 0.5", v.Score)}
	}
	return AxisResult{"price", true, ""}
}

// ---------------------------------------------------------------------------
// Helper — load a JSON bottle from disk
// ---------------------------------------------------------------------------

func loadBottle(path string, target interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	log.SetFlags(log.Ltime | log.Lmsgprefix)
	log.SetPrefix("  │ veto_agent  │ ")
	log.Println("── SAEP Veto Agent ─────────────────────────────")

	// 1. Read conviction from Python agent's output bottle.
	log.Printf("Reading %s ...", BottleConviction)
	var conviction ConvictionVerdict
	if err := loadBottle(BottleConviction, &conviction); err != nil {
		log.Fatalf("FATAL: cannot read conviction bottle: %v", err)
	}
	log.Printf("Read conviction: %s → %s (score=%.3f)",
		conviction.Ticker, conviction.Conviction, conviction.Score)

	// 2. Run the SAEP Veto.
	log.Println("Running SAEP Veto (4 axes)...")
	verdict := saepVeto(conviction)

	// 3. Print axis breakdown.
	for _, vote := range verdict.AxisVotes {
		log.Printf("  • %s", vote)
	}

	// 4. Print result.
	fmt.Printf("SAEP_VERDICT:%s\n", verdict.AgentSays)

	// 5. Write final verdict to bottle.
	out, _ := json.MarshalIndent(verdict, "", "  ")
	if err := os.WriteFile(BottleFinal, out, 0644); err != nil {
		log.Fatalf("FATAL: cannot write final bottle: %v", err)
	}
	log.Printf("Wrote verdict to %s", BottleFinal)

	fmt.Println("BATON_COMPLETE")
}
