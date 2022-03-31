package steps

import (
	"fmt"
	"sync"
)

type status int

const (
	Unstarted = status(0)
	Started   = status(1)
	Done      = status(2)
)

type SState struct {
	Name   string    `json:"name"`
	Errs   []string  `json:"errs"`
	Status status    `json:"status"`
	States []*SState `json:"states"`

	statesLock sync.RWMutex
	errslock   sync.Mutex
}

var _ StateX = &SState{}

func (state *SState) Succeed() {
	state.Status = Done
	state.Errs = []string{}
}

func (state *SState) Fail(err error) {
	state.Status = Done
	state.Errs = []string{err.Error()}
}

func (state *SState) Failed() bool {
	return len(state.Errs) > 0 && state.Status == Done
}

func (state *SState) Succeeded() bool {
	return len(state.Errs) == 0 && state.Status == Done
}

func (state *SState) Proceeding() bool {
	return state.Status == Started
}

func (state *SState) Start() {
	state.Status = Started
}

// Started return whether step is started
func (state *SState) Started() bool {
	return state.Status == Started
}

func (state *SState) SyncResult(s StateX) {
	state.errslock.Lock()
	defer state.errslock.Unlock()
	for _, err := range s.(*SState).Errs {
		state.Errs = append(state.Errs, s.(*SState).Name+": "+err)
	}
	state.Status = s.(*SState).Status
}

func (state *SState) Derive(name string) StateX {
	s := &SState{Name: name, States: make([]*SState, 0)}
	state.Add(s)
	return s
}

func (state *SState) Add(s StateX) {
	state.statesLock.Lock()
	defer state.statesLock.Unlock()
	state.States = append(state.States, s.(*SState))
}

func (state *SState) StateAt(idx int) StateX {
	state.statesLock.RLock()
	defer state.statesLock.RUnlock()
	if len(state.States) > idx {
		return state.States[idx]
	}
	return nil
}

// Get get sub state thru dot divided string like test.hello.world
func (state *SState) Get(path string) StateX {
	if s := state.get(path); s != nil {
		return s
	}
	panic(fmt.Sprintf("path [%s] not found", path))
}

// Has check if path exists which means func of specified path has executed
func (state *SState) Has(path string) bool {
	return state.get(path) != nil
}

func (state *SState) get(path string) *SState {
	name := popFirst(&path)
	for _, s := range state.States {
		if name != s.Name {
			continue
		}
		if path == "" {
			return s
		}
		return s.get(path)
	}
	return nil
}
