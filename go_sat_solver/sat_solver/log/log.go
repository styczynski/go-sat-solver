package log

import (
	"fmt"
	"io"
	"strings"
	"time"
)

type IDProvider interface {
	GetID() uint
}

type DescribableResult interface {
	Brief() string
	String() string
}

type EventCollector interface {
	StartProcessing(stageName string, IDProvider IDProvider, formatString string, formatArgs... interface{}) (error, uint)
	EndProcessing(id uint, result DescribableResult) error
	Trace(id uint, eventName string, formatString string, formatArgs... interface{}) error
	MustTrace(id uint, eventName string, formatString string, formatArgs... interface{})
}

type EventLogger struct {
	processName        map[uint]string
	currentProcesses   map[string]uint
	reverse            map[uint]string
	startTime          map[uint]int64
	providedIDs        map[uint]uint
	levels             map[uint]uint
	newID              uint
	output             io.Writer
}

func NewEventLogger(output io.Writer) *EventLogger {
	return &EventLogger{
		processName:      map[uint]string{},
		currentProcesses: map[string]uint{},
		reverse:          map[uint]string{},
		startTime:        map[uint]int64{},
		providedIDs:      map[uint]uint{},
		levels:           map[uint]uint{},
		newID:            1,
		output:           output,
	}
}

func (l *EventLogger) Trace(id uint, eventName string, formatString string, formatArgs... interface{}) error {
	prefix := ""
	nestLevel := int(l.levels[l.providedIDs[id]])
	if nestLevel > 0 {
		prefix = strings.Repeat("  ", nestLevel) + "| "
	}
	_, err := l.output.Write([]byte(fmt.Sprintf("[%d]%s [%s] %s: %s\n", l.providedIDs[id], prefix, eventName, l.processName[id], fmt.Sprintf(formatString, formatArgs...))))
	if err != nil {
		return err
	}
	return nil
}

func (l *EventLogger) MustTrace(id uint, eventName string, formatString string, formatArgs... interface{}) {
	err := l.Trace(id, eventName, formatString, formatArgs...)
	if err != nil {
		panic(err)
	}
}

func (l *EventLogger) StartProcessing(stageName string, IDProvider IDProvider, formatString string, formatArgs... interface{}) (error, uint) {
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

	_, err := l.output.Write([]byte(fmt.Sprintf("[%d]%s Starting: %s%s\n", processID, prefix, stageName, description)))
	if err != nil {
		return err, 0
	}

	return nil, newID
}

func (l *EventLogger) EndProcessing(id uint, result DescribableResult) error {
	key := l.reverse[id]
	l.levels[l.providedIDs[id]] = l.levels[l.providedIDs[id]]-1

	prefix := ""
	nestLevel := int(l.levels[l.providedIDs[id]])
	if nestLevel > 0 {
		prefix = strings.Repeat("  ", nestLevel) + "| "
	}
	_, err := l.output.Write([]byte(fmt.Sprintf("[%d]%s Finished: %s (took %d ms) %s\n", l.providedIDs[id], prefix, l.processName[id], (time.Now().UnixNano()-l.startTime[id])/1000000, result.Brief())))
	if err != nil {
		return err
	}

	delete(l.reverse, id)
	delete(l.currentProcesses, key)
	delete(l.processName, id)
	delete(l.providedIDs, id)
	delete(l.startTime, id)
	return nil
}
