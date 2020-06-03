package cdcl_solver

import (
	"fmt"
	"strings"

	"github.com/go-sat-solver/sat_solver"
)

// This file contains the assignmentTrace-related functions for the solver.

// Assignments returns the assigned variables and their value (true or false).
// This is only valid if Solve returned true, in which case this is the
// solution.
func (solver *CDCLSolver) Assignments() map[sat_solver.CNFLiteral]bool {
	result := make(map[sat_solver.CNFLiteral]bool)
	for k, v := range solver.currentAssignment {
		if v != TERNARY_UNDEFINED {
			result[k] = v == TERNARY_TRUE
		}
	}

	return result
}

// ValueLit reads the currently set value for a literal.
func (solver *CDCLSolver) currentLiteralValue(l sat_solver.CNFLiteral) Ternary {
	v := l
	if v < 0 {
		v = -v
	}
	result, ok := solver.currentAssignment[v]
	if !ok || result == TERNARY_UNDEFINED {
		return TERNARY_UNDEFINED
	}

	// If the literal is negative (signed), then XOR 1 will cause the bool
	// to flip. If result is undef, this has no affect.
	if l < 0 {
		result = - result
	}

	return result
}

func (solver *CDCLSolver) newDecision(literal sat_solver.CNFLiteral) {
	if solver.enableDebugLogging {
		solver.context.Trace("decide", "Create new decision for %s", literal.String(solver.vars))
	}
	solver.decisionTrace = append(solver.decisionTrace, len(solver.assignmentTrace))
	solver.performLiteralAssertion(literal, nil)
}

func (solver *CDCLSolver) performLiteralAssertion(literal sat_solver.CNFLiteral, from sat_solver.CNFClause) {
	v := literal
	if v < 0 {
		v = -v
	}
	solver.varsInfo[v] = NewVariableInformation(solver, from)
	solver.assignmentTrace = append(solver.assignmentTrace, literal)
	solver.currentAssignment[v] = BoolToTernary(literal > 0)
}

func (solver *CDCLSolver) getDecisionLevelForVar(v sat_solver.CNFLiteral) int {
	return solver.varsInfo[v].decisionLevel
}

func (solver *CDCLSolver) getDecisionLevel() int {
	return len(solver.decisionTrace)
}

// reverseToDecisionLevel trims the assignmentTrace down to the given getDecisionLevel (including
// that getDecisionLevel).
func (solver *CDCLSolver) reverseToDecisionLevel(decisionLevel int) {
	if solver.enableDebugLogging {
		solver.context.Trace("reverse", "Jumping back to getDecisionLevel %d.", decisionLevel)
	}

	if solver.getDecisionLevel() <= decisionLevel {
		return
	}

	lastIdx := solver.decisionTrace[decisionLevel]

	// Unassign anything in the assignmentTrace in higher levels
	for i := len(solver.assignmentTrace) - 1; i >= lastIdx; i-- {
		trailVar := solver.assignmentTrace[i]
		if trailVar < 0 {
			trailVar = -trailVar
		}
		delete(solver.currentAssignment, trailVar)
	}

	// Update our queue head
	solver.currentTraceCheckIndex = lastIdx

	// Reset the assignmentTrace length
	solver.assignmentTrace = solver.assignmentTrace[:lastIdx]
	solver.decisionTrace = solver.decisionTrace[:decisionLevel]
}

// getDecisionTraceString is used for debugging
func (solver *CDCLSolver) getDecisionTraceString() string {
	vs := make([]string, len(solver.assignmentTrace))
	for i, l := range solver.assignmentTrace {
		decision := ""
		for _, idx := range solver.decisionTrace {
			if idx == i {
				decision = "| "
				break
			}
		}

		vs[i] = fmt.Sprintf("%s%s", decision, l.String(solver.vars))
	}

	return fmt.Sprintf("[%s]", strings.Join(vs, ", "))
}
