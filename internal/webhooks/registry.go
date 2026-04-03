package webhooks

import "sync"

var (
	providers = map[string]Provider{}
	mu        sync.RWMutex
)

func Register(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	providers[p.Name()] = p
}

func Get(name string) (Provider, bool) {
	mu.RLock()
	defer mu.RUnlock()
	p, ok := providers[name]
	return p, ok
}
