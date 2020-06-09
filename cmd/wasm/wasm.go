package main

import (
	"fmt"
	"strings"
	"syscall/js"
	"time"

	"github.com/styczynski/go-sat-solver/sat_solver"
	"github.com/styczynski/go-sat-solver/sat_solver/core"
	"github.com/styczynski/go-sat-solver/sat_solver/log"
)


type JSLogger struct {
	processName        map[uint]string
	currentProcesses   map[string]uint
	reverse            map[uint]string
	startTime          map[uint]int64
	providedIDs        map[uint]uint
	levels             map[uint]uint
	newID              uint

	handler js.Value
}

func NewJSLogger(handler js.Value) *JSLogger {
	return &JSLogger{
		processName:      map[uint]string{},
		currentProcesses: map[string]uint{},
		reverse:          map[uint]string{},
		startTime:        map[uint]int64{},
		providedIDs:      map[uint]uint{},
		levels:           map[uint]uint{},
		newID:            1,
		handler:         handler,
	}
}

func (l *JSLogger) Trace(id uint, eventName string, formatString string, formatArgs... interface{}) error {
	prefix := ""
	nestLevel := int(l.levels[l.providedIDs[id]])
	if nestLevel > 0 {
		prefix = strings.Repeat("  ", nestLevel) + "| "
	}

	l.handler.Call("trace", js.ValueOf(fmt.Sprintf("[%d]%s [%s] %s: %s\n", l.providedIDs[id], prefix, eventName, l.processName[id], fmt.Sprintf(formatString, formatArgs...))))
	return nil
}

func (l *JSLogger) MustTrace(id uint, eventName string, formatString string, formatArgs... interface{}) {
	err := l.Trace(id, eventName, formatString, formatArgs...)
	if err != nil {
		panic(err)
	}
}

func (l *JSLogger) StartProcessing(stageName string, IDProvider log.IDProvider, formatString string, formatArgs... interface{}) (error, uint) {
	f := make([]int, 10)
	for i := 0; i < 20000000; i++ {
		f[i%10] = i
	}

	processID := IDProvider.GetID()
	if _, ok := l.levels[processID]; !ok {
		l.levels[processID] = 1
	} else {
		l.levels[processID] = l.levels[processID]+1
	}

	key := fmt.Sprintf("<%s>_%d", stageName, processID)
	newID := l.newID
	l.currentProcesses[key] = newID
	l.reverse[newID] = key
	l.processName[newID] = stageName
	l.providedIDs[newID] = processID
	l.startTime[newID] = time.Now().UnixNano()
	l.newID++

	description := ""
	descriptionContent := fmt.Sprintf(formatString, formatArgs...)
	if len(descriptionContent) > 0 {
		description = fmt.Sprintf(" (%s)", descriptionContent)
	}

	prefix := ""
	nestLevel := int(l.levels[processID])-1
	if nestLevel > 0 {
		prefix = strings.Repeat("  ", nestLevel) + "| "
	}

	l.handler.Call("startProcessing", js.ValueOf(fmt.Sprintf("[%d]%s Starting: %s%s\n", processID, prefix, stageName, description)))

	r := make(chan bool)
	js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		r <- true
		return nil
	}), 0)
	<-r
	close(r)

	return nil, newID
}

func (l *JSLogger) EndProcessing(id uint, result log.DescribableResult) error {
	key := l.reverse[id]
	l.levels[l.providedIDs[id]] = l.levels[l.providedIDs[id]]-1

	prefix := ""
	nestLevel := int(l.levels[l.providedIDs[id]])
	if nestLevel > 0 {
		prefix = strings.Repeat("  ", nestLevel) + "| "
	}
	l.handler.Call("endProcessing", js.ValueOf(fmt.Sprintf("[%d]%s Finished: %s (took %d ms) %s\n", l.providedIDs[id], prefix, l.processName[id], (time.Now().UnixNano()-l.startTime[id])/1000000, result.Brief())))

	delete(l.reverse, id)
	delete(l.currentProcesses, key)
	delete(l.processName, id)
	delete(l.providedIDs, id)
	delete(l.startTime, id)
	return nil
}


func solve(this js.Value, p []js.Value) interface{} {
	go func() {
		conf := sat_solver.DefaultSATConfiguration()
		conf.EnableEventCollector = true

		conf.LoaderName = p[3].String()
		conf.SolverName = p[4].String()

		err, result := core.RunSATSolverOnString(p[0].String(), sat_solver.NewSATContextWithEventCollector(conf, NewJSLogger(p[1])))
		if err != nil {
			p[2].Call("call", p[2], js.ValueOf(err.Error()), js.Undefined(), js.Undefined())
			return
		}
		assgn := map[string]interface{}{}
		assgnRaw := result.GetSatisfyingAssignment()
		for k, v := range assgnRaw {
			assgn[k] = js.ValueOf(v)
		}
		p[2].Call("call", p[2], js.Undefined(), js.ValueOf(result.ToInt()), js.ValueOf(assgn))
	}()
	return js.Null()
}

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("solve", js.FuncOf(solve))
	<-c
}