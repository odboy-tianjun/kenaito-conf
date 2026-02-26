package client

import "sync"

// ChangeHandler is called when configs change
type ChangeHandler func(changed map[string]string)

// Listener manages config change callbacks
type Listener struct {
	mu       sync.RWMutex
	handlers []ChangeHandler
}

func NewListener() *Listener {
	return &Listener{
		handlers: make([]ChangeHandler, 0),
	}
}

func (l *Listener) OnChange(handler ChangeHandler) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.handlers = append(l.handlers, handler)
}

func (l *Listener) Notify(changed map[string]string) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for _, h := range l.handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// log and continue
				}
			}()
			h(changed)
		}()
	}
}
