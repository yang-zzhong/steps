package steps

import (
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Step
type Step struct {
	state        *State  // state of step
	parent       *Step   // the parent step which generate this one
	children     []*Step // last executed step of the current step
	asyncName    string
	inAsync      bool
	wg           sync.WaitGroup
	childrenLock sync.RWMutex
}

// New new a step with state
func New(state *State) *Step {
	return &Step{state: state}
}

// Done mark the step is done successfully
func (s *Step) Done() *Step {
	now := time.Now()
	s.state.Errs = []string{}
	s.state.DoneAt = &now
	return s
}

// Fail mark the step is done with fail
func (s *Step) Fail(err error) *Step {
	now := time.Now()
	s.state.Errs = []string{err.Error()}
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
	if !s.inAsync {
		return s.do(name, do)
	}
	s.wg.Add(1)
	go func() {
		s.do(name, do)
		s.wg.Done()
	}()
	return s
}

func (s *Step) newAvail() bool {
	s.childrenLock.RLock()
	defer s.childrenLock.RUnlock()
	for _, child := range s.children {
		if child.state.Proceeding() || child.state.Failed() {
			return false
		}
	}
	return true
}

func (s *Step) do(name string, do func(s *Step)) *Step {
	if !s.inAsync && !s.newAvail() {
		return s
	}
	if s.state.StartedAt == nil {
		now := time.Now()
		s.state.StartedAt = &now
	}
	var cur *State = s.state.stateAt(len(s.children))
	if cur == nil {
		if s.inAsync {
			name = s.asyncName + ":" + name
		}
		cur = newState(name)
	}
	s.doState(cur, do)
	s.state.append(cur)
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

// Async is an async do group
// example:
//   s.Do("hello", func(s *Step) {
//      fmt.Printf("hello\n")
//      s.Done()
//   })
//   s.Async(func() {
//      s.DoX(func(s *Step) {
//         fmt.Printf("\tearth\n")
//      })
//      s.DoX(func(s *Step) {
//         fmt.Printf("\tmars\n")
//      })
//      s.DoX(func(s *Step) {
//         fmt.Printf("\tvenus\n")
//      })
//   })
func (s *Step) Async(name string, do func()) *Step {
	s.inAsync = true
	s.asyncName = name
	do()
	s.wg.Wait()
	s.inAsync = false
	return s
}

// State get the step's state
func (s *Step) State() *State {
	return s.state
}

func (s *Step) doState(newState *State, do func(s *Step)) {
	step := &Step{state: newState, parent: s}
	if newState.DoneAt != nil {
		return
	}
	if newState.StartedAt == nil {
		now := time.Now()
		newState.StartedAt = &now
	}
	do(step)
	if s.inAsync {
		s.childrenLock.Lock()
		s.children = append(s.children, step)
		s.childrenLock.Unlock()
	} else {
		s.children = append(s.children, step)
	}
	step.syncState()
}

// sync state with parents
func (s *Step) syncState() {
	t := s
	for t.parent != nil {
		for _, err := range t.state.Errs {
			t.parent.state.Errs = append(t.parent.state.Errs, s.state.Name+": "+err)
		}
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
