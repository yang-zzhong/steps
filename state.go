package steps

import (
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
	Err string `json:"err"`
	// what time the step start to execute
	StartedAt *time.Time `json:"startedAt"`
	// what time the step is done, it will be filled whatever success or fail
	DoneAt *time.Time `json:"doneAt"`
	// sub step's states
	States []*State `json:"states"`

	lock sync.Mutex
}

func newState(name string) *State {
	return &State{Name: name, States: make([]*State, 0)}
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
	return state.DoneAt != nil && state.Err != ""
}

// Succeeded return whether steps succeed or not
func (state *State) Succeeded() bool {
	return state.DoneAt != nil && state.Err == ""
}

// Proceeding return whether steps under proceeding or not
func (state *State) Proceeding() bool {
	return state.DoneAt == nil
}

// Get get sub state thru dot divided string like test.hello.world
func (state *State) Get(path string) *State {
	name := popFirst(&path)
	if name != state.Name {
		return nil
	}
	if path == "" {
		return state
	}
	for _, s := range state.States {
		if cs := s.Get(path); cs != nil {
			return cs
		}
	}
	return nil
}

// Recover set steps execute result to unexecuted state
func (state *State) Recover() {
	state.Err = ""
	state.DoneAt = nil
	if len(state.States) == 0 {
		state.StartedAt = nil
		return
	}
	state.States[len(state.States)-1].Recover()
}
