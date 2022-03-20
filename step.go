package steps

import (
	"reflect"
	"runtime"
	"strings"
	"time"
)

// Step
type Step struct {
	state     *State // state of step
	parent    *Step  // the parent step which generate this one
	lastChild *Step  // last executed step of the current step
}

// New new a step with state
func New(state *State) *Step {
	return &Step{state: state}
}

// Done mark the step is done successfully
func (s *Step) Done() *Step {
	now := time.Now()
	s.state.Err = ""
	s.state.DoneAt = &now
	return s
}

// Fail mark the step is done with fail
func (s *Step) Fail(err error) *Step {
	now := time.Now()
	s.state.Err = err.Error()
	s.state.DoneAt = &now
	return s
}

// With set the state.Info
func (s *Step) With(info interface{}) *Step {
	s.state.Info = info
	return s
}

// Info invoke handler with state.Info
func (s *Step) Info(handle func(interface{})) {
	handle(s.state.Info)
}

// Do do the step with a name
func (s *Step) Do(name string, do func(s *Step)) *Step {
	var startedAt *time.Time
	startedAt = s.state.StartedAt
	if startedAt == nil {
		now := time.Now()
		s.state.StartedAt = &now
	}
	if s.lastChild != nil && (s.lastChild.state.Proceeding() || s.lastChild.state.Failed()) {
		return s
	}
	var cur *State = nil
	for _, s := range s.state.States {
		if s.Name == name {
			cur = s
			break
		}
	}
	if cur == nil {
		cur = newState(name)

	}
	s.doState(cur, do)
	s.state.lock.Lock()
	defer s.state.lock.Unlock()
	s.state.States = append(s.state.States, cur)
	return s
}

// Dox do the step without a name, it will set state.Name with "" and you can not Get the state with state.Get
func (s *Step) DoX(do func(s *Step)) *Step {
	return s.Do("", do)
}

// DoR will auto get func name with reflect and runtime. it has performance issue, use it carefully
func (s *Step) DoR(do func(s *Step)) *Step {
	name := runtime.FuncForPC(reflect.ValueOf(do).Pointer()).Name()
	idx := strings.LastIndex(name, ".")
	return s.Do(name[idx+1:], do)
}

// State get the step's state
func (s *Step) State() *State {
	return s.state
}

func (s *Step) doState(newState *State, do func(s *Step)) {
	s.lastChild = &Step{state: newState, parent: s}
	if newState.DoneAt != nil {
		return
	}
	if newState.StartedAt == nil {
		now := time.Now()
		newState.StartedAt = &now
	}
	do(s.lastChild)
	s.lastChild.syncState()
}

// sync state with parents
func (s *Step) syncState() {
	t := s
	for t.parent != nil {
		t.parent.state.Err = s.state.Err
		t.parent.state.DoneAt = s.state.DoneAt
		t = t.parent
	}
}

func popFirst(cur *string) string {
	idx := strings.Index(*cur, ".")
	if idx > 0 {
		first := (*cur)[:idx]
		*cur = (*cur)[idx+1:]
		return first
	}
	ret := *cur
	*cur = ""
	return ret
}
