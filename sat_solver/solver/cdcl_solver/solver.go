package cdcl_solver

import (
	"fmt"

	"github.com/go-sat-solver/sat_solver"
	"github.com/go-sat-solver/sat_solver/solver"
)

type CDCLSolver struct {
	processID              uint
	enableDebugLogging     bool
	context                *sat_solver.SATContext
	clauses                []sat_solver.CNFClause // clauses to solve
	vars                   *sat_solver.SATVariableMapping
	currentTraceCheckIndex int
	watchedLiterals        map[sat_solver.CNFLiteral][]*TWLRecord
	currentLearnedClause   sat_solver.CNFClause
	assignmentTrace        []sat_solver.CNFLiteral
	decisionTrace          []int
	visited                map[sat_solver.CNFLiteral]bool
	currentAssignment      map[sat_solver.CNFLiteral]Ternary
	varsInfo               map[sat_solver.CNFLiteral]VariableAssignmentInformation
	result                 SatResult
}

func NewCDCLSolver() *CDCLSolver {
	return &CDCLSolver{
		result:               SAT_RESULT_UNDEFINED,
		currentAssignment:    map[sat_solver.CNFLiteral]Ternary{},
		varsInfo:             map[sat_solver.CNFLiteral]VariableAssignmentInformation{},
		watchedLiterals:      map[sat_solver.CNFLiteral][]*TWLRecord{},
		visited:              map[sat_solver.CNFLiteral]bool{},
		currentLearnedClause: make([]sat_solver.CNFLiteral, 10),
	}
}

func (solver *CDCLSolver) Solve(formula *sat_solver.SATFormula, context *sat_solver.SATContext) (error, solver.SolverResult) {
	solver.context = context
	//solver.enableDebugLogging = context.IsSolverTracingEnabled()
	if f, ok := formula.Formula().(*sat_solver.CNFFormula); ok {
		solver.vars = formula.Variables()
		for _, newClause := range f.Variables {
			if len(newClause) == 0 {
				solver.result = SAT_RESULT_UNSAT
				continue
			} else if len(newClause) == 1 {
				solver.performLiteralAssertion(newClause[0], nil)
				solver.performUnitPropagation()
				continue
			} else {
				solver.clauses = append(solver.clauses, newClause)
				solver.watchClause(newClause)
			}
		}

		if solver.enableDebugLogging {
			solver.context.Trace("start", "Started solver.")
		}

		if solver.result != SAT_RESULT_UNDEFINED {
			return nil, solver.foundResult(solver.result)
		}

		for {
			conflictingClause := solver.performUnitPropagation()
			if conflictingClause == nil {
				lit, hasAnyLiterals := solver.findNextLiteralForDecision()

				if !hasAnyLiterals {
					return nil, solver.foundResult(SAT_RESULT_SAT)
				}

				solver.newDecision(lit)
			} else {
				if solver.enableDebugLogging {
					solver.context.Trace("conflict", "Conflicting clause detected on unit propagation. Decision trace: %s.", solver.getDecisionTraceString())
				}
				if solver.getDecisionLevel() == 0 {
					return nil, solver.foundResult(SAT_RESULT_UNSAT)
				}

				newLevel := solver.learnClause(conflictingClause)
				solver.reverseToDecisionLevel(newLevel)

				if len(solver.currentLearnedClause) == 1 {
					solver.performLiteralAssertion(solver.currentLearnedClause[0], nil)
				} else {
					learnedClause :=solver.currentLearnedClause.Copy()
					solver.clauses = append(solver.clauses, learnedClause)
					solver.watchClause(learnedClause)
					solver.performLiteralAssertion(learnedClause[0], learnedClause)
				}
			}
		}
	}
	return fmt.Errorf("CDCL Solver supports only CNF formulas."), SAT_RESULT_UNDEFINED
}