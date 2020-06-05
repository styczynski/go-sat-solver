package solver

import (
	"fmt"
	"io"

	"github.com/go-sat-solver/sat_solver"
)

type LoadedFormula interface {
	CanBeConvertedToFormula() bool
	CanBeConvertedToAST() bool
	ConvertToFormula() *sat_solver.SATFormula
	ConvertToAST()  *sat_solver.Entry
	IsCNF() bool
}

type Loader interface {
	Load(inputFormula io.Reader, context *sat_solver.SATContext) (error, LoadedFormula)
}

type LoaderFactory interface {
	CreateLoader(context *sat_solver.SATContext) Loader
	GetName() string
}

var DEFAULT_LOADER_NAME = "haskell"
var LOADER_FACTORIES = map[string]LoaderFactory{}

func RegisterLoaderFactory(factory LoaderFactory) {
	LOADER_FACTORIES[factory.GetName()] = factory
}

func LoadFormula(name string, inputFormula io.Reader, context *sat_solver.SATContext) (error, LoadedFormula) {
	if len(name) == 0 {
		if defaultFactory, ok := LOADER_FACTORIES[DEFAULT_LOADER_NAME]; ok {
			name = defaultFactory.GetName()
		} else {
			for factoryName, _ := range LOADER_FACTORIES {
				name = factoryName
				break
			}
		}
	}
	if loaderFactory, ok := LOADER_FACTORIES[name]; ok {
		loader := loaderFactory.CreateLoader(context)
		return loader.Load(inputFormula, context)
	} else {
		return fmt.Errorf("Loader with name '%s' not found.", name), nil
	}
}