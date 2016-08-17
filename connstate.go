package cmd

import (
	"net"
	"sync"
)

const (
	stateMapCapacity = 50000
)

type activeConns struct {
	mu     sync.Mutex
	states map[net.Conn]struct{}
}

func newActiveConns() *activeConns {
	return &activeConns{
		states: make(map[net.Conn]struct{}, stateMapCapacity),
	}
}

func (a *activeConns) SetActive(conn net.Conn) {
	a.mu.Lock()
	a.states[conn] = struct{}{}
	a.mu.Unlock()
}

func (a *activeConns) SetInactive(conn net.Conn) bool {
	a.mu.Lock()
	_, ok := a.states[conn]
	if ok {
		delete(a.states, conn)
	}
	a.mu.Unlock()
	return ok
}
