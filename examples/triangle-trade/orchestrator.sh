#!/usr/bin/env bash
# =============================================================================
# orchestrator.sh — Triangle Trade: Python ⇢ Go A2A pipeline
# =============================================================================
#
# Topology:
#                          ┌──────────────┐
#          proposal ──────►│ TernaryL     │
#          (bottle)        │ Gate (Python)│
#                          └──────┬───────┘
#                                 │ conviction
#                                 ▼
#                          ┌──────────────┐
#                          │ SAEP Veto    │
#                          │ Gate (Go)    │
#                          └──────┬───────┘
#                                 │ approved / blocked
#                                 ▼
#                          ┌──────────────┐
#                          │ Orchestrator │
#                          │ (this script)│
#                          └──────┬───────┘
#                                 │ results to terminal
#
# Why a triangle?  Three legs:
#   1. Trade Proposal (input)
#   2. TernaryL Gate (conviction)
#   3. SAEP Veto (approval)
#
# Three agents, three files, one pipeline.  That's the triangle.
#
# Usage:
#   cd examples/triangle-trade && bash orchestrator.sh
#   # or from repo root: make demo
# =============================================================================

set -euo pipefail

# Colours for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Colour
BOLD='\033[1m'

DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$DIR"

# Clean any previous bottles
rm -f bottle_proposal.json bottle_conviction.json bottle_final.json

echo -e "${BOLD}${CYAN}"
echo "╔══════════════════════════════════════════════════════════╗"
echo "║   Triangle Trade · A2A Pipeline                         ║"
echo "║   Python (TernaryL Gate) → Go (SAEP Veto)               ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# =============================================================================
# STEP 0 — Ensure Go binary exists
# =============================================================================
if [ ! -x "./veto_agent" ]; then
    echo -e "${YELLOW}→ Building veto_agent...${NC}"
    go build -o veto_agent veto_agent.go
    echo -e "${GREEN}✓  Built veto_agent${NC}"
fi

# =============================================================================
# STEP 1 — Inject a trade proposal into the input bottle
# =============================================================================
echo ""
echo -e "${BOLD}Step 1: Creating trade proposal...${NC}"
cat > bottle_proposal.json <<'JSON'
{
  "ticker": "NVDA",
  "side": "buy",
  "quantity": 50,
  "price": 124.50,
  "rationale": "Strong AI tailwinds, recent Blackwell GPU launch, data center revenue growing 200% YoY. Q3 earnings beat consensus by 15%."
}
JSON
echo -e "  ${GREEN}✔${NC} Wrote ${BOLD}bottle_proposal.json${NC}"
cat bottle_proposal.json | python3 -m json.tool --no-ensure-ascii 2>/dev/null || cat bottle_proposal.json

# =============================================================================
# STEP 2 — Run the Python TernaryL gate agent
# =============================================================================
echo ""
echo -e "${BOLD}Step 2: Running TernaryL Gate (Python)...${NC}"
echo -e "  ${BLUE}│${NC}"
python3 trade_agent.py
echo -e "  ${BLUE}│${NC}"
echo -e "  ${GREEN}✔${NC} Conviction written"

echo ""
echo -e "Conviction bottle contents:"
cat bottle_conviction.json | python3 -m json.tool 2>/dev/null || cat bottle_conviction.json

# =============================================================================
# STEP 3 — Run the Go SAEP Veto agent
# =============================================================================
echo ""
echo -e "${BOLD}Step 3: Running SAEP Veto (Go)...${NC}"
echo -e "  ${BLUE}│${NC}"
./veto_agent
echo -e "  ${BLUE}│${NC}"
echo -e "  ${GREEN}✔${NC} Veto complete"

# =============================================================================
# STEP 4 — Display final results
# =============================================================================
echo ""
echo -e "${BOLD}${CYAN}═══════════════════  RESULT  ═══════════════════${NC}"
echo ""

if [ -f bottle_final.json ]; then
    # Pretty-print the final verdict
    python3 -c "
import json, sys
with open('bottle_final.json') as f:
    v = json.load(f)
print(f'  Ticker:     {v[\"ticker\"]}')
print(f'  Approved:   {v[\"approved\"]}')
if v.get('veto_reason'):
    print(f'  Veto:       {v[\"veto_reason\"]}')
print(f'  Axis votes:')
for a in v.get('axis_votes', []):
    print(f'    • {a}')
print(f'  Summary:    {v[\"agent_says\"]}')
print(f'  Timestamp:  {v[\"timestamp\"]}')
"
    echo ""
    echo -e "${BOLD}Full verdict JSON:${NC}"
    python3 -m json.tool bottle_final.json
else
    echo -e "${RED}No final verdict bottle found — pipeline may have failed.${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}${BOLD}✓  Triangle trade pipeline complete.${NC}"
echo ""
echo -e "Files created:"
ls -la bottle_*.json
