package cnf_tseytins

import (
	"fmt"

	"github.com/styczynski/go-sat-solver/sat_solver"
)

func convertToCnfNaive(expr *sat_solver.Formula, vars *sat_solver.SATVariableMapping, ts *[]*sat_solver.Formula) (error, *sat_solver.Formula, bool) {
	// For variable return formula unmodified
	if expr.Variable != nil {
		return nil, expr, true
	} else if expr.And != nil {
		err, leftVar, _ := convertToCnfNaive(expr.And.Arg1, vars, ts)
		if err != nil {
			return err, nil, false
		}
		err, rightVar, _ := convertToCnfNaive(expr.And.Arg2, vars, ts)
		if err != nil {
			return err, nil, false
		}
		name, _ := vars.Fresh()
		*ts = append(*ts, sat_solver.MakeIff(sat_solver.MakeVar(name), sat_solver.MakeAnd(leftVar, rightVar)))
		return nil, sat_solver.MakeVar(name), false
	} else if expr.Or != nil {
		err, leftVar, _ := convertToCnfNaive(expr.Or.Arg1, vars, ts)
		if err != nil {
			return err, nil, false
		}
		err, rightVar, _ := convertToCnfNaive(expr.Or.Arg2, vars, ts)
		if err != nil {
			return err, nil, false
		}
		name, _ := vars.Fresh()
		*ts = append(*ts, sat_solver.MakeIff(sat_solver.MakeVar(name), sat_solver.MakeOr(leftVar, rightVar)))
		return nil, sat_solver.MakeVar(name), false
	} else if expr.Not != nil {
		if expr.Not.Formula.Variable != nil {
			return nil, expr, true
		}
		err, argVar, _ := convertToCnfNaive(expr.Not.Formula, vars, ts)
		if err != nil {
			return err, nil, false
		}
		name, _ := vars.Fresh()
		*ts = append(*ts, sat_solver.MakeIff(sat_solver.MakeVar(name), sat_solver.MakeNot(argVar)))
		return nil, sat_solver.MakeVar(name), false
	} else if expr.Implies != nil {
		err, leftVar, _ := convertToCnfNaive(expr.Implies.Arg1, vars, ts)
		if err != nil {
			return err, nil, false
		}
		err, rightVar, _ := convertToCnfNaive(expr.Implies.Arg2, vars, ts)
		if err != nil {
			return err, nil, false
		}
		name, _ := vars.Fresh()
		*ts = append(*ts, sat_solver.MakeIff(sat_solver.MakeVar(name), sat_solver.MakeImplies(leftVar, rightVar)))
		return nil, sat_solver.MakeVar(name), false
	} else if expr.Iff != nil {
		err, leftVar, _ := convertToCnfNaive(expr.Iff.Arg1, vars, ts)
		if err != nil {
			return err, nil, false
		}
		err, rightVar, _ := convertToCnfNaive(expr.Iff.Arg2, vars, ts)
		if err != nil {
			return err, nil, false
		}
		name, _ := vars.Fresh()
		*ts = append(*ts, sat_solver.MakeIff(sat_solver.MakeVar(name), sat_solver.MakeIff(leftVar, rightVar)))
		return nil, sat_solver.MakeVar(name), false
	} else if expr.Constant != nil {
		return nil, expr, true
	}

	return fmt.Errorf("Invalid formula given to convertToCnf: %#v", expr), nil, false
}