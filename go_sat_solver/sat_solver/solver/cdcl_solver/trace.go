package cdcl_solver

import (
	"fmt"
	"strings"

	"github.com/styczynski/go-sat-solver/sat_solver"
)

type CDCLSolverDecisionTrace struct {
	// This list stores indexes for each decision level
	// For example let's suppose that:
	//   assignmentTrace = [ -12, 13, 9, -5, -2, -1 ]
	// Where those numbers identify variables (negative numbers are negated variables).
	// So in that context decisionTrace[1] = 2 means that the decision on level 1 starts in assignmentTrace on index 2,
	// Let's assume decisionTrace[2] = 4
	// so the decision node has clause [9 v not(5)]
	//
	decisionTrace          []int
	// Current assignment of variables
	currentAssignment      map[sat_solver.CNFLiteral]Ternary
	// Meta information attached to currently assigned variables
	// Mosty information how we assigned those variables
	varsInfo               map[sat_solver.CNFLiteral]VariableAssignmentInformation
	// Assignment trace is a list of literals captured when decision was made
	assignmentTrace        []sat_solver.CNFLiteral
}

// Get current value for a literal
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
		result = result.Negate()
	}

	return result
}

/**
 * Create new decision for a given literal.
 */
func (solver *CDCLSolver) newDecision(literal sat_solver.CNFLiteral) {
	if solver.enableDebugLogging {
		solver.context.Trace("decide", "Create new decision for %s (%s)", literal.String(solver.vars), literal.DebugString())
	}
	solver.decisionTrace = append(solver.decisionTrace, len(solver.assignmentTrace))
	solver.performLiteralAssertion(literal, nil)
}

/**
 * Assert the literal.
 * This function is used when we encounter unit clause to force the variable value
 * or when creating decisions about a variable. If the from is nil that means arbitrary decision coming from select.go
 * If it's not empty then the clause means that the assignment was forces by a occurring conflict.
 *
 */
func (solver *CDCLSolver) performLiteralAssertion(literal sat_solver.CNFLiteral, from sat_solver.CNFClause) {
	v := literal
	if v < 0 {
		v = -v
	}
	solver.varsInfo[v] = NewVariableInformation(solver, from)
	solver.assignmentTrace = append(solver.assignmentTrace, literal)
	solver.currentAssignment[v] = BoolToTernary(literal >= 0)
}

/**
 * Determine the decision level that the variable is assigned.
 */
func (solver *CDCLSolver) getDecisionLevelForVar(v sat_solver.CNFLiteral) int {
	return solver.varsInfo[v].decisionLevel
}

/**
 * Get current decision level
 */
func (solver *CDCLSolver) getDecisionLevel() int {
	return len(solver.decisionTrace)
}

/**
 * Go back to the given decision level
 */
func (solver *CDCLSolver) reverseToDecisionLevel(decisionLevel int) {
	if solver.enableDebugLogging {
		solver.context.Trace("reverse", "Jumping back to getDecisionLevel %d.", decisionLevel)
	}

	if solver.getDecisionLevel() <= decisionLevel {
		return
	}

	lastDecisionIndex := solver.decisionTrace[decisionLevel]

	// Unassign anything in the assignmentTrace in higher levels
	for i := len(solver.assignmentTrace) - 1; i >= lastDecisionIndex; i-- {
		trailVar := solver.assignmentTrace[i]
		if trailVar < 0 {
			trailVar = -trailVar
		}
		delete(solver.currentAssignment, trailVar)
	}

	// Remove values from the trace
	solver.decisionTrace = solver.decisionTrace[:decisionLevel]
	solver.assignmentTrace = solver.assignmentTrace[:lastDecisionIndex]

	// Update the index for a checked decision levels
	// This means that we notify propagation algorithm at what level it should start
	solver.currentTraceCheckIndex = lastDecisionIndex
}

/**
 * Get human-readable string describing the current solver decision trace.
 */
func (solver *CDCLSolver) getDecisionTraceString() string {
	rows := []string{}
	for i, l := range solver.assignmentTrace {
		decision := ""
		for _, decisionID := range solver.decisionTrace {
			if decisionID == i {
				decision = "> "
				break
			}
		}
		rows = append(rows, fmt.Sprintf("%s%s", decision, l.DebugString()))
	}
	return fmt.Sprintf("[%s]", strings.Join(rows, ", "))
}
