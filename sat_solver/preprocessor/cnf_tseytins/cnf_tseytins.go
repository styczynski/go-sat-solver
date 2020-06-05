package cnf_tseytins

import (
	"fmt"

	"github.com/go-sat-solver/sat_solver"
)

func notArr(vars []int64) []int64 {
	ret := make([]int64, len(vars))
	for i := 0; i < len(vars); i++ {
		ret[i] = -vars[i]
	}
	return ret
}

func copyAndAppend(i []int64, vals ...int64) []int64 {
	j := make([]int64, len(i), len(i)+len(vals))
	copy(j, i)
	return append(j, vals...)
}

func convertToCnf(expr *sat_solver.Formula, vars *sat_solver.SATVariableMapping, ts *[]sat_solver.CNFClause) (error, sat_solver.CNFLiteral, sat_solver.CNFLiteral) {
	// For variable return formula unmodified
	if expr.Variable != nil {
		v := vars.Get(expr.Variable.Name)
		return nil, v, v
	} else if expr.And != nil {
		err, leftVar, _ := convertToCnf(expr.And.Arg1, vars, ts)
		if err != nil {
			return err, 0, 0
		}
		err, rightVar, _ := convertToCnf(expr.And.Arg2, vars, ts)
		if err != nil {
			return err, 0, 0
		}
		_, newVar := vars.Fresh()

		a := newVar
		b := leftVar
		c := rightVar

		// (~a | b) & (~a | c) & (~b | ~c | a)
		*ts = append(*ts, sat_solver.CNFClause{-a, b}, sat_solver.CNFClause{-a, c}, sat_solver.CNFClause{a, -b, -c})
		//*ts = append(*ts, sat_solver.MakeOr(sat_solver.MakeNot(a), b), sat_solver.MakeOr(sat_solver.MakeNot(a), c),
		//	sat_solver.MakeOr(sat_solver.MakeOr(sat_solver.MakeNot(b), sat_solver.MakeNot(c)), a))
		return nil, a, 0
	} else if expr.Or != nil {
		err, leftVar, _ := convertToCnf(expr.Or.Arg1, vars, ts)
		if err != nil {
			return err, 0, 0
		}
		err, rightVar, _ := convertToCnf(expr.Or.Arg2, vars, ts)
		if err != nil {
			return err, 0, 0
		}
		_, newVar := vars.Fresh()

		a := newVar
		b := leftVar
		c := rightVar

		// (~a | b | c) & (~b | a) & (~c | a)
		*ts = append(*ts, sat_solver.CNFClause{-a, b, c}, sat_solver.CNFClause{ -b, a }, sat_solver.CNFClause{ -c, a })
		//*ts = append(*ts, sat_solver.MakeOr(sat_solver.MakeOr(sat_solver.MakeNot(a), sat_solver.MakeNot(b)), c),
		//	sat_solver.MakeOr(sat_solver.MakeNot(b), a), sat_solver.MakeOr(sat_solver.MakeNot(c), a))
		return nil, a, 0
	} else if expr.Not != nil {
		if expr.Not.Formula.Variable != nil {
			v := vars.Get(expr.Not.Formula.Variable.Name)
			return nil, -v, -v
		}
		err, argVar, _ := convertToCnf(expr.Not.Formula, vars, ts)
		if err != nil {
			return err, 0, 0
		}
		_, newVar := vars.Fresh()

		a := newVar
		b := argVar

		// (~a | ~b) & (b | a)
		*ts = append(*ts, sat_solver.CNFClause{-a, -b}, sat_solver.CNFClause{b, a})
		return nil, a, 0
	} else if expr.Implies != nil {
		err, leftVar, _ := convertToCnf(expr.Implies.Arg1, vars, ts)
		if err != nil {
			return err, 0, 0
		}
		err, rightVar, _ := convertToCnf(expr.Implies.Arg2, vars, ts)
		if err != nil {
			return err, 0, 0
		}
		_, newVar := vars.Fresh()

		a := newVar
		b := leftVar
		c := rightVar

		// (~a | ~b | c) & (b | a) & (~c | a)
		*ts = append(*ts, sat_solver.CNFClause{-a, -b, c}, sat_solver.CNFClause{b, a}, sat_solver.CNFClause{-c, a})
		//*ts = append(*ts, sat_solver.MakeOr(sat_solver.MakeOr(sat_solver.MakeNot(a), sat_solver.MakeNot(b)), c),
		//	sat_solver.MakeOr(b, a),
		//	sat_solver.MakeOr(sat_solver.MakeNot(c), a),)
		return nil, a, 0
	} else if expr.Iff != nil {
		err, leftVar, _ := convertToCnf(expr.Iff.Arg1, vars, ts)
		if err != nil {
			return err, 0, 0
		}
		err, rightVar, _ := convertToCnf(expr.Iff.Arg2, vars, ts)
		if err != nil {
			return err, 0, 0
		}
		_, newVar := vars.Fresh()

		a := newVar
		b := leftVar
		c := rightVar

		//(a | b | c) & (~b | ~a | c) & (~c | ~a | b) & (~c | ~b | a)
		*ts = append(*ts,
			sat_solver.CNFClause{a, b, c},
			sat_solver.CNFClause{-b, -a, c},
			sat_solver.CNFClause{-c, -a, b},
			sat_solver.CNFClause{-c, -b, a})

		//*ts = append(*ts,
		//	sat_solver.MakeOr(sat_solver.MakeOr(a, b), c),
		//	sat_solver.MakeOr(sat_solver.MakeOr(a, sat_solver.MakeNot(a)), c),
		//	sat_solver.MakeOr(sat_solver.MakeOr(b, sat_solver.MakeNot(b)), c),
		//	sat_solver.MakeOr(sat_solver.MakeOr(sat_solver.MakeNot(b), sat_solver.MakeNot(a)), c),
		//	sat_solver.MakeOr(sat_solver.MakeOr(sat_solver.MakeNot(c), sat_solver.MakeNot(a)), b),
		//	sat_solver.MakeOr(sat_solver.MakeOr(sat_solver.MakeNot(c), sat_solver.MakeNot(b)), a))
		return nil, a, 0
	} else if expr.Constant != nil {
		if expr.Constant.Bool == "T" {
			return nil, 1, 0
		} else {
			return nil, -1, 0
		}
	}

	return fmt.Errorf("Invalid formula given to convertToCnf: %#v", expr), 0, 0
}

func eliminateCNFTF(formula *sat_solver.SATFormula) (error, *sat_solver.SATFormula) {
	if v, ok := formula.Formula().(*sat_solver.CNFFormula); ok {
		newVars := make([]sat_solver.CNFClause, 0, len(v.Variables))
		for _, clause := range v.Variables {
			if len(clause) == 0 {
				return sat_solver.NewUnsatError(sat_solver.NewUnsatReasonCNFNormalization()), nil
			}
			newClause := make(sat_solver.CNFClause, 0, len(clause))
			occursVariable := true
			occursTrue := false
			for _, varID := range clause {
				if varID == 1 {
					occursTrue = true
				} else if varID == -1 {
					// Do nothing
				} else {
					occursVariable = true
					newClause = append(newClause, varID)
				}
			}

			if !occursTrue {
				if !occursVariable {
					return sat_solver.NewUnsatError(sat_solver.NewUnsatReasonCNFNormalization()), nil
				} else {
					newVars = append(newVars, newClause)
				}
			}
		}

		res := sat_solver.NewSATFormula(&sat_solver.CNFFormula{
			Variables: newVars,
		}, formula.Variables(), nil)

		return nil, res
	}
	return fmt.Errorf("Expected CNF formula."), nil
}

func ConvertToCNFTseytins(formula *sat_solver.Formula, context *sat_solver.SATContext) (error, *sat_solver.SATFormula) {
	err, newContext := context.StartProcessing("Convert to CNF using Tseytins transformation", "")
	if err != nil {
		return err, nil
	}

	vars := sat_solver.NewSATVariableMapping()
	ts := []sat_solver.CNFClause{}
	err, f, topLevelVar := convertToCnf(formula, vars, &ts)
	if err != nil {
		return err, nil
	}

	// Add substitution for the entire formula
	if topLevelVar != 0 {
		ts = append(ts, sat_solver.CNFClause{ topLevelVar })
	} else {
		ts = append(ts, sat_solver.CNFClause{ f })
	}

	tseytinsCnf := sat_solver.NewSATFormula(&sat_solver.CNFFormula{
		Variables: ts,
	}, vars, nil)
	err, tseytinsCnf = eliminateCNFTF(tseytinsCnf)
	if err != nil {
		return err, nil
	}

	err = newContext.EndProcessingFormula(tseytinsCnf)
	if err != nil {
		return err, nil
	}

	return nil, tseytinsCnf
}