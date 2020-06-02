package cnf_naive

import (
	"fmt"

	"github.com/go-sat-solver/sat_solver"
)

func ConvertToCnfAndChain(exprs []*sat_solver.Formula, vars *sat_solver.SATVariableMapping) (error, *sat_solver.CNFFormula) {
	err, curExpr := convertToCnf(exprs[0], vars)
	if err != nil {
		return err, nil
	}

	for _, expr := range exprs[1:] {
		err, nextExpr := convertToCnf(expr, vars)
		if err != nil {
			return err, nil
		}
		curExpr.AndWith(nextExpr)
	}
	return nil, curExpr
}

func convertToCnf(expr *sat_solver.Formula, vars *sat_solver.SATVariableMapping) (error, *sat_solver.CNFFormula) {
	// For variable return formula unmodified
	if expr.Variable != nil {
		return nil, &sat_solver.CNFFormula{
			[][]int64{ { vars.Get(expr.Variable.Name), } },
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

		len1 := len(arg1.Variables)
		len2 := len(arg2.Variables)

		if (len1 == 1 || len2 == 1) {
			arg1.MulWith(arg2)
			return nil, arg1
		} else {
			// Alternative method for avoiding exponential formula growth:
			//   Use fresh variable Z
			//   Return CNF((Z -> P) ^ (~Z -> Q))
			z, _ := vars.Fresh()
			return convertToCnf(sat_solver.MakeAnd(
				sat_solver.MakeImplies(sat_solver.MakeVar(z), expr.Or.Arg1),
				sat_solver.MakeImplies(sat_solver.MakeNot(sat_solver.MakeVar(z)), expr.Or.Arg2)), vars)
		}
	} else if expr.Not != nil {
		inner := expr.Not.Formula
		// Not with variable
		if inner.Variable != nil {
			return nil, &sat_solver.CNFFormula{
				[][]int64{ { -vars.Get(inner.Variable.Name), } },
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
				inner.Implies.Arg1,
				sat_solver.MakeNot(inner.Implies.Arg2)), vars)
		} else if inner.Iff != nil {
			return convertToCnf(sat_solver.MakeOr(
				sat_solver.MakeAnd(inner.Iff.Arg1, sat_solver.MakeNot(inner.Iff.Arg2)),
				sat_solver.MakeAnd(inner.Iff.Arg2, sat_solver.MakeNot(inner.Iff.Arg1))), vars)
		} else if inner.Constant != nil {
			if inner.Constant.Bool == "F" {
				return nil, &sat_solver.CNFFormula{
					[][]int64{ },
				}
			} else {
				return nil, &sat_solver.CNFFormula{
					[][]int64{ { } },
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
				[][]int64{ },
			}
		} else {
			return nil, &sat_solver.CNFFormula{
				[][]int64{ { } },
			}
		}
	}

	return fmt.Errorf("Invalid formula given to convertToCnf: %#v", expr), nil
}

func ConvertToCNFNaive(formula sat_solver.Entry, context *sat_solver.SATContext) (error, *sat_solver.SATFormula) {

	err, processID := context.StartProcessing("Convert to CNF using Tseytins transformation (naive)", "")
	if err != nil {
		return err, nil
	}

	vars := sat_solver.NewSATVariableMapping()
	err, cnfFormula := convertToCnf(formula.Formula, vars)
	if err != nil {
		return err, nil
	}
	newFormula := sat_solver.NewSATFormula(cnfFormula, vars, nil)

	err = context.EndProcessingFormula(processID, newFormula)
	if err != nil {
		return err, nil
	}

	return nil, newFormula
}