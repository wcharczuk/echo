package web

import (
	"os"
	"os/signal"
	"syscall"
)

// Shutdowner is a server that can start and shutdown.
type Shutdowner interface {
	Start() error
	Shutdown() error
}

// GracefulShutdown is an alias to StartWithGracefulShutdown.
var GracefulShutdown = StartWithGracefulShutdown

// StartWithGracefulShutdown starts an app and responds to SIGINT and SIGTERM to shut the app down.
// It will return any errors returned by app.Start() that are not caused by shutting down the server.
func StartWithGracefulShutdown(app Shutdowner) error {
	terminateSignal := make(chan os.Signal, 1)
	signal.Notify(terminateSignal, os.Interrupt, syscall.SIGTERM)
	return startWithGracefulShutdownBySignal(app, terminateSignal)
}

func startWithGracefulShutdownBySignal(app Shutdowner, terminateSignal chan os.Signal) error {
	shutdown := make(chan struct{})
	shutdownAbort := make(chan struct{})
	shutdownComplete := make(chan struct{})
	server := make(chan struct{})
	errors := make(chan error, 2)

	go func() {
		if err := app.Start(); err != nil {
			errors <- err
		}
		close(server)
	}()

	go func() {
		select {
		case <-shutdown:
			if err := app.Shutdown(); err != nil {
				errors <- err
			}
			close(shutdownComplete)
			return
		case <-shutdownAbort:
			close(shutdownComplete)
			return
		}
	}()

	select {
	case <-terminateSignal: // if we've issued a shutdown, wait for the server to exit
		close(shutdown)
		<-shutdownComplete
		<-server
	case <-server: // if the server exited
		close(shutdownAbort) // quit the signal listener
		<-shutdownComplete
	}

	if len(errors) > 0 {
		return <-errors
	}
	return nil
}
