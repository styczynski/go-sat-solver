package preprocessor

import (
	"fmt"

	"github.com/go-sat-solver/sat_solver"
	"github.com/go-sat-solver/sat_solver/preprocessor/cnf_tseytins"
	"github.com/go-sat-solver/sat_solver/test_utils"
)

func PreprocessAST(formula *sat_solver.Formula) (error, *sat_solver.SATFormula) {

	err, satFormula := cnf_tseytins.ConvertToCNFTseytins(formula) //cnf_naive.ConvertToCNFNaive(formula)
	if err != nil {
		return err, nil
	}

	test_utils.AssertSatResult(satFormula, false)

	fmt.Printf("Preprocessed formula:\n %s\n", satFormula.Brief())

	err, simplFormula := Optimize(satFormula)
	if err != nil {
		return err, nil
	}

	fmt.Printf("Simplified formula:\n %s\n", simplFormula.Brief())

	return nil, simplFormula
}