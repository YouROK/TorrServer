package bus

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// TestGroupWaitsForWorkers проверяет, что Close возвращается только после выхода горутин.
func TestGroupWaitsForWorkers(t *testing.T) {
	g := NewGroup(context.Background())
	var finished atomic.Bool

	g.Go(func(ctx context.Context) {
		<-ctx.Done()
		time.Sleep(50 * time.Millisecond)
		finished.Store(true)
	})

	if !g.Close(2 * time.Second) {
		t.Fatal("Close timed out waiting for the worker")
	}
	if !finished.Load() {
		t.Fatal("Close returned before the worker finished")
	}
}

// TestGroupCloseTimeout проверяет, что горутина без реакции на отмену не вешает Close.
func TestGroupCloseTimeout(t *testing.T) {
	g := NewGroup(context.Background())
	release := make(chan struct{})
	defer close(release)

	g.Go(func(ctx context.Context) {
		<-release
	})

	start := time.Now()
	if g.Close(100 * time.Millisecond) {
		t.Fatal("Close reported success while the worker was still running")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("Close waited too long: %v", elapsed)
	}
}

// TestGroupCloseIsIdempotent проверяет, что повторное завершение не паникует.
func TestGroupCloseIsIdempotent(t *testing.T) {
	g := NewGroup(context.Background())
	g.Go(func(ctx context.Context) { <-ctx.Done() })

	if !g.Close(time.Second) {
		t.Fatal("first Close failed")
	}
	if !g.Close(time.Second) {
		t.Fatal("second Close failed")
	}
}

// TestGroupRejectsWorkAfterClose проверяет, что после завершения горутина не стартует.
func TestGroupRejectsWorkAfterClose(t *testing.T) {
	g := NewGroup(context.Background())
	g.Close(time.Second)

	var started atomic.Bool
	g.Go(func(ctx context.Context) { started.Store(true) })

	time.Sleep(50 * time.Millisecond)
	if started.Load() {
		t.Fatal("worker started after Close")
	}
}

// TestGroupContextCancelsWorkers проверяет, что отмена доходит до горутин.
func TestGroupContextCancelsWorkers(t *testing.T) {
	g := NewGroup(context.Background())
	stopped := make(chan struct{})

	g.Go(func(ctx context.Context) {
		<-ctx.Done()
		close(stopped)
	})

	time.Sleep(20 * time.Millisecond)
	if !g.Close(time.Second) {
		t.Fatal("Close failed")
	}

	select {
	case <-stopped:
	default:
		t.Fatal("worker did not observe cancellation")
	}
}

// TestGroupParentContextCancels проверяет, что отмена родителя останавливает группу.
func TestGroupParentContextCancels(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	g := NewGroup(parent)
	stopped := make(chan struct{})

	g.Go(func(ctx context.Context) {
		<-ctx.Done()
		close(stopped)
	})

	cancelParent()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("parent cancellation did not reach the worker")
	}

	if !g.Close(time.Second) {
		t.Fatal("Close failed")
	}
}

// TestGroupGoCtxStopsOnParent проверяет, что отмена родительского контекста
// останавливает задачу, привязанную и к группе, и к запросу.
func TestGroupGoCtxStopsOnParent(t *testing.T) {
	g := NewGroup(context.Background())
	parent, cancelParent := context.WithCancel(context.Background())
	stopped := make(chan struct{})

	g.GoCtx(parent, func(ctx context.Context) {
		<-ctx.Done()
		close(stopped)
	})

	cancelParent()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("parent cancellation did not reach the GoCtx worker")
	}
	if !g.Close(time.Second) {
		t.Fatal("Close failed")
	}
}

// TestGroupWaitDoesNotCancel проверяет, что Wait ждет горутины, но не отменяет их.
func TestGroupWaitDoesNotCancel(t *testing.T) {
	g := NewGroup(context.Background())
	release := make(chan struct{})
	done := make(chan struct{})

	g.Go(func(ctx context.Context) {
		<-release
		close(done)
	})

	waited := make(chan struct{})
	go func() {
		g.Wait()
		close(waited)
	}()

	select {
	case <-waited:
		t.Fatal("Wait returned before the worker finished")
	case <-time.After(50 * time.Millisecond):
	}

	close(release)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not finish")
	}
	select {
	case <-waited:
	case <-time.After(time.Second):
		t.Fatal("Wait did not return after the worker finished")
	}
}
