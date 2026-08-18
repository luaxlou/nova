package registry

import (
	"fmt"
	"sync"
)

// Builder defines how a named instance is constructed.
type Builder[T any] func(name string) (T, error)

// Instance is the public handle returned by Registry.Named().
//
// It provides lifecycle operations while keeping startup lazy.
type Instance[T any] struct {
	name string
	r    *Registry[T]
}

// Registry unifies "Init -> Named -> Get/Reload/Close" behavior for starters.
type Registry[T any] struct {
	mu          sync.RWMutex
	defaultName string
	definitions map[string]Builder[T]
	instances   map[string]*entry[T]
}

type entry[T any] struct {
	once   sync.Once
	lock   sync.Mutex
	value  T
	err    error
	closed bool
}

// New creates an empty Registry.
func New[T any]() *Registry[T] {
	return &Registry[T]{
		definitions: make(map[string]Builder[T]),
		instances:   make(map[string]*entry[T]),
	}
}

// Init only sets default name and optional predefined definitions.
func (r *Registry[T]) Init(defaultName string, defs map[string]Builder[T]) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.defaultName = defaultName
	if defs != nil {
		r.definitions = make(map[string]Builder[T], len(defs))
		for name, fn := range defs {
			r.definitions[name] = fn
		}
	} else if r.definitions == nil {
		r.definitions = make(map[string]Builder[T])
	}
}

// Register adds/overwrites a builder definition for a named instance.
func (r *Registry[T]) Register(name string, builder Builder[T]) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.definitions == nil {
		r.definitions = map[string]Builder[T]{}
	}
	r.definitions[name] = builder
	delete(r.instances, name)
}

// RegisterMap merges builders in batch.
func (r *Registry[T]) RegisterMap(defs map[string]Builder[T]) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.definitions == nil {
		r.definitions = map[string]Builder[T]{}
	}

	for name, builder := range defs {
		r.definitions[name] = builder
		delete(r.instances, name)
	}
}

// Named returns an instance handle. Empty name resolves to default instance.
func (r *Registry[T]) Named(name string) *Instance[T] {
	if name == "" {
		name = r.defaultName
	}
	return &Instance[T]{name: name, r: r}
}

// Get is an explicit default-instance access helper.
func (r *Registry[T]) Get() *Instance[T] {
	return r.Named("")
}

// Definitions returns defined instance keys.
func (r *Registry[T]) Definitions() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	keys := make([]string, 0, len(r.definitions))
	for name := range r.definitions {
		keys = append(keys, name)
	}
	return keys
}

// Get resolves and returns the instance value.
func (h *Instance[T]) Get() (T, error) {
	return h.r.getOrBuild(h.name)
}

// Reload clears cache and rebuilds lazily on next Get.
func (h *Instance[T]) Reload() error {
	h.r.reset(h.name)
	_, err := h.Get()
	return err
}

// Close closes a single cached value if possible and removes it from cache.
func (h *Instance[T]) Close() error {
	h.r.mu.Lock()
	e, ok := h.r.instances[h.name]
	h.r.mu.Unlock()
	if !ok || e == nil {
		return nil
	}

	e.lock.Lock()
	if e.closed {
		e.lock.Unlock()
		return nil
	}

	var closeErr error
	if closer, ok := any(e.value).(interface{ Close() error }); ok {
		closeErr = closer.Close()
	}
	e.closed = true
	e.value = *new(T)
	e.err = nil
	e.once = sync.Once{}
	e.lock.Unlock()

	h.r.mu.Lock()
	delete(h.r.instances, h.name)
	h.r.mu.Unlock()

	return closeErr
}

// CloseAll closes all cached entries and resets registry state.
func (r *Registry[T]) CloseAll() error {
	r.mu.Lock()
	instances := make(map[string]*entry[T], len(r.instances))
	for name, e := range r.instances {
		instances[name] = e
	}
	r.instances = make(map[string]*entry[T])
	r.mu.Unlock()

	var err error
	for _, e := range instances {
		e.lock.Lock()
		if !e.closed {
			if closer, ok := any(e.value).(interface{ Close() error }); ok {
				if closeErr := closer.Close(); closeErr != nil && err == nil {
					err = closeErr
				}
			}
			e.closed = true
		}
		e.lock.Unlock()
	}

	return err
}

func (r *Registry[T]) reset(name string) {
	r.mu.Lock()
	delete(r.instances, name)
	r.mu.Unlock()
}

func (r *Registry[T]) getOrBuild(name string) (T, error) {
	r.ensureReady(name)

	r.mu.Lock()
	e := r.instances[name]
	builder := r.definitions[name]
	r.mu.Unlock()
	if builder == nil {
		var zero T
		return zero, fmt.Errorf("registry instance %q is not defined", name)
	}

	e.once.Do(func() {
		e.lock.Lock()
		defer e.lock.Unlock()
		e.value, e.err = builder(name)
	})

	if e.err != nil {
		return *new(T), e.err
	}

	e.lock.Lock()
	e.closed = false
	e.lock.Unlock()
	return e.value, nil
}

func (r *Registry[T]) ensureReady(name string) {
	r.mu.Lock()
	e, ok := r.instances[name]
	if ok && e != nil {
		r.mu.Unlock()
		return
	}

	if r.instances == nil {
		r.instances = make(map[string]*entry[T])
	}
	r.instances[name] = &entry[T]{closed: true}
	r.mu.Unlock()
}
