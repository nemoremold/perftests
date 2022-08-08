package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

var (
	onlyOneShutdownSignalHandler = make(chan struct{})

	// Shutdown signals include: SIGINT, SIGTERM, *SIGQUIT, *SIGKILL (we don't block signals marked with '*'):
	// - SIGINT:  can be blocked, it is a signal to terminate a process, usually requested by the user (e.g. via Ctrl-C).
	// - SIGTERM: can be blocked, it is a signal to terminate a process (e.g. via 'kill' command).
	// - SIGQUIT: can be blocked, it is a signal to terminate a process, generating a core dump before exiting. It is
	//            not recommended to do clean-ups when SIGQUIT is received, because the left-overs might be useful for
	//            further examination in conjunction with the core dump.
	// - SIGKILL: cannot be blocked, it is a signal to terminate a process that can't be ignored and that doesn't allow
	//            for execution of any clean-up routine.
	shutdownSignals = []os.Signal{
		os.Interrupt,
		syscall.SIGTERM,
	}
)

// NewContextWithShutdownSignalHandler returns a context that is safe-guarded from certain shutdown signals (SIGTERM and
// SIGINT). The first time such signal is received, it is blocked and the cancellation of the context is triggered. The
// second time such signal is received, the program will be terminated with exit code 1.
//   NOTE: This should only be called once because there should always be only one handler for such signals. Multiple
//         invocations of this function will cause panics.
func NewContextWithShutdownSignalHandler() context.Context {
	// Panics when called twice, thus ensuring only one signal handler exists.
	close(onlyOneShutdownSignalHandler)

	ctx, cancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, shutdownSignals...)

	go func() {
		<-signals
		cancel()
		<-signals
		os.Exit(1)
	}()

	return ctx
}
