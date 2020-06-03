package sat_solver

import (
	"context"
	"os"

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
	inputFile string
	expectedResult *bool
	enableSelfVerification bool
	enableEventCollector bool
	enableSolverTracing bool
}

func NewSATContextDebug(inputFile string) *SATContext {
	return NewSATContext(SATConfiguration{
		inputFile:              inputFile,
		expectedResult:         nil,
		enableSelfVerification: true,
		enableEventCollector:   true,
		enableSolverTracing:    true,
	})
}

func NewSATContextAssert(inputFile string, expectedResult bool) *SATContext {
	return NewSATContext(SATConfiguration{
		inputFile:              inputFile,
		expectedResult:         &expectedResult,
		enableSelfVerification: true,
		enableEventCollector:   true,
		enableSolverTracing:    true,
	})
}

func NewSATContext(conf SATConfiguration) *SATContext {
	return &SATContext{
		context: context.Background(),
		contextID: 0,
		configuration: &conf,
		eventCollector: log.NewEventLogger(os.Stdout),
	}
}

func (context *SATContext) IsSolverTracingEnabled() bool {
	return context.configuration.enableSolverTracing
}

func (context *SATContext) IsSelfVerificationEnabled() bool {
	return context.configuration.enableSelfVerification
}

func (context *SATContext) AssertSelfVerification(formula *SATFormula) {
	if context.configuration.enableSelfVerification {
		if context.configuration.expectedResult != nil {
			AssertSatResult(formula, *context.configuration.expectedResult)
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
	if l.configuration.enableEventCollector && l.eventCollector != nil {
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
	if l.configuration.enableEventCollector && l.eventCollector != nil {
		err := l.eventCollector.EndProcessing(l.processID, result)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *SATContext) Trace(eventName string, formatString string, formatArgs... interface{}) {
	if l.configuration.enableEventCollector && l.eventCollector != nil {
		l.eventCollector.MustTrace(l.processID, eventName, formatString, formatArgs...)
	}
}

func (l *SATContext) EndProcessingFormula(result FormulaLikeResult) error {
	if l.configuration.enableEventCollector && l.eventCollector != nil {
		err := l.eventCollector.EndProcessing(l.processID, result)
		if err != nil {
			return err
		}
		l.AssertSelfVerification(result.ToSATFormula())
	}
	return nil
}
