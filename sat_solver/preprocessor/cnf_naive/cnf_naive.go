package cnf_naive

import (
	"fmt"

	"github.com/go-sat-solver/sat_solver"
)

func convertToCnf(expr *sat_solver.Formula, vars *sat_solver.SATVariableMapping) (error, *sat_solver.CNFFormula) {
	// For variable return formula unmodified
	if expr.Variable != nil {
		return nil, &sat_solver.CNFFormula{
			[][]int{ { vars.Get(expr.Variable.String), } },
		}
	} else if expr.And != nil {
		err, arg1 := convertToCnf(expr.And.Arg1, vars)
		if err != nil {
			return err, nil
		}
		err, arg2 := convertToCnf(expr.And.Arg2, vars)
		if err != nil {
			return err, nil
		}
		arg1.AndWith(arg2)
		return nil, arg1
	} else if expr.Or != nil {
		err, arg1 := convertToCnf(expr.Or.Arg1, vars)
		if err != nil {
			return err, nil
		}
		err, arg2 := convertToCnf(expr.Or.Arg2, vars)
		if err != nil {
			return err, nil
		}
		arg1.MulWith(arg2)
		return nil, arg1
	} else if expr.Not != nil {
		inner := expr.Not.Formula
		// Not with variable
		if inner.Variable != nil {
			return nil, &sat_solver.CNFFormula{
				[][]int{ { -vars.Get(inner.Variable.String), } },
			}
		} else if inner.Not != nil {
			// Double not
			return convertToCnf(inner.Not.Formula, vars)
		} else if inner.And != nil {
			return convertToCnf(sat_solver.MakeOr(
				sat_solver.MakeNot(inner.And.Arg1),
				sat_solver.MakeNot(inner.And.Arg2)), vars)
		} else if inner.Or != nil {
			return convertToCnf(sat_solver.MakeAnd(
				sat_solver.MakeNot(inner.Or.Arg1),
				sat_solver.MakeNot(inner.Or.Arg2)), vars)
		} else if inner.Implies != nil {
			return convertToCnf(sat_solver.MakeAnd(
				inner.Or.Arg1,
				sat_solver.MakeNot(inner.Implies.Arg2)), vars)
		} else if inner.Iff != nil {
			return convertToCnf(sat_solver.MakeOr(
				sat_solver.MakeAnd(inner.Iff.Arg1, sat_solver.MakeNot(inner.Iff.Arg2)),
				sat_solver.MakeAnd(inner.Iff.Arg2, sat_solver.MakeNot(inner.Iff.Arg1))), vars)
		} else if inner.Constant != nil {
			if inner.Constant.Bool == "F" {
				return nil, &sat_solver.CNFFormula{
					[][]int{ { -1, } },
				}
			} else {
				return nil, &sat_solver.CNFFormula{
					[][]int{ { 1, } },
				}
			}
		}
	} else if expr.Implies != nil {
		return convertToCnf(sat_solver.MakeOr(
			sat_solver.MakeNot(expr.Implies.Arg1),
			expr.Implies.Arg2), vars)
	} else if expr.Iff != nil {
		return convertToCnf(sat_solver.MakeOr(
			sat_solver.MakeAnd(expr.Iff.Arg1, expr.Iff.Arg2),
			sat_solver.MakeAnd(sat_solver.MakeNot(expr.Iff.Arg1), sat_solver.MakeNot(expr.Iff.Arg2))), vars)
	} else if expr.Constant != nil {
		if expr.Constant.Bool == "T" {
			return nil, &sat_solver.CNFFormula{
				[][]int{ { 1, } },
			}
		} else {
			return nil, &sat_solver.CNFFormula{
				[][]int{ { -1, } },
			}
		}
	}

	return fmt.Errorf("Invalid formula given to convertToCnf: %#v", expr), nil
}

func ConvertToCNFNaive(formula sat_solver.Entry) (error, *sat_solver.SATFormula) {
	vars := sat_solver.NewSATVariableMapping()
	err, cnfFormula := convertToCnf(formula.Formula, vars)
	if err != nil {
		return err, nil
	}
	return nil, sat_solver.NewSATFormula(cnfFormula, vars)
}