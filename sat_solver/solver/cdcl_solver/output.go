package cdcl_solver

/**
 * Get assignments for a variables when we found SAT and want to return satisfying assingment.
 */
func (solver *CDCLSolver) getOutputVariableAssignments() map[string]bool {
	result := make(map[string]bool)
	for k, v := range solver.currentAssignment {
		if !v.IsUndefined() {
			if k < 0 {
				k = -k
			}
			// If the variable was introduced later during optimizations we discard it
			if solver.vars.IsFounderVariable(k) {
				result[solver.vars.Reverse(k)] = TernaryToBool(v)
			}
		}
	}

	return result
}