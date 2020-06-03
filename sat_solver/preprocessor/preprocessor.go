package preprocessor

import (
	"github.com/go-sat-solver/sat_solver"
	"github.com/go-sat-solver/sat_solver/preprocessor/cnf_tseytins"
)

func PreprocessAST(formula *sat_solver.Formula, context *sat_solver.SATContext) (error, *sat_solver.SATFormula) {

	err, newContext := context.StartProcessing("Convert formula","")
	if err != nil {
		return err, nil
	}
	err, satFormula := cnf_tseytins.ConvertToCNFTseytins(formula, context)
	if err != nil {
		return err, nil
	}
	if satFormula.IsQuickUNSAT() {
		return nil, satFormula
	}
	err = newContext.EndProcessingFormula(satFormula)
	if err != nil {
		return err, nil
	}

	err, newContext = context.StartProcessing("Preprocess formula","")
	if err != nil {
		return err, nil
	}
	err, simplFormula := Optimize(satFormula, context)
	if err != nil {
		return err, nil
	}
	err = newContext.EndProcessingFormula(simplFormula)
	if err != nil {
		return err, nil
	}

	return nil, simplFormula
}