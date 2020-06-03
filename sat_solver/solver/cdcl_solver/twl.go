package cdcl_solver

import (
	"fmt"

	"github.com/go-sat-solver/sat_solver"
)

/*
 * Part of the watched literal algorithm.
 * TWLRecord contains a single literal and its context i.e the specified clause in which this literal occurs.
 */
type TWLRecord struct {
	Clause  sat_solver.CNFClause
	Literal sat_solver.CNFLiteral
}

func NewTWLRecord(literal sat_solver.CNFLiteral, clause  sat_solver.CNFClause) *TWLRecord {
	return &TWLRecord{
		Clause:  clause,
		Literal: literal,
	}
}

func (twl *TWLRecord) String(s *CDCLSolver) string {
	return fmt.Sprintf("TWL[%s in %s]", twl.Literal.String(s.vars), twl.Clause.String(s.vars))
}