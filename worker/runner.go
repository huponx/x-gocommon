package worker

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// Runner tracks in-flight jobs and drains them on shutdown.
// Job functions receive a context that keeps parent values (e.g. requestctx)
// and is cancelled only when Drain's deadline expires.
type Runner struct {
	logger   *zap.Logger
	mu       sync.Mutex
	stopped  bool
	wg       sync.WaitGroup
	jobs     context.Context
	stopJobs context.CancelFunc
}

func New(opts ...Option) *Runner {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	jobs, stopJobs := context.WithCancel(context.Background())
	return &Runner{
		logger:   o.logger,
		jobs:     jobs,
		stopJobs: stopJobs,
	}
}

// Go starts fn in a goroutine. Returns false if Drain has started;
// the caller should nack or requeue the message.
func (r *Runner) Go(ctx context.Context, fn func(context.Context) error) bool {
	if ctx == nil {
		ctx = context.Background()
	}
	r.mu.Lock()
	if r.stopped {
		r.mu.Unlock()
		return false
	}
	r.wg.Add(1)
	r.mu.Unlock()

	go func() {
		defer r.wg.Done()
		jobCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		stop := context.AfterFunc(r.jobs, cancel)
		defer stop()
		if err := fn(jobCtx); err != nil {
			r.logger.Error("worker job", zap.Error(err))
		}
	}()
	return true
}

// Drain stops accepting new jobs and waits for in-flight work.
// If ctx expires, in-flight job contexts are cancelled and Drain returns ctx.Err().
func (r *Runner) Drain(ctx context.Context) error {
	r.mu.Lock()
	r.stopped = true
	r.mu.Unlock()

	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		r.stopJobs()
		return ctx.Err()
	}
}
