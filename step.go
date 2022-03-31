package steps

import (
	"reflect"
	"runtime"
	"strings"
	"sync"
)

type StateX interface {
	Start()
	Succeed()
	Fail(err error)

	Succeeded() bool
	Failed() bool
	Started() bool

	Derive(name string) StateX
	Add(StateX)
	SyncResult(StateX)

	StateAt(idx int) StateX

	Get(path string) StateX
	Has(path string) bool
}

// Step
type Step struct {
	state        StateX  // state of step
	parent       *Step   // the parent step which generate this one
	children     []*Step // last executed step of the current step
	asyncName    string
	inAsync      bool
	wg           sync.WaitGroup
	childrenLock sync.RWMutex
}

// New new a step with state
func New(state StateX) *Step {
	return &Step{state: state}
}

// Done mark the step is done successfully
func (s *Step) Succeed() *Step {
	s.state.Succeed()
	return s
}

// Fail mark the step is done with fail
func (s *Step) Fail(err error) *Step {
	s.state.Fail(err)
	return s
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
		if child.proceeding() || child.state.Failed() {
			return false
		}
	}
	return true
}

func (s *Step) proceeding() bool {
	return s.state.Started() && !s.state.Succeeded() && !s.state.Failed()
}

func (s *Step) do(name string, do func(s *Step)) *Step {
	if !s.inAsync && !s.newAvail() {
		return s
	}
	if !s.state.Started() {
		s.state.Start()
	}
	var cl int
	s.childrenLock.RLock()
	cl = len(s.children)
	s.childrenLock.RUnlock()
	cur := s.state.StateAt(cl)
	if cur != nil {
		s.state.Add(cur)
	} else {
		if s.inAsync {
			name = s.asyncName + "/" + name
		}
		cur = s.state.Derive(name)
	}
	s.doState(cur, do)
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
	s.wg = sync.WaitGroup{}
	do()
	s.wg.Wait()
	s.inAsync = false
	return s
}

// State get the step's state
func (s *Step) State() StateX {
	return s.state
}

func (s *Step) doState(newState StateX, do func(s *Step)) {
	step := &Step{state: newState, parent: s}
	if step.State().Succeeded() || step.State().Failed() {
		return
	}
	if !step.State().Started() {
		step.State().Start()
	}
	do(step)
	s.childrenLock.Lock()
	s.children = append(s.children, step)
	s.childrenLock.Unlock()
	step.syncState()
}

// sync state with parents
func (s *Step) syncState() {
	t := s
	for t.parent != nil {
		t.parent.state.SyncResult(s.state)
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
