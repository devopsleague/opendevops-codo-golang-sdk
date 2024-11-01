package cascmd

import "sync"

type MemoryImpl struct {
	mu sync.Mutex
	m  map[string]string
}

func NewMemoryImpl() *MemoryImpl {
	return &MemoryImpl{
		m: make(map[string]string),
	}
}

func (x *MemoryImpl) CAS(key, src, dst string) bool {
	x.mu.Lock()
	defer x.mu.Unlock()
	if x.m[key] != src {
		return false
	}
	if dst == "" {
		delete(x.m, key)
		return true
	}
	x.m[key] = dst
	return true
}
