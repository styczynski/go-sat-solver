package cdcl_solver

import (
	"github.com/styczynski/go-sat-solver/sat_solver"
)

/**
 * This is an implementation of Lit Solver::pickBranchLit() from Minisat
 * Browse code here:
 *   https://github.com/JLiangWaterloo/vsids/blob/7f1b4e90bf41eb2e60dbdc5b39334d04b53cf5e2/adaptvsids/core/Solver.cc#L231
 *
 * This function returns a variable that will be selected for another decision.
 * This is crucial for CDCL and can speed up or slow down its search times significantly.
 * Current implementation uses AVSIDS heuristics and falls back to naive selection.
 */
func (solver *CDCLSolver) findNextLiteralForDecision() (sat_solver.CNFLiteral, bool) {
	l := solver.vars.GetAllVariables()

	// Default algorithm: Use AVSIDS suggestions to get variable for decision
	avsidsSuggestion, ok := solver.avsidsSuggestSelect()
	if ok {
		return avsidsSuggestion, true
	}

	// Fallback algorithm: Choose first variable that we can assign
	// Remember that go maps do not have any guarantee on iteration order, so this should give same reliable order, but
	// completely different ones if you run the solver twice.
	// If you want to debug the solver it's recommended to place here a sort function to have a deterministic, known
	// order of suggestions and it's good idea to disable AVSIDS when you are debugging.
	for _, raw := range l {
		if raw < 0 {
			raw = -raw
		}
		if _, ok := solver.currentAssignment[raw]; !ok {
			return raw, true
		}
	}

	return sat_solver.CNF_UNDEFINED, false
}

