package bus

import (
	"sync"
	"sync/atomic"
)

type Handler func(payload any)
type UnsubscribeFunc func()

type subscription struct {
	id      uint64
	owner   string
	handler Handler
}

type Bus struct {
	mu     sync.RWMutex
	nextID atomic.Uint64
	topics map[string][]subscription
}

var globalBus = &Bus{
	topics: make(map[string][]subscription),
}

type Client struct {
	owner string
}

// Get возвращает контекст шины для конкретного модуля или плагина
func Get(owner string) *Client {
	return &Client{owner: owner}
}

// Clear полностью вычищает все подписки владельца
func Clear(owner string) {
	globalBus.clearOwner(owner)
}

func (c *Client) On(topic string, handler Handler) UnsubscribeFunc {
	return globalBus.subscribe(c.owner, topic, handler)
}

func (c *Client) Emit(topic string, payload any) {
	globalBus.publish(topic, payload)
}

func (c *Client) UnsubscribeAll() {
	globalBus.clearOwner(c.owner)
}

func (b *Bus) subscribe(owner, topic string, handler Handler) UnsubscribeFunc {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.nextID.Add(1)
	sub := subscription{id: id, owner: owner, handler: handler}
	b.topics[topic] = append(b.topics[topic], sub)

	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		b.removeSub(topic, id)
	}
}

func (b *Bus) publish(topic string, payload any) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sub := range b.topics[topic] {
		h := sub.handler
		go h(payload)
	}
}

func (b *Bus) clearOwner(owner string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for topic, subs := range b.topics {
		var filtered []subscription
		for _, s := range subs {
			if s.owner != owner {
				filtered = append(filtered, s)
			}
		}
		b.topics[topic] = filtered
	}
}

func (b *Bus) removeSub(topic string, id uint64) {
	subs := b.topics[topic]
	for i, s := range subs {
		if s.id == id {
			b.topics[topic] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
}
