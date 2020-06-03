package cdcl_solver

import "github.com/go-sat-solver/sat_solver"

// SatResult is an enum type for the state of the SAT solver.
type SatResult byte

const (
	SAT_RESULT_UNDEFINED  SatResult  = 0
	SAT_RESULT_UNSAT      SatResult  = 1
	SAT_RESULT_SAT        SatResult  = 2
)

func (result SatResult) String() string {
	switch result {
	case  SAT_RESULT_UNDEFINED:
		return "Undefined"
	case SAT_RESULT_SAT:
		return "SAT"
	case SAT_RESULT_UNSAT:
		return "UNSAT"
	}
	return "Undefined"
}

func (result SatResult) Brief() string {
	return result.String()
}

func (result SatResult) ToBool() bool {
	return result == SAT_RESULT_SAT
}

// VariableAssignmentInformation just stores some basic information about assigned variables
type VariableAssignmentInformation struct {
	reasonClause  sat_solver.CNFClause // reasonClause is the clause that caused this assignment
	decisionLevel int                  // getDecisionLevel is the decision getDecisionLevel of this assignment
}

func NewVariableInformation(solver *CDCLSolver, causeOfAssignment sat_solver.CNFClause) VariableAssignmentInformation {
	return VariableAssignmentInformation{
		reasonClause:  causeOfAssignment,
		decisionLevel: solver.getDecisionLevel(),
	}
}

func (solver *CDCLSolver) foundResult(result SatResult) SatResult {
	if solver.enableDebugLogging {
		solver.context.Trace("result", "Found result %#v.", result)
	}

	solver.result = result
	return result
}