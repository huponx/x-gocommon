package worker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestGoAndDrain(t *testing.T) {
	r := New()
	var ran atomic.Bool
	if !r.Go(context.Background(), func(context.Context) error {
		ran.Store(true)
		return nil
	}) {
		t.Fatal("Go rejected")
	}
	if err := r.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !ran.Load() {
		t.Fatal("job did not run")
	}
}

func TestGoRejectedAfterDrain(t *testing.T) {
	r := New()
	if err := r.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}
	if r.Go(context.Background(), func(context.Context) error { return nil }) {
		t.Fatal("expected reject after drain")
	}
}

func TestDrainTimeoutCancelsJob(t *testing.T) {
	r := New()
	started := make(chan struct{})
	cancelled := make(chan struct{})
	r.Go(context.Background(), func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		close(cancelled)
		return ctx.Err()
	})
	<-started

	drainCtx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := r.Drain(drainCtx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("drain err %v", err)
	}

	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("job ctx not cancelled")
	}
}

func TestGoKeepsParentValues(t *testing.T) {
	type key struct{}
	r := New()
	parent := context.WithValue(context.Background(), key{}, "corr")
	got := make(chan string, 1)
	r.Go(parent, func(ctx context.Context) error {
		v, _ := ctx.Value(key{}).(string)
		got <- v
		return nil
	})
	if err := r.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}
	if v := <-got; v != "corr" {
		t.Fatalf("value %q", v)
	}
}
