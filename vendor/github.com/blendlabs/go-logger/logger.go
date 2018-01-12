package logger

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
)

const (

	// DefaultListenerName is a default.
	DefaultListenerName = "default"
)

var (
	// DefaultListenerWorkers is the default number of workers per listener.
	DefaultListenerWorkers = runtime.NumCPU()
)

// New returns a new agent with a given set of enabled flags.
func New(flags ...Flag) *Logger {
	return &Logger{
		recoverPanics: true,
		flags:         NewFlagSet(flags...),
	}
}

// NewFromEnv returns a new agent with settings read from the environment.
func NewFromEnv() *Logger {
	return New().WithFlagsFromEnv().WithWriter(NewWriterFromEnv())
}

// All returns a valid agent that fires any and all events.
func All() *Logger {
	return New().WithFlags(NewFlagSetAll()).WithWriter(NewWriterFromEnv())
}

// None returns a valid agent that won't fire any events.
func None() *Logger {
	return &Logger{
		flags: NewFlagSetNone(),
	}
}

// Logger is a handler for various logging events with descendent handlers.
type Logger struct {
	writer Writer

	flagsLock sync.Mutex
	flags     *FlagSet

	workersLock sync.Mutex
	workers     map[Flag]map[string]*Worker

	recoverPanics bool

	writeWorkerLock      sync.Mutex
	writeWorker          *Worker
	writeErrorWorkerLock sync.Mutex
	writeErrorWorker     *Worker
}

// Writer returns the inner Logger for the diagnostics agent.
func (l *Logger) Writer() Writer {
	return l.writer
}

// WithWriter sets the logger writer.
func (l *Logger) WithWriter(writer Writer) *Logger {
	l.writer = writer
	return l
}

// RecoverPanics returns if we should recover panics in logger listeners.
func (l *Logger) RecoverPanics() bool {
	return l.recoverPanics
}

// WithRecoverPanics sets the recoverPanics field.
func (l *Logger) WithRecoverPanics(value bool) *Logger {
	l.recoverPanics = value
	return l
}

// Flags returns the EventSet
func (l *Logger) Flags() *FlagSet {
	return l.flags
}

// WithFlags sets the logger flags.
func (l *Logger) WithFlags(flags *FlagSet) *Logger {
	l.flagsLock.Lock()
	defer l.flagsLock.Unlock()
	l.flags = flags
	return l
}

// WithFlagsFromEnv adds events from the environment.
func (l *Logger) WithFlagsFromEnv() *Logger {
	l.flagsLock.Lock()
	defer l.flagsLock.Unlock()
	if l.flags != nil {
		l.flags.CoalesceWith(NewFlagSetFromEnv())
	} else {
		l.flags = NewFlagSetFromEnv()
	}
	return l
}

// WithEnabled flips the bit flag for a given set of events.
func (l *Logger) WithEnabled(flags ...Flag) *Logger {
	l.Enable(flags...)
	return l
}

// Enable flips the bit flag for a given set of events.
func (l *Logger) Enable(flags ...Flag) {
	l.flagsLock.Lock()
	defer l.flagsLock.Unlock()

	if l.flags != nil {
		for _, flag := range flags {
			l.flags.Enable(flag)
		}
	} else {
		l.flags = NewFlagSet(flags...)
	}
}

// WithDisabled flips the bit flag for a given set of events.
func (l *Logger) WithDisabled(flags ...Flag) *Logger {
	l.Disable(flags...)
	return l
}

// Disable flips the bit flag for a given set of events.
func (l *Logger) Disable(flags ...Flag) {
	if l.flags == nil {
		return
	}

	l.flagsLock.Lock()
	defer l.flagsLock.Unlock()
	for _, flag := range flags {
		l.flags.Disable(flag)
	}
}

// IsEnabled asserts if a flag value is set or not.
func (l *Logger) IsEnabled(flag Flag) bool {
	if l.flags == nil {
		return false
	}
	return l.flags.IsEnabled(flag)
}

// HasListeners returns if there are registered listener for an event.
func (l *Logger) HasListeners(flag Flag) bool {
	if l == nil {
		return false
	}
	if l.workers == nil {
		return false
	}
	workers, hasWorkers := l.workers[flag]
	if !hasWorkers {
		return false
	}
	return len(workers) > 0
}

// HasListener returns if a specific listener is registerd for a flag.
func (l *Logger) HasListener(flag Flag, listenerName string) bool {
	if l == nil {
		return false
	}
	if l.workers == nil {
		return false
	}
	workers, hasWorkers := l.workers[flag]
	if !hasWorkers {
		return false
	}
	_, hasWorker := workers[listenerName]
	return hasWorker
}

// Listen adds a listener for a given flag.
func (l *Logger) Listen(flag Flag, listenerName string, listener Listener) {
	l.workersLock.Lock()
	defer l.workersLock.Unlock()

	if l.workers == nil {
		l.workers = map[Flag]map[string]*Worker{}
	}

	w := NewWorker(l, listener).WithRecoverPanics(l.recoverPanics)
	if listeners, hasListeners := l.workers[flag]; hasListeners {
		listeners[listenerName] = w
	} else {
		l.workers[flag] = map[string]*Worker{
			listenerName: w,
		}
	}
	w.Start()
}

// RemoveListeners clears *all* listeners for a Flag.
func (l *Logger) RemoveListeners(flag Flag) {
	if l.workers == nil {
		return
	}

	l.workersLock.Lock()
	defer l.workersLock.Unlock()

	listeners, hasListeners := l.workers[flag]
	if !hasListeners {
		return
	}

	for _, w := range listeners {
		w.Close()
	}

	delete(l.workers, flag)
}

// RemoveListener clears a specific listener for a Flag.
func (l *Logger) RemoveListener(flag Flag, listenerName string) {
	if l.workers == nil {
		return
	}

	l.workersLock.Lock()
	defer l.workersLock.Unlock()

	listeners, hasListeners := l.workers[flag]
	if !hasListeners {
		return
	}

	worker, hasWorker := listeners[listenerName]
	if !hasWorker {
		return
	}

	worker.Close()
	delete(listeners, listenerName)

	if len(listeners) == 0 {
		delete(l.workers, flag)
	}
}

// Trigger fires the currently configured event listeners.
func (l *Logger) Trigger(e Event) {
	if l == nil {
		return
	}
	if e == nil {
		return
	}
	if l.flags == nil {
		return
	}
	l.ensureInitialized()

	flag := e.Flag()
	if l.flags.IsEnabled(flag) {
		if l.workers != nil {
			if workers, hasWorkers := l.workers[flag]; hasWorkers {
				for _, worker := range workers {
					worker.Work <- e
				}
			}
		}

		if typed, isTyped := e.(EventWritable); isTyped && !typed.ShouldWrite() {
			return
		}

		if typed, isTyped := e.(EventErrorWritable); isTyped && typed.IsError() {
			l.writeErrorWorker.Work <- e
		} else {
			l.writeWorker.Work <- e
		}
	}

}

// SyncTrigger fires the currently configured event listeners synchronously.
func (l *Logger) SyncTrigger(e Event) {
	if l == nil {
		return
	}
	if e == nil {
		return
	}
	if l.flags == nil {
		return
	}

	flag := e.Flag()
	if l.flags.IsEnabled(flag) {
		if workers, hasWorkers := l.workers[flag]; hasWorkers {
			for _, worker := range workers {
				worker.Listener(e)
			}
		}

		if typed, isTyped := e.(EventWritable); isTyped && !typed.ShouldWrite() {
			return
		}

		if typed, isTyped := e.(EventErrorWritable); isTyped && typed.IsError() {
			l.WriteError(e)
		} else {
			l.Write(e)
		}
	}
}

// Sillyf logs an incredibly verbose message to the output stream.
func (l *Logger) Sillyf(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.Trigger(Messagef(Silly, format, args...))
}

// SyncSillyf logs an incredibly verbose message to the output stream synchronously.
func (l *Logger) SyncSillyf(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.SyncTrigger(Messagef(Silly, format, args...))
}

// Infof logs an informational message to the output stream.
func (l *Logger) Infof(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.Trigger(Messagef(Info, format, args...))
}

// SyncInfof logs an informational message to the output stream synchronously.
func (l *Logger) SyncInfof(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.SyncTrigger(Messagef(Info, format, args...))
}

// Debugf logs a debug message to the output stream.
func (l *Logger) Debugf(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.Trigger(Messagef(Debug, format, args...))
}

// SyncDebugf logs an debug message to the output stream synchronously.
func (l *Logger) SyncDebugf(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.SyncTrigger(Messagef(Debug, format, args...))
}

// Warningf logs a debug message to the output stream.
func (l *Logger) Warningf(format string, args ...interface{}) error {
	if l == nil {
		return nil
	}
	return l.Warning(fmt.Errorf(format, args...))
}

// SyncWarningf logs an warning message to the output stream synchronously.
func (l *Logger) SyncWarningf(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.SyncTrigger(Errorf(Warning, format, args...))
}

// Warning logs a warning error to std err.
func (l *Logger) Warning(err error) error {
	if l != nil {
		l.Trigger(NewError(Warning, err))
	}
	return err
}

// SyncWarning synchronously logs a warning to std err.
func (l *Logger) SyncWarning(err error) error {
	if l != nil {
		l.SyncTrigger(NewError(Warning, err))
	}
	return err
}

// WarningWithReq logs a warning error to std err with a request.
func (l *Logger) WarningWithReq(err error, req *http.Request) error {
	if l != nil {
		l.Trigger(NewErrorWithState(Warning, err, req))
	}
	return err
}

// Errorf writes an event to the log and triggers event listeners.
func (l *Logger) Errorf(format string, args ...interface{}) error {
	if l == nil {
		return nil
	}
	return l.Error(fmt.Errorf(format, args...))
}

// SyncErrorf synchronously triggers a error.
func (l *Logger) SyncErrorf(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.SyncTrigger(Errorf(Error, format, args...))
}

// Error logs an error to std err.
func (l *Logger) Error(err error) error {
	if l != nil {
		l.Trigger(NewError(Error, err))
	}
	return err
}

// SyncError synchronously logs an error to std err.
func (l *Logger) SyncError(err error) error {
	if l != nil {
		l.SyncTrigger(NewError(Error, err))
	}
	return err
}

// ErrorWithReq logs an error to std err with a request.
func (l *Logger) ErrorWithReq(err error, req *http.Request) error {
	if l != nil {
		l.Trigger(NewErrorWithState(Error, err, req))
	}
	return err
}

// Fatalf writes an event to the log and triggers event listeners.
func (l *Logger) Fatalf(format string, args ...interface{}) error {
	if l == nil {
		return nil
	}
	return l.Fatal(fmt.Errorf(format, args...))
}

// SyncFatalf synchronously triggers a fatal.
func (l *Logger) SyncFatalf(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.SyncTrigger(Errorf(Fatal, format, args...))
}

// Fatal logs the result of a panic to std err.
func (l *Logger) Fatal(err error) error {
	if l != nil {
		l.Trigger(NewError(Fatal, err))
	}
	return err
}

// SyncFatal synchronously logs a fatal to std err.
func (l *Logger) SyncFatal(err error) error {
	if l != nil {
		l.SyncTrigger(NewError(Fatal, err))
	}
	return err
}

// FatalWithReq logs the result of a fatal error to std err with a request.
func (l *Logger) FatalWithReq(err error, req *http.Request) error {
	if l != nil {
		l.Trigger(NewErrorWithState(Fatal, err, req))
	}
	return err
}

// SyncFatalExit logs the result of a fatal error to std err and calls `exit(1)`
func (l *Logger) SyncFatalExit(err error) {
	if l == nil || l.flags == nil {
		os.Exit(1)
	}

	l.Fatal(err)
	l.Drain()
	os.Exit(1)
}

// --------------------------------------------------------------------------------
// finalizers
// --------------------------------------------------------------------------------

// Close releases shared resources for the agent.
func (l *Logger) Close() (err error) {
	if l == nil {
		return nil
	}
	l.flagsLock.Lock()
	defer l.flagsLock.Unlock()

	if l.flags != nil {
		l.flags.SetNone()
	}

	l.workersLock.Lock()
	defer l.workersLock.Unlock()

	for _, workers := range l.workers {
		for _, worker := range workers {
			worker.Close()
		}
	}

	if l.writeWorker != nil {
		l.writeWorkerLock.Lock()
		defer l.writeWorkerLock.Unlock()
		l.writeWorker.Close()
	}

	if l.writeErrorWorker != nil {
		l.writeErrorWorkerLock.Lock()
		defer l.writeErrorWorkerLock.Unlock()
		l.writeErrorWorker.Close()
	}

	return nil
}

// Drain waits for the agent to finish its queue of events before closing.
func (l *Logger) Drain() error {
	if l == nil {
		return nil
	}

	l.workersLock.Lock()
	defer l.workersLock.Unlock()

	for _, workers := range l.workers {
		for _, worker := range workers {
			worker.Drain()
		}
	}

	if l.writeWorker != nil {
		l.writeWorkerLock.Lock()
		defer l.writeWorkerLock.Unlock()
		l.writeWorker.Drain()
	}

	if l.writeErrorWorker != nil {
		l.writeErrorWorkerLock.Lock()
		defer l.writeErrorWorkerLock.Unlock()
		l.writeErrorWorker.Drain()
	}

	return nil
}

// --------------------------------------------------------------------------------
// write helpers
// --------------------------------------------------------------------------------

func (l *Logger) ensureInitialized() {
	if l.writeWorker == nil {
		l.writeWorkerLock.Lock()
		defer l.writeWorkerLock.Unlock()
		if l.writeWorker == nil {
			l.writeWorker = NewWorker(l, l.Write)
			l.writeWorker.Start()
		}
	}
	if l.writeErrorWorker == nil {
		l.writeErrorWorkerLock.Lock()
		defer l.writeErrorWorkerLock.Unlock()
		if l.writeErrorWorker == nil {
			l.writeErrorWorker = NewWorker(l, l.WriteError)
			l.writeErrorWorker.Start()
		}
	}
}

// Write writes to the writer.
func (l *Logger) Write(e Event) {
	if l.writer != nil {
		l.writer.Write(e)
	}
}

// WriteError writes to the error writer.
func (l *Logger) WriteError(e Event) {
	if l.writer != nil {
		l.writer.WriteError(e)
	}
}
