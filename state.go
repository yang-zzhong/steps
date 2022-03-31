package steps

import (
	"fmt"
	"sync"
	"time"
)

// State a execute state of step
type State struct {
	// name is a important attr for state, it represents the State itself
	Name string `json:"name"`
	// this is for user logic state
	Info interface{} `json:"info"`
	// if step failed, this attr will record the err.Error()
	Errs []string `json:"errs"`
	// what time the step start to execute
	StartedAt *time.Time `json:"startedAt"`
	// what time the step is done, it will be filled whatever success or fail
	DoneAt *time.Time `json:"doneAt"`
	// sub step's states
	States []*State `json:"states"`

	statesLock sync.RWMutex

	errsLock sync.Mutex
}

var _ StateX = &State{}

func newState(name string) *State {
	return &State{Name: name, States: make([]*State, 0)}
}

func (state *State) Succeed() {
	now := time.Now()
	state.Errs = []string{}
	state.DoneAt = &now
}

func (state *State) Fail(err error) {
	now := time.Now()
	state.Errs = []string{err.Error()}
	state.DoneAt = &now
}

func (state *State) With(info interface{}) {
	state.Info = info
}

// LastPath get last path of the state. it will return dot divided string like test.initializing.initEnv
func (state *State) LastPath() string {
	ret := state.Name
	if len(state.States) > 0 {
		ret += state.States[len(state.States)-1].LastPath()
	}
	return ret
}

// Failed return whether steps fail or not
func (state *State) Failed() bool {
	return state.DoneAt != nil && len(state.Errs) > 0
}

// Succeeded return whether steps succeed or not
func (state *State) Succeeded() bool {
	return state.DoneAt != nil && len(state.Errs) == 0
}

// Proceeding return whether steps under proceeding or not
func (state *State) Proceeding() bool {
	return state.StartedAt != nil && state.DoneAt == nil
}

// Started return whether step is started
func (state *State) Started() bool {
	return state.StartedAt != nil
}

func (state *State) Start() {
	now := time.Now()
	state.StartedAt = &now
}

// Get get sub state thru dot divided string like test.hello.world
func (state *State) Get(path string) StateX {
	if s := state.get(path); s != nil {
		return s
	}
	panic(fmt.Sprintf("path [%s] not found", path))
}

func (state *State) SyncResult(s StateX) {
	state.errsLock.Lock()
	defer state.errsLock.Unlock()
	for _, err := range s.(*State).Errs {
		state.Errs = append(state.Errs, s.(*State).Name+": "+err)
	}
	state.DoneAt = s.(*State).DoneAt
}

func (state *State) get(path string) *State {
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

// Has check if path exists which means func of specified path has executed
func (state *State) Has(path string) bool {
	return state.get(path) != nil
}

// Recover set steps execute result to unexecuted state
func (state *State) Recover() {
	state.Errs = []string{}
	state.DoneAt = nil
	if len(state.States) == 0 {
		state.StartedAt = nil
		return
	}
	state.States[len(state.States)-1].Recover()
}

func (state *State) Derive(name string) StateX {
	s := &State{Name: name, States: make([]*State, 0)}
	state.Add(s)
	return s
}

func (state *State) Add(s StateX) {
	state.statesLock.Lock()
	defer state.statesLock.Unlock()
	state.States = append(state.States, s.(*State))
}

func (state *State) StateAt(idx int) StateX {
	state.statesLock.RLock()
	defer state.statesLock.RUnlock()
	if len(state.States) > idx {
		return state.States[idx]
	}
	return nil
}
