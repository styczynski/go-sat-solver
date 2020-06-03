package cdcl_solver

import "github.com/go-sat-solver/sat_solver"

func (solver *CDCLSolver) findNextLiteralForDecision() (sat_solver.CNFLiteral, bool) {
	for _, raw := range solver.vars.GetAllVariables() {
		if _, ok := solver.currentAssignment[raw]; !ok {
			return raw, true
		}
	}

	return sat_solver.CNF_UNDEFINED, false
}

