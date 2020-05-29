package preprocessor

import (
	"fmt"

	"github.com/go-sat-solver/sat_solver"
	"github.com/go-sat-solver/sat_solver/preprocessor/cnf_naive"
)

func PreprocessAST(formula sat_solver.Entry) (error, *sat_solver.SATFormula) {
	err, satFormula := cnf_naive.ConvertToCNFNaive(formula)
	if err != nil {
		return err, nil
	}

	fmt.Printf("Preprocessed formula:\n %s\n", satFormula.String())
	return nil, satFormula
}