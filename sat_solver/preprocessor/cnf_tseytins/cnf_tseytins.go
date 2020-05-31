package cnf_tseytins

import (
	"fmt"

	"github.com/go-sat-solver/sat_solver"
	"github.com/go-sat-solver/sat_solver/preprocessor/cnf_naive"
)

func convertToCnf(expr *sat_solver.Formula, vars *sat_solver.SATVariableMapping, ts *[]*sat_solver.Formula) (error, *sat_solver.Formula, bool) {
	// For variable return formula unmodified
	if expr.Variable != nil {
		return nil, expr, true
	} else if expr.And != nil {
		err, leftVar, _ := convertToCnf(expr.And.Arg1, vars, ts)
		if err != nil {
			return err, nil, false
		}
		err, rightVar, _ := convertToCnf(expr.And.Arg2, vars, ts)
		if err != nil {
			return err, nil, false
		}
		name, _ := vars.Fresh()

		a := sat_solver.MakeVar(name)
		b := leftVar
		c := rightVar

		*ts = append(*ts, sat_solver.MakeOr(sat_solver.MakeNot(a), b), sat_solver.MakeOr(sat_solver.MakeNot(a), c),
			sat_solver.MakeOr(sat_solver.MakeOr(sat_solver.MakeNot(b), sat_solver.MakeNot(c)), a))
		return nil, a, false
	} else if expr.Or != nil {
		err, leftVar, _ := convertToCnf(expr.Or.Arg1, vars, ts)
		if err != nil {
			return err, nil, false
		}
		err, rightVar, _ := convertToCnf(expr.Or.Arg2, vars, ts)
		if err != nil {
			return err, nil, false
		}
		name, _ := vars.Fresh()

		a := sat_solver.MakeVar(name)
		b := leftVar
		c := rightVar

		*ts = append(*ts, sat_solver.MakeOr(sat_solver.MakeOr(sat_solver.MakeNot(a), sat_solver.MakeNot(b)), c),
			sat_solver.MakeOr(sat_solver.MakeNot(b), a), sat_solver.MakeOr(sat_solver.MakeNot(c), a))
		return nil, a, false
	} else if expr.Not != nil {
		if expr.Not.Formula.Variable != nil {
			return nil, expr, false
		}
		err, argVar, _ := convertToCnf(expr.Not.Formula, vars, ts)
		if err != nil {
			return err, nil, false
		}
		name, _ := vars.Fresh()

		a := sat_solver.MakeVar(name)
		b := argVar

		*ts = append(*ts, sat_solver.MakeOr(sat_solver.MakeNot(a), sat_solver.MakeNot(b)), sat_solver.MakeOr(b, a))
		return nil, a, false
	} else if expr.Implies != nil {
		err, leftVar, _ := convertToCnf(expr.Implies.Arg1, vars, ts)
		if err != nil {
			return err, nil, false
		}
		err, rightVar, _ := convertToCnf(expr.Implies.Arg2, vars, ts)
		if err != nil {
			return err, nil, false
		}
		name, _ := vars.Fresh()

		a := sat_solver.MakeVar(name)
		b := leftVar
		c := rightVar

		// (~a | ~b | c) & (b | a) & (~c | a)
		*ts = append(*ts, sat_solver.MakeOr(sat_solver.MakeOr(sat_solver.MakeNot(a), sat_solver.MakeNot(b)), c),
			sat_solver.MakeOr(b, a),
			sat_solver.MakeOr(sat_solver.MakeNot(c), a),)
		return nil, a, false
	} else if expr.Iff != nil {
		err, leftVar, _ := convertToCnf(expr.Iff.Arg1, vars, ts)
		if err != nil {
			return err, nil, false
		}
		err, rightVar, _ := convertToCnf(expr.Iff.Arg2, vars, ts)
		if err != nil {
			return err, nil, false
		}
		name, _ := vars.Fresh()

		a := sat_solver.MakeVar(name)
		b := leftVar
		c := rightVar

		*ts = append(*ts,
			sat_solver.MakeOr(sat_solver.MakeOr(a, b), c),
			sat_solver.MakeOr(sat_solver.MakeOr(a, sat_solver.MakeNot(a)), c),
			sat_solver.MakeOr(sat_solver.MakeOr(b, sat_solver.MakeNot(b)), c),
			sat_solver.MakeOr(sat_solver.MakeOr(sat_solver.MakeNot(b), sat_solver.MakeNot(a)), c),
			sat_solver.MakeOr(sat_solver.MakeOr(sat_solver.MakeNot(c), sat_solver.MakeNot(a)), b),
			sat_solver.MakeOr(sat_solver.MakeOr(sat_solver.MakeNot(c), sat_solver.MakeNot(b)), a))
		return nil, a, false
	} else if expr.Constant != nil {
		return nil, expr, true
	}

	return fmt.Errorf("Invalid formula given to convertToCnf: %#v", expr), nil, false
}

func ConvertToCNFTseytins(formula *sat_solver.Formula) (error, *sat_solver.SATFormula) {
	vars := sat_solver.NewSATVariableMapping()
	ts := []*sat_solver.Formula{}
	err, f, _ := convertToCnf(formula, vars, &ts)
	if err != nil {
		return err, nil
	}

	// Add subsitution for the entire formula
	ts = append(ts, f)

	//fmt.Printf("Tseytins input formula:\n %s\n", formula.String())
	//fmt.Printf("Tseytins output chain:\n %s\n", sat_solver.AndChainToString(ts))

	err, cnfFormula := cnf_naive.ConvertToCnfAndChain(ts, vars)
	if err != nil {
		return err, nil
	}
	return nil, sat_solver.NewSATFormula(cnfFormula, vars, nil)
}