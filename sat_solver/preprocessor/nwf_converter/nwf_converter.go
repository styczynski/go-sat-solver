package nwf_converter

import (
	"fmt"

	"github.com/go-sat-solver/sat_solver"
)

func convert(expr *sat_solver.Formula, vars *sat_solver.SATVariableMapping) (error, *sat_solver.NWFFormula) {
	if expr.Variable != nil {
		v := vars.Get(expr.Variable.Name)
		return nil, &sat_solver.NWFFormula{
			And: nil,
			Or:  nil,
			Variable: &sat_solver.NWFVar{
				ID:    v,
			},
		}
	} else if expr.Not != nil {
		err, e := convert(expr.Not.Formula, vars)
		if err != nil {
			return err, nil
		}
		e.Negate()
		return nil, e
	} else if expr.Or != nil {
		err, e1 := convert(expr.Or.Arg1, vars)
		if err != nil {
			return err, nil
		}

		err, e2 := convert(expr.Or.Arg2, vars)
		if err != nil {
			return err, nil
		}

		return nil, (&sat_solver.NWFFormula{
			And: nil,
			Or: &sat_solver.NWFOr{
				Arg1:  e1,
				Arg2:  e2,
				IsNeg: false,
			},
			Variable: nil,
		}).UpdateTopNodeMetrics()
	} else if expr.And != nil {
		err, e1 := convert(expr.And.Arg1, vars)
		if err != nil {
			return err, nil
		}

		err, e2 := convert(expr.And.Arg2, vars)
		if err != nil {
			return err, nil
		}

		return nil, (&sat_solver.NWFFormula{
			And: &sat_solver.NWFAnd{
				Arg1:  e1,
				Arg2:  e2,
				IsNeg: false,
			},
		}).UpdateTopNodeMetrics()
	} else if expr.Constant != nil {
		if expr.Constant.Bool == "T" {
			return nil, &sat_solver.NWFFormula{
				Const: &sat_solver.NWFConst{
					Value: true,
				},
			}
		} else if expr.Constant.Bool == "F" {
			return nil, &sat_solver.NWFFormula{
				Const: &sat_solver.NWFConst{
					Value: false,
				},
			}
		}
	} else if expr.Implies != nil {
		err, e1 := convert(expr.Implies.Arg1, vars)
		if err != nil {
			return err, nil
		}

		err, e2 := convert(expr.Implies.Arg2, vars)
		if err != nil {
			return err, nil
		}
		e1.Negate()

		return nil, (&sat_solver.NWFFormula{
			And:      nil,
			Or:       &sat_solver.NWFOr{
				Arg1:  e1,
				Arg2:  e2,
				IsNeg: false,
			},
			Variable: nil,
		}).UpdateTopNodeMetrics()
	} else if expr.Iff != nil {
		err, e1 := convert(expr.Iff.Arg1, vars)
		if err != nil {
			return err, nil
		}

		err, e2 := convert(expr.Iff.Arg2, vars)
		if err != nil {
			return err, nil
		}
		ne1 := e1.Copy()
		ne1.Negate()
		ne2 := e2.Copy()
		ne2.Negate()

		return nil, (&sat_solver.NWFFormula{
			And: &sat_solver.NWFAnd{
				Arg1: (&sat_solver.NWFFormula{
					And:      nil,
					Or:       &sat_solver.NWFOr{
						Arg1:  ne1,
						Arg2:  e2,
						IsNeg: false,
					},
					Variable: nil,
				}).UpdateTopNodeMetrics(),
				Arg2: (&sat_solver.NWFFormula{
					And:      nil,
					Or:       &sat_solver.NWFOr{
						Arg1:  ne2,
						Arg2:  e1,
						IsNeg: false,
					},
					Variable: nil,
				}).UpdateTopNodeMetrics(),
				IsNeg: false,
			},
			Or:       nil,
			Variable: nil,
		}).UpdateTopNodeMetrics()
	}

	return fmt.Errorf("NWF Could not convert unknown boolean expression."), nil
}

func optimizeTree(formula *sat_solver.NWFFormula, changeDetected *bool) (error, *sat_solver.NWFFormula) {
	if formula.Or != nil {
		err, opt1 := optimizeTree(formula.Or.Arg1, changeDetected)
		if err != nil {
			return err, nil
		}
		err, opt2 := optimizeTree(formula.Or.Arg2, changeDetected)
		if err != nil {
			return err, nil
		}
		if opt1.Const != nil && opt2.Const != nil {
			*changeDetected = true
			return nil, &sat_solver.NWFFormula{
				Const: &sat_solver.NWFConst{
					Value: (opt1.Const.Value || opt2.Const.Value) != formula.Or.IsNeg,
				},
			}
		}
		if opt1.Const != nil {
			if opt1.Const.Value && !formula.Or.IsNeg {
				*changeDetected = true
				return nil, &sat_solver.NWFFormula{
					Const: &sat_solver.NWFConst{
						Value: true,
					},
				}
			} else if opt1.Const.Value && formula.Or.IsNeg {
				*changeDetected = true
				return nil, &sat_solver.NWFFormula{
					Const: &sat_solver.NWFConst{
						Value: false,
					},
				}
			} else if !opt1.Const.Value {
				*changeDetected = true
				if formula.Or.IsNeg {
					opt2.Negate()
				}
				return nil, opt2
			}
		}
		if opt2.Const != nil {
			if opt2.Const.Value && !formula.Or.IsNeg {
				*changeDetected = true
				return nil, &sat_solver.NWFFormula{
					Const: &sat_solver.NWFConst{
						Value: true,
					},
				}
			} else if opt2.Const.Value && formula.Or.IsNeg {
				*changeDetected = true
				return nil, &sat_solver.NWFFormula{
					Const: &sat_solver.NWFConst{
						Value: false,
					},
				}
			} else if !opt2.Const.Value {
				*changeDetected = true
				if formula.Or.IsNeg {
					opt1.Negate()
				}
				return nil, opt1
			}
		}

		_, opt1complex := opt1.NodeMetrics()
		_, opt2complex := opt2.NodeMetrics()
		if opt1complex + opt2complex <= 40 {
			// This is expensive
			opt1s := opt1.Serialize()
			opt2s := opt2.Serialize()
			//fmt.Printf("collapse %s and %s (%d vs %d)\n", opt1s, opt2s, opt1complex, opt2complex)
			if opt1s == opt2s {
				*changeDetected = true
				return nil, opt1.UpdateTopNodeMetrics()
			}
		}

		return nil, (&sat_solver.NWFFormula{
			Or: &sat_solver.NWFOr{
				Arg1:  opt1,
				Arg2:  opt2,
				IsNeg: formula.Or.IsNeg,
			},
		}).UpdateTopNodeMetrics()
	} else if formula.And != nil {
		err, opt1 := optimizeTree(formula.And.Arg1, changeDetected)
		if err != nil {
			return err, nil
		}
		err, opt2 := optimizeTree(formula.And.Arg2, changeDetected)
		if err != nil {
			return err, nil
		}
		if opt1.Const != nil && opt2.Const != nil {
			*changeDetected = true
			return nil, &sat_solver.NWFFormula{
				Const: &sat_solver.NWFConst{
					Value: (opt1.Const.Value && opt2.Const.Value) != formula.And.IsNeg,
				},
			}
		}
		if opt1.Const != nil {
			if !opt1.Const.Value && !formula.And.IsNeg {
				*changeDetected = true
				return nil, &sat_solver.NWFFormula{
					Const: &sat_solver.NWFConst{
						Value: false,
					},
				}
			} else if !opt1.Const.Value && formula.And.IsNeg {
				*changeDetected = true
				return nil, &sat_solver.NWFFormula{
					Const: &sat_solver.NWFConst{
						Value: true,
					},
				}
			} else if opt1.Const.Value {
				*changeDetected = true
				if formula.And.IsNeg {
					opt2.Negate()
				}
				return nil, opt2
			}
		}
		if opt2.Const != nil {
			if !opt2.Const.Value && !formula.And.IsNeg {
				*changeDetected = true
				return nil, &sat_solver.NWFFormula{
					Const: &sat_solver.NWFConst{
						Value: false,
					},
				}
			} else if opt2.Const.Value && formula.And.IsNeg {
				*changeDetected = true
				return nil, &sat_solver.NWFFormula{
					Const: &sat_solver.NWFConst{
						Value: true,
					},
				}
			} else if opt2.Const.Value {
				*changeDetected = true
				if formula.And.IsNeg {
					opt1.Negate()
				}
				return nil, opt1
			}
		}

		_, opt1complex := opt1.NodeMetrics()
		_, opt2complex := opt2.NodeMetrics()
		if opt1complex + opt2complex <= 40 {
			// This is expensive
			opt1s := opt1.Serialize()
			opt2s := opt2.Serialize()
			//fmt.Printf("collapse %s and %s (%d vs %d)\n", opt1s, opt2s, opt1complex, opt2complex)
			if opt1s == opt2s {
				*changeDetected = true
				return nil, opt1.UpdateTopNodeMetrics()
			}
		}

		return nil, (&sat_solver.NWFFormula{
			And: &sat_solver.NWFAnd{
				Arg1:  opt1,
				Arg2:  opt2,
				IsNeg: formula.And.IsNeg,
			},
		}).UpdateTopNodeMetrics()
	}
	return nil, formula
}

func ConvertToNWF(formula *sat_solver.Entry) (error, *sat_solver.SATFormula) {
	vars := sat_solver.NewSATVariableMapping()
	err, f := convert(formula.Formula, vars)
	if err != nil {
		return err, nil
	}
	optF := f
	for {
		changeDetected := false
		err, optF = optimizeTree(optF, &changeDetected)
		if err != nil {
			return err, nil
		}
		if !changeDetected {
			break
		}
	}

	return nil, sat_solver.NewSATFormula(optF, vars, nil)
}
