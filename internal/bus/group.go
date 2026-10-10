package bus

import (
	"context"
	"sync"
	"time"
)

// Group запускает фоновые горутины владельца и завершает их вместе с ним.
// Отмена останавливает работу, ожидание не дает закрыть зависимости раньше горутин.
type Group struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu     sync.Mutex
	closed bool
}

// NewGroup создает группу горутин, привязанную к родительскому контексту.
func NewGroup(parent context.Context) *Group {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	return &Group{ctx: ctx, cancel: cancel}
}

// Context возвращает контекст группы: он отменяется при завершении.
func (g *Group) Context() context.Context {
	return g.ctx
}

// Go запускает горутину в группе. После завершения горутина не запускается.
func (g *Group) Go(fn func(ctx context.Context)) {
	g.mu.Lock()
	if g.closed {
		g.mu.Unlock()
		return
	}
	g.wg.Add(1)
	g.mu.Unlock()

	go func() {
		defer g.wg.Done()
		fn(g.ctx)
	}()
}

// GoCtx запускает горутину с контекстом, который отменяется вместе с группой
// или вместе с parent. Нужен для задач, привязанных и к владельцу, и к запросу.
func (g *Group) GoCtx(parent context.Context, fn func(ctx context.Context)) {
	g.mu.Lock()
	if g.closed {
		g.mu.Unlock()
		return
	}
	g.wg.Add(1)
	g.mu.Unlock()

	ctx, cancel := context.WithCancel(g.ctx)
	if parent != nil {
		stop := context.AfterFunc(parent, cancel)
		go func() {
			defer g.wg.Done()
			defer stop()
			defer cancel()
			fn(ctx)
		}()
		return
	}

	go func() {
		defer g.wg.Done()
		defer cancel()
		fn(ctx)
	}()
}

// Wait ждет завершения горутин группы, не отменяя их.
func (g *Group) Wait() {
	g.wg.Wait()
}

// Close отменяет горутины группы и ждет их завершения не дольше timeout.
// Возвращает false, если горутины не успели завершиться. Повторный вызов безопасен.
func (g *Group) Close(timeout time.Duration) bool {
	g.mu.Lock()
	if g.closed {
		g.mu.Unlock()
		return true
	}
	g.closed = true
	g.mu.Unlock()

	g.cancel()

	done := make(chan struct{})
	go func() {
		g.wg.Wait()
		close(done)
	}()

	if timeout <= 0 {
		<-done
		return true
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-done:
		return true
	case <-timer.C:
		return false
	}
}
