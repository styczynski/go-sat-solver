package cdcl_solver

import (
	"fmt"

	"github.com/styczynski/go-sat-solver/sat_solver"
	"github.com/styczynski/go-sat-solver/sat_solver/solver"
)

/*
 * CDCL solver factory
 */
type CDCLSolverFactory struct {}

func (cdclsf CDCLSolverFactory) CanSolveFormula(formula *sat_solver.SATFormula, context *sat_solver.SATContext) bool {
	_, ok := formula.Formula().(*sat_solver.CNFFormula)
	return ok
}

func (cdclsf CDCLSolverFactory) CreateSolver(formula *sat_solver.SATFormula, context *sat_solver.SATContext) solver.Solver {
	return NewCDCLSolver()
}

func (cdclsf CDCLSolverFactory) GetName() string {
	return "cdcl"
}

// Register solver factory
func init() {
	solver.RegisterSolverFactory(CDCLSolverFactory{})
}

/**
 * Complete solver state
 */
type CDCLSolver struct {
	// AVSIDS state
	AVSIDS
	// Decision trace
	CDCLSolverDecisionTrace
	// TWL state
	SolverTWLState
	// State information used for learning
	SolverLearnState
	// The process ID is used for SATContext and mostly debugging
	processID              uint
	// Enable debug output
	enableDebugLogging     bool
	// Stores result of solving process
	result                 SatResult
	// Context and formula that we work on
	// Please note that the formula may change after new clauses are learnt,
	// but variables mapping should be fine.
	context                *sat_solver.SATContext
	clauses                []sat_solver.CNFClause
	vars                   *sat_solver.SATVariableMapping
    // This index is used as qhead in Minisat.
    // It points to the next clause to check on a trace (used by solver.performUnitPropagation() fucntion)
	currentTraceCheckIndex int
}

/**
 * Create new CDCL solver instance
 */
func NewCDCLSolver() *CDCLSolver {
	return &CDCLSolver{
		result:               SatResultUndefined(),
		CDCLSolverDecisionTrace: CDCLSolverDecisionTrace{
			currentAssignment: map[sat_solver.CNFLiteral]Ternary{},
			varsInfo:          map[sat_solver.CNFLiteral]VariableAssignmentInformation{},
		},
		SolverTWLState: SolverTWLState{
			watchedLiterals: map[sat_solver.CNFLiteral][]*TWLRecord{},
		},
		SolverLearnState: SolverLearnState{
			visited:              map[sat_solver.CNFLiteral]bool{},
			currentLearnedClause: make([]sat_solver.CNFLiteral, 10),
		},
	}
}

/**
 * Solve sat formula
 */
func (solver *CDCLSolver) Solve(formula *sat_solver.SATFormula, context *sat_solver.SATContext) (error, solver.SolverResult) {
	solver.context = context
	solver.enableDebugLogging = context.IsSolverTracingEnabled()
	if f, ok := formula.Formula().(*sat_solver.CNFFormula); ok {
		/**
		 * Prepare solver state
		 */
		solver.vars = formula.Variables()
		for _, newClause := range f.Variables {
			if len(newClause) == 0 {
				solver.result = SatResultUnsat()
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
		solver.avsidsInit()

		if solver.enableDebugLogging {
			solver.context.Trace("start", "Started solver.")
		}

		for {
			// Unit propagation
			conflictingClause := solver.performUnitPropagation()
			if conflictingClause == nil {
				// Make a new decision
				lit, hasAnyLiterals := solver.findNextLiteralForDecision()

				if !hasAnyLiterals || lit == sat_solver.CNF_UNDEFINED {
					return nil, solver.foundResult(SatResultSat(solver))
				}

				solver.newDecision(lit)
			} else {
				// We have conflict
				if solver.enableDebugLogging {
					solver.context.Trace("conflict", "Conflicting clause detected on unit propagation. Decision trace: %s.", solver.getDecisionTraceString())
				}
				if solver.getDecisionLevel() == 0 {
					return nil, solver.foundResult(SatResultUnsat())
				}

				// Remeber a new clause
				newLevel := solver.learnClause(conflictingClause)
				solver.avsidsClauseLearnt(&conflictingClause)

				// Go backwards
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
	return fmt.Errorf("CDCL Solver supports only CNF formulas."), SatResultUndefined()
}