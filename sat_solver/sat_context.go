package sat_solver

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-sat-solver/sat_solver/log"
)

type SATContext struct {
	context context.Context
	configuration *SATConfiguration
	eventCollector log.EventCollector
	contextID uint
	processID uint
}

type SATConfiguration struct {
	InputFile              string
	ExpectedResult         *bool
	EnableSelfVerification bool
	EnableEventCollector   bool
	EnableSolverTracing    bool
	EnableCNFConversion    bool
	EnableASTOptimization  bool
	EnableCNFOptimizations bool
	SolverName             string
}

func DefaultSATConfiguration() SATConfiguration {
	return SATConfiguration{
		InputFile: "",
		EnableSelfVerification: false,
		EnableEventCollector: false,
		EnableSolverTracing: false,
		EnableCNFConversion: true,
		EnableASTOptimization: true,
		EnableCNFOptimizations: true,
		SolverName: "",
	}
}

func DefaultSATContext() *SATContext {
	return NewSATContext(DefaultSATConfiguration())
}

func NewSATContext(conf SATConfiguration) *SATContext {
	return &SATContext{
		context: context.Background(),
		contextID: 0,
		configuration: &conf,
		eventCollector: log.NewEventLogger(os.Stdout),
	}
}

func boolToStr(v bool) string {
	if v {
		return "[X]"
	}
	return "[ ]"
}

func (context *SATContext) DescribeConfiguration() string {
	conf := *context.configuration

	expectedResultStr := "N/A"
	if conf.ExpectedResult != nil {
		expectedResultStr = fmt.Sprintf("%t", *conf.ExpectedResult)
	}

	return strings.Join([]string{
		fmt.Sprintf("\tInput file                => %s", conf.InputFile),
		fmt.Sprintf("\tExpected result           => %s", expectedResultStr),
		fmt.Sprintf("\tUsed solver               => '%s'", conf.SolverName),
		fmt.Sprintf("\tEnable event collector?   => %s", boolToStr(conf.EnableEventCollector)),
		fmt.Sprintf("\tEnable self verification? => %s", boolToStr(conf.EnableSelfVerification)),
		fmt.Sprintf("\tEnable solver tracing?    => %s", boolToStr(conf.EnableSolverTracing)),
		fmt.Sprintf("\tEnable CNF conversion?    => %s", boolToStr(conf.EnableCNFConversion)),
		fmt.Sprintf("\tEnable CNF optimizations? => %s", boolToStr(conf.EnableCNFConversion && conf.EnableCNFOptimizations)),
		fmt.Sprintf("\tEnable AST optimization?  => %s", boolToStr(conf.EnableASTOptimization)),
	}, "\n")
}

func (context *SATContext) GetConfiguration() *SATConfiguration {
	return context.configuration
}

func (context *SATContext) IsSolverTracingEnabled() bool {
	return context.configuration.EnableSolverTracing
}

func (context *SATContext) IsSelfVerificationEnabled() bool {
	return context.configuration.EnableSelfVerification
}

func (context *SATContext) AssertSelfVerification(formula *SATFormula) {
	if context.configuration.EnableSelfVerification {
		if context.configuration.ExpectedResult != nil {
			AssertSatResult(formula, *context.configuration.ExpectedResult)
		}
	}
}

func (l *SATContext) GetID() uint {
	return l.contextID
}

type FormulaLikeResult interface {
	log.DescribableResult
	ToSATFormula() *SATFormula
}

func (l *SATContext) StartProcessing(stageName string, formatString string, formatArgs... interface{}) (error, *SATContext) {
	if l.configuration.EnableEventCollector && l.eventCollector != nil {
		err, processID := l.eventCollector.StartProcessing(stageName, l, formatString, formatArgs...)
		if err != nil {
			return err, nil
		}
		return nil, &SATContext{
			context:        l.context,
			configuration:  l.configuration,
			eventCollector: l.eventCollector,
			contextID:      l.contextID,
			processID:      processID,
		}
	}
	return nil, l
}

func (l *SATContext) EndProcessing(result log.DescribableResult) error {
	if l.configuration.EnableEventCollector && l.eventCollector != nil {
		err := l.eventCollector.EndProcessing(l.processID, result)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *SATContext) Trace(eventName string, formatString string, formatArgs... interface{}) {
	if l.configuration.EnableEventCollector && l.eventCollector != nil {
		l.eventCollector.MustTrace(l.processID, eventName, formatString, formatArgs...)
	}
}

func (l *SATContext) EndProcessingFormula(result FormulaLikeResult) error {
	if l.configuration.EnableEventCollector && l.eventCollector != nil {
		err := l.eventCollector.EndProcessing(l.processID, result)
		if err != nil {
			return err
		}
		l.AssertSelfVerification(result.ToSATFormula())
	}
	return nil
}
