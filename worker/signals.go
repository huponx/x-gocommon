package worker

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Signals cancels the returned context on SIGINT or SIGTERM.
// The consume loop should stop accepting work when ctx is done;
// in-flight jobs are not cancelled until Drain's deadline.
func Signals() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}
