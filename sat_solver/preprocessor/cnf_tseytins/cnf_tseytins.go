package cnf_tseytins

import (
	"fmt"

	"github.com/go-sat-solver/sat_solver"
	"github.com/go-sat-solver/sat_solver/preprocessor/cnf_naive"
)

func convertToCnf(expr *sat_solver.Formula, vars *sat_solver.SATVariableMapping, ts *[]*sat_solver.Formula) (error, *sat_solver.Formula) {
	// For variable return formula unmodified
	if expr.Variable != nil {
		return nil, expr
	} else if expr.And != nil {
		err, leftVar := convertToCnf(expr.And.Arg1, vars, ts)
		if err != nil {
			return err, nil
		}
		err, rightVar := convertToCnf(expr.And.Arg2, vars, ts)
		if err != nil {
			return err, nil
		}
		name, _ := vars.Fresh()
		*ts = append(*ts, sat_solver.MakeIff(sat_solver.MakeVar(name), sat_solver.MakeAnd(leftVar, rightVar)))
		return nil, sat_solver.MakeVar(name)
	} else if expr.Or != nil {
		err, leftVar := convertToCnf(expr.Or.Arg1, vars, ts)
		if err != nil {
			return err, nil
		}
		err, rightVar := convertToCnf(expr.Or.Arg2, vars, ts)
		if err != nil {
			return err, nil
		}
		name, _ := vars.Fresh()
		*ts = append(*ts, sat_solver.MakeIff(sat_solver.MakeVar(name), sat_solver.MakeOr(leftVar, rightVar)))
		return nil, sat_solver.MakeVar(name)
	} else if expr.Not != nil {
		if expr.Not.Formula.Variable != nil {
			return nil, expr
		}
		err, argVar := convertToCnf(expr.Not.Formula, vars, ts)
		if err != nil {
			return err, nil
		}
		name, _ := vars.Fresh()
		*ts = append(*ts, sat_solver.MakeIff(sat_solver.MakeVar(name), sat_solver.MakeNot(argVar)))
		return nil, sat_solver.MakeVar(name)
	} else if expr.Implies != nil {
		err, leftVar := convertToCnf(expr.Implies.Arg1, vars, ts)
		if err != nil {
			return err, nil
		}
		err, rightVar := convertToCnf(expr.Implies.Arg2, vars, ts)
		if err != nil {
			return err, nil
		}
		name, _ := vars.Fresh()
		*ts = append(*ts, sat_solver.MakeIff(sat_solver.MakeVar(name), sat_solver.MakeImplies(leftVar, rightVar)))
		return nil, sat_solver.MakeVar(name)
	} else if expr.Iff != nil {
		err, leftVar := convertToCnf(expr.Iff.Arg1, vars, ts)
		if err != nil {
			return err, nil
		}
		err, rightVar := convertToCnf(expr.Iff.Arg2, vars, ts)
		if err != nil {
			return err, nil
		}
		name, _ := vars.Fresh()
		*ts = append(*ts, sat_solver.MakeIff(sat_solver.MakeVar(name), sat_solver.MakeIff(leftVar, rightVar)))
		return nil, sat_solver.MakeVar(name)
	} else if expr.Constant != nil {
		return nil, expr
	}

	return fmt.Errorf("Invalid formula given to convertToCnf: %#v", expr), nil
}

func ConvertToCNFTseytins(formula sat_solver.Entry) (error, *sat_solver.SATFormula) {
	vars := sat_solver.NewSATVariableMapping()
	ts := []*sat_solver.Formula{}
	err, _ := convertToCnf(formula.Formula, vars, &ts)
	if err != nil {
		return err, nil
	}

	// Add subsitution for the entire formula
	if len(ts) > 0 {
		ts = append(ts, ts[len(ts)-1].Iff.Arg1)
	} else {
		// Append entire formula
		ts = append(ts, formula.Formula)
	}

	fmt.Printf("Tseytins input formula:\n %s\n", formula.Formula.String())
	fmt.Printf("Tseytins output chain:\n %s\n", sat_solver.AndChainToString(ts))

	err, cnfFormula := cnf_naive.ConvertToCnfAndChain(ts, vars)
	if err != nil {
		return err, nil
	}
	return nil, sat_solver.NewSATFormula(cnfFormula, vars)
}