package solver

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-sat-solver/sat_solver"
)

/**
 * Solver instance
 */
type Solver interface {
	Solve(formula *sat_solver.SATFormula, context *sat_solver.SATContext) (error, SolverResult)
}

/**
 * Solver result representation
 */
type SolverResult interface {
	ToBool() bool
	ToInt() int
	String() string
	Brief() string
	GetSatisfyingAssignment() map[string]bool
	IsSAT() bool
	IsUNSAT() bool
	IsUndefined() bool
}

func GetSolverResultSatisfyingAssignmentString(result SolverResult) string {
	if result.IsSAT() {
		assgn := result.GetSatisfyingAssignment()
		rows := make([]string, len(assgn))
		i := 0
		for k, v := range assgn {
			rows[i] = fmt.Sprintf("\t| %s  =>  %t", k, v)
			i++
		}
		sort.Strings(rows)
		return fmt.Sprintf("SATAssignment:\n%s", strings.Join(rows, "\n"))
	} else if result.IsUndefined() {
		return "SATAssignment: N/A"
	} else if result.IsUNSAT() {
		return "SATAssignment: N/A"
	}
	return "SatAssignment: N/A"
}

type EmptySolverResult struct {}

func (EmptySolverResult) ToBool() bool {
	return false
}

func (EmptySolverResult) ToInt() int {
	return 0
}

func (EmptySolverResult) String() string {
	return "SAT_ERROR"
}

func (EmptySolverResult) Brief() string {
	return "SAT_ERROR"
}

func (EmptySolverResult) GetSatisfyingAssignment() map[string]bool {
	return map[string]bool{}
}

func (EmptySolverResult) IsSAT() bool {
	return false
}

func (EmptySolverResult) IsUNSAT() bool {
	return false
}

func (EmptySolverResult) IsUndefined() bool {
	return true
}

type SolverQuickUnsatResult struct {}

func (SolverQuickUnsatResult) ToBool() bool {
	return false
}

func (SolverQuickUnsatResult) ToInt() int {
	return 0
}

func (SolverQuickUnsatResult) String() string {
	return "UNSAT"
}

func (SolverQuickUnsatResult) Brief() string {
	return "UNSAT"
}

func (SolverQuickUnsatResult) GetSatisfyingAssignment() map[string]bool {
	return map[string]bool{}
}

func (SolverQuickUnsatResult) IsSAT() bool {
	return false
}

func (SolverQuickUnsatResult) IsUNSAT() bool {
	return true
}

func (SolverQuickUnsatResult) IsUndefined() bool {
	return false
}

func Solve(formula *sat_solver.SATFormula, solverName string, context *sat_solver.SATContext) (error, SolverResult) {
	err, solvingContext := context.StartProcessing("Solve formula (CDCL solver)", "")
	if err != nil {
		return err, EmptySolverResult{}
	}
	err, solver := CreateSolver(solverName, formula, context)
	if err != nil {
		return err, EmptySolverResult{}
	}
	err, result := solver.Solve(formula, context)
	if err != nil {
		return err, EmptySolverResult{}
	}
	err = solvingContext.EndProcessing(result)
	if err != nil {
		return err, EmptySolverResult{}
	}

	return nil, result
}