package typeid

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
)

var (
	ErrInvalidValue      = errors.New("INVALID_VALUE")
	ErrValueTypeMismatch = errors.New("VALUE_TYPE_MISMATCH")
)

type FuncGenerator struct {
	store   map[uintptr]uint64
	counter atomic.Uint64
	mutex   sync.RWMutex
}

func NewFuncGenerator() *FuncGenerator {
	return &FuncGenerator{}
}

func (g *FuncGenerator) GenID(fn any) (uint64, error) {
	v := reflect.ValueOf(fn)

	if !v.IsValid() {
		return 0, ErrInvalidValue
	}

	if v.Kind() != reflect.Func {
		return 0, ErrValueTypeMismatch
	}

	ptr := v.Pointer()

	g.mutex.RLock()
	if id, ok := g.store[ptr]; ok {
		g.mutex.RUnlock()
		return id, nil
	}
	g.mutex.RUnlock()

	id := g.counter.Add(1)

	g.mutex.Lock()
	if id, ok := g.store[ptr]; ok {
		g.mutex.Unlock()
		return id, nil
	}
	g.mutex.Unlock()

	return id, nil
}

func (g *FuncGenerator) MustGenID(fn any) uint64 {
	id, err := g.GenID(fn)

	if err != nil {
		panic(err)
	}

	return id
}
