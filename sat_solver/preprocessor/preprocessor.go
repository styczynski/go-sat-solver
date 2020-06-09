package preprocessor

import (
	"github.com/styczynski/go-sat-solver/sat_solver"
	solver "github.com/styczynski/go-sat-solver/sat_solver/loaders"
	"github.com/styczynski/go-sat-solver/sat_solver/preprocessor/cnf_tseytins"
)

func PreprocessAST(formula solver.LoadedFormula, context *sat_solver.SATContext) (error, *sat_solver.SATFormula) {

	var satFormula *sat_solver.SATFormula

	skipOpt := false
	if formula.CanBeConvertedToFormula() {
		satFormula = formula.ConvertToFormula()
		skipOpt = true
		if !satFormula.IsCNF() && satFormula.CanBeConvertedToAST() {
			skipOpt = false
		}
	}

	if skipOpt {
		// Do nothing
	} else {
		err, newContext := context.StartProcessing("Convert formula", "")
		if err != nil {
			return err, nil
		}
		err, satFormula = cnf_tseytins.ConvertToCNFTseytins(formula.ConvertToAST().Formula, context)
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
	}

	if context.GetConfiguration().EnableCNFOptimizations {
		err, newContext := context.StartProcessing("Preprocess formula", "")
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
	} else {
		return nil, satFormula
	}
}