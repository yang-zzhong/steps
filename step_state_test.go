package steps

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func Test_stepWithstate(t *testing.T) {
	s, do := stepWithState(nil, false)
	do()
	isPathRight(t, s, "step1.step1")
	isPathRight(t, s, "step1.step2")
	isPathRight(t, s, "step2.step1")
	isPathRight(t, s, "step2.step2")
	isPathRight(t, s, "step2.step3")
	isPathRight(t, s, "step3.step1")
	isPathRight(t, s, "step3.step2")
}

func Test_stepWithStateFailed(t *testing.T) {
	s, do := stepWithState(nil, true)
	do()
	cs := s.State().Get("step2.step2")
	if !(cs != nil && cs.(*State).Errs[0] == "failed because withFail setted") {
		t.Fatalf("state is wrong when failed")
	}
	if s.State().Has("step2.step3") {
		t.Fatalf("next step executed after failed")
	}
	if s.State().Has("step3") {
		t.Fatalf("next steps executed after failed")
	}
}

func Test_stepWithStateRecover(t *testing.T) {
	s, do := stepWithState(nil, true)
	do()
	s.State().(*State).Recover()
	cs := s.State().Get("step2.step2")
	if len(cs.(*State).Errs) != 0 || cs.(*State).DoneAt != nil {
		t.Fatalf("recover failed")
	}
	cs = s.State().Get("step2")
	if len(cs.(*State).Errs) != 0 || cs.(*State).DoneAt != nil {
		t.Fatalf("parent recover failed")
	}
	cs = s.State()
	if len(cs.(*State).Errs) != 0 || cs.(*State).DoneAt != nil {
		t.Fatalf("parent's parent recover failed")
	}
}

func helloWorld(s *Step) {}

func Test_stepWithStateDor(t *testing.T) {
	s := New(&State{})
	s.DoR(helloWorld)
	if s.State().Get("helloWorld") == nil {
		t.Fatalf("auto get func name failed")
	}
}

func Test_stepWithStateAsync(t *testing.T) {
	s := New(&State{Name: "test"})
	s.Async("step1", func() {
		s.Do("work1", func(s *Step) {
			s.Succeed()
		})
		s.Do("work2", func(s *Step) {
			s.Succeed()
		})
		s.Do("work3", func(s *Step) {
			s.Succeed()
		})
	})
	s.Do("step2", func(s *Step) {
		s.Succeed()
	})
	// printx(s.State())
	isPathRight(t, s, "step1/work1")
	isPathRight(t, s, "step1/work2")
	isPathRight(t, s, "step1/work3")
}

// func Test_stepWithStateAsyncFail(t *testing.T) {
// 	s := New(&State{})
// 	s.Async("step2", func() {
// 		s.Do("work1", func(s *Step) {
// 			s.Fail(errors.New("failed"))
// 		})
// 		s.Do("work2", func(s *Step) {
// 			s.Fail(errors.New("failed"))
// 		})
// 		s.Do("work3", func(s *Step) {
// 			s.Fail(errors.New("failed"))
// 		})
// 	})
// 	printx(s.State())
// 	if len(s.State().(*State).Errs) != 3 {
// 		t.Fatal("async fail errs error")
// 	}
// }

func isPathRight(t *testing.T, s *Step, path string) {
	if !s.State().Has(path) {
		t.Fatalf("%s error", path)
	}
}

func stepWithState(state *State, withFail bool) (*Step, func()) {
	if state == nil {
		state = &State{}
	}
	step := New(state)
	return step, func() {
		step.Do("step1", func(s *Step) {
			s.Do("step1", func(s *Step) {
				s.State().(*State).With("step 1 executed")
				s.Succeed()
			})
			s.Do("step2", func(s *Step) {
				s.State().(*State).With("step 2 executed")
				s.Succeed()
			})
		})
		step.Do("step2", func(s *Step) {
			s.Do("step1", func(s *Step) {
				s.State().(*State).With("step 1 executed")
				s.Succeed()
			})
			s.Do("step2", func(s *Step) {
				if withFail {
					s.Fail(errors.New("failed because withFail setted"))
					return
				}
				s.Succeed()
			})
			s.Do("step3", func(s *Step) {
				s.Succeed()
			})
		})
		step.Do("step3", func(s *Step) {
			s.Do("step1", func(s *Step) {
				s.State().(*State).With("step 1 executed")
				s.Succeed()
			})
			s.Do("step2", func(s *Step) {
				s.Succeed()
			})
		})
	}
}

func printx(state interface{}) {
	bs, _ := json.Marshal(state)
	var buf bytes.Buffer
	json.Indent(&buf, bs, "", "\t")
	fmt.Printf("%s\n", buf.String())
}

func Benchmark_StepWithState_Do(b *testing.B) {
	s := New(&State{})
	for i := 0; i < b.N; i++ {
		s.Do("", func(s *Step) {})
	}
}

func Benchmark_StepWithState_DoX(b *testing.B) {
	s := New(&State{})
	for i := 0; i < b.N; i++ {
		s.DoX(func(s *Step) {})
	}
}

func Benchmark_StepWithState_DoR(b *testing.B) {
	s := New(&State{})
	for i := 0; i < b.N; i++ {
		s.DoR(func(s *Step) {})
	}
}

func Benchmark_StepWithState_Recover(b *testing.B) {
	s, do := stepWithState(nil, true)
	do()
	state := s.State().(*State)
	for i := 0; i < b.N; i++ {
		state.Recover()
	}
}

func Benchmark_StepWithState_Done(b *testing.B) {
	s := New(&State{})
	for i := 0; i < b.N; i++ {
		s.Succeed()
	}
}

func Benchmark_StepWithState_Fail(b *testing.B) {
	s := New(&State{})
	err := errors.New("hello")
	for i := 0; i < b.N; i++ {
		s.Fail(err)
	}
}
