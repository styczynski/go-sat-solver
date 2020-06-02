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
}

type SATConfiguration struct {
	inputFile string
	expectedResult *bool
	enableSelfVerification bool
	enableEventCollector bool
}

func NewSATContextDebug(inputFile string) *SATContext {
	return NewSATContext(SATConfiguration{
		inputFile:              inputFile,
		expectedResult:         nil,
		enableSelfVerification: true,
		enableEventCollector:   true,
	})
}

func NewSATContextAssert(inputFile string, expectedResult bool) *SATContext {
	return NewSATContext(SATConfiguration{
			inputFile:              inputFile,
			expectedResult:         &expectedResult,
			enableSelfVerification: true,
			enableEventCollector:   true,
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

func (l *SATContext) StartProcessing(stageName string, formatString string, formatArgs... interface{}) (error, uint) {
	if l.configuration.enableEventCollector && l.eventCollector != nil {
		return l.eventCollector.StartProcessing(stageName, l, formatString, formatArgs...)
	}
	return nil, 0
}

func (l *SATContext) EndProcessing(id uint, result log.DescribableResult) error {
	if l.configuration.enableEventCollector && l.eventCollector != nil {
		err := l.eventCollector.EndProcessing(id, result)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *SATContext) EndProcessingFormula(id uint, result FormulaLikeResult) error {
	if l.configuration.enableEventCollector && l.eventCollector != nil {
		err := l.eventCollector.EndProcessing(id, result)
		if err != nil {
			return err
		}
		l.AssertSelfVerification(result.ToSATFormula())
	}
	return nil
}
