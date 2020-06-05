package cdcl_solver

import "github.com/go-sat-solver/sat_solver"

/**
 * Result of the execution of the solver
 */
type SatResult struct {
	// Type of the result
	resultType SatResultType
	// Optionally a variables' assignment leading to SAT
	assgn map[string]bool
}

// Type of the SAT result
type SatResultType int8

const (
	// Solution was not found
	SAT_RESULT_UNDEFINED  SatResultType  = 0
	// Formula cannot be satisfied
	SAT_RESULT_UNSAT      SatResultType  = 1
	// Formula can be satisified
	SAT_RESULT_SAT        SatResultType  = 2
)

/*
 * Return human readable representation of the result
 */
func (result SatResult) String() string {
	switch result.resultType {
	case  SAT_RESULT_UNDEFINED:
		return "Undefined"
	case SAT_RESULT_SAT:
		return "SAT"
	case SAT_RESULT_UNSAT:
		return "UNSAT"
	}
	return "Undefined"
}

/*
 * Return human readable representation of the result
 */
func (result SatResult) Brief() string {
	return result.String()
}

/**
 * Convert result to boolean (SAT = true, anything other = false)
 */
func (result SatResult) ToBool() bool {
	return result.resultType == SAT_RESULT_SAT
}

/**
 * Convert result to boolean (SAT = 1, anything other = 0)
 */
func (result SatResult) ToInt() int {
	if result.ToBool() {
		return 1
	}
	return 0
}

/**
 * Check if result is undefined
 */
func (result SatResult) IsUndefined() bool {
	return result.resultType == SAT_RESULT_UNDEFINED
}

/**
 * Check if result is SAT
 */
func (result SatResult) IsSAT() bool {
	return result.resultType == SAT_RESULT_SAT
}

/**
 * Check if result is USNAT
 */
func (result SatResult) IsUNSAT() bool {
	return result.resultType == SAT_RESULT_UNSAT
}

/**
 * Get assignment that leads to SAT.
 */
func (result SatResult) GetSatisfyingAssignment() map[string]bool {
	return result.assgn
}

/**
 * Structure to store a metadata about the literals values:
 *   - what clause and on what decision level is a cause of the current assignment?
 */
type VariableAssignmentInformation struct {
	// What clause caused the variable assignment?
	reasonClause  sat_solver.CNFClause
	// Decision level when this variable was assigned
	decisionLevel int
}

/**
 * Create new VariableAssignmentInformation object
 */
func NewVariableInformation(solver *CDCLSolver, causeOfAssignment sat_solver.CNFClause) VariableAssignmentInformation {
	return VariableAssignmentInformation{
		reasonClause:  causeOfAssignment,
		decisionLevel: solver.getDecisionLevel(),
	}
}

/**
 * Save the result when we found something.
 */
func (solver *CDCLSolver) foundResult(result SatResult) SatResult {
	if solver.enableDebugLogging {
		solver.context.Trace("result", "Found result %s.", result.String())
	}

	solver.result = result
	return solver.result
}

/**
 * Create new UNDEFINED result
 */
func SatResultUndefined() SatResult {
	return SatResult{
		resultType: SAT_RESULT_UNDEFINED,
		assgn:      map[string]bool{},
	}
}

/**
 * Create new UNSAT result
 */
func SatResultUnsat() SatResult {
	return SatResult{
		resultType: SAT_RESULT_UNSAT,
		assgn:      map[string]bool{},
	}
}

/**
 * Create new SAT result
 */
func SatResultSat(solver *CDCLSolver) SatResult {
	return SatResult{
		resultType: SAT_RESULT_SAT,
		assgn:      solver.getOutputVariableAssignments(),
	}
}