package preprocessor

import (
	"fmt"

	"github.com/go-sat-solver/sat_solver"
	"github.com/go-sat-solver/sat_solver/preprocessor/cnf_tseytins"
)

func PreprocessAST(formula sat_solver.Entry) (error, *sat_solver.SATFormula) {
	err, satFormula := cnf_tseytins.ConvertToCNFTseytins(formula) //cnf_naive.ConvertToCNFNaive(formula)
	if err != nil {
		return err, nil
	}

	fmt.Printf("Preprocessed formula:\n %s\n", satFormula.String())

	err, simplFormula := VariableElimination(satFormula)
	if err != nil {
		return err, nil
	}

	fmt.Printf("Simplified formula:\n %s\n", simplFormula.String())

	err, simplFormula2 := VariableElimination(simplFormula)
	if err != nil {
		return err, nil
	}

	fmt.Printf("Simplified formula2:\n %s\n", simplFormula2.String())


	return nil, satFormula
}