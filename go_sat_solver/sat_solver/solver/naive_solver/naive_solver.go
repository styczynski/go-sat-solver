package naive_solver

import (
	"fmt"
	"math"

	"github.com/styczynski/go-sat-solver/sat_solver"
	solv "github.com/styczynski/go-sat-solver/sat_solver/solver"
)

/**
 * Trivial solver that goes from 0 to 2^NoOfVariables(formula) and checks if it found a solution.
 * Please note that this solver will raise an error if NoOfVariables(formula) >= 64, because the permutation
 * representation does not fit into int64 type and it would really take a while so why do that do yourself?
 * The solver runs in O(2^n) time.
 * This code is mainly used for educational/testing purpose.
 * I really does not recommend using NaiveSolver if you have CDCL right next to it (XD).
 */
type NaiveSolver struct {}

func NewNaiveSolver() *NaiveSolver {
	return &NaiveSolver{}
}

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
	return false
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

/*
 * Naive solver factory
 */
type NaiveSolverFactory struct {}

func (nsf NaiveSolverFactory) CanSolveFormula(formula *sat_solver.SATFormula, context *sat_solver.SATContext) bool {
	return true
}

func (nsf NaiveSolverFactory) CreateSolver(formula *sat_solver.SATFormula, context *sat_solver.SATContext) solv.Solver {
	return NewNaiveSolver()
}

func (nsf NaiveSolverFactory) GetName() string {
	return "naive"
}

// Register sovler factory
func init() {
	solv.RegisterSolverFactory(NaiveSolverFactory{})
}

/**
 * Solve input formula.
 */
func (solver *NaiveSolver) Solve(formula *sat_solver.SATFormula, context *sat_solver.SATContext) (error, solv.SolverResult) {
	fmt.Printf("Naive solver input:\n %s\n", formula.String())

	err, vars := formula.Normalize()
	if err != nil {
		return err, solv.EmptySolverResult{}
	}

	varCount := int64(len(vars))
	if varCount > 63 {
		return fmt.Errorf("Too many variables for naive solver (%d)", varCount), solv.EmptySolverResult{}
	}

	values := int64(0)
	/*
	 * Goes through all of the combinations and increment int64 one by one.
	 */
	iterCount := int64(math.Exp2(float64(varCount)))
	for i := int64(0); i < iterCount; i++ {
		for j := int64(0); j < varCount; j++ {
			vars[j] = (int64(1) >> uint64(j)) & values != 0
		}
		//
		// Evaluate formula if it's true then we print the result
		//
		if formula.Evaluate(vars) {
			formulaVars := formula.Variables()
			result := map[string]bool{}
			for k, v := range vars {
				// If the variable was introduced later during optimizations we discard it
				if formulaVars.IsFounderVariable(sat_solver.CNFLiteral(k+1)) {
					result[formulaVars.Reverse(sat_solver.CNFLiteral(k+1))] = v
				}
			}
			return nil, SatResult{
				resultType: SAT_RESULT_SAT,
				assgn:      result,
			}
		}
		values++
	}

	// We do not found any satisfying assignment
	return nil, SatResult{
		resultType: SAT_RESULT_UNSAT,
		assgn:      map[string]bool{},
	}
}