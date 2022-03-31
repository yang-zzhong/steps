package steps

import (
	"errors"
	"testing"
)

func Test_stepWithSState(t *testing.T) {
	s, do := stepWithSState(nil, false)
	do()
	isPathRight(t, s, "step1.step1")
	isPathRight(t, s, "step1.step2")
	isPathRight(t, s, "step2.step1")
	isPathRight(t, s, "step2.step2")
	isPathRight(t, s, "step2.step3")
	isPathRight(t, s, "step3.step1")
	isPathRight(t, s, "step3.step2")
}

func Test_stepWithSStateFailed(t *testing.T) {
	s, do := stepWithSState(nil, true)
	do()
	cs := s.State().Get("step2.step2")
	if !(cs != nil && cs.(*SState).Errs[0] == "failed because withFail setted") {
		t.Fatalf("state is wrong when failed")
	}
	if s.State().Has("step2.step3") {
		t.Fatalf("next step executed after failed")
	}
	if s.State().Has("step3") {
		t.Fatalf("next steps executed after failed")
	}
}

// func Test_recover(t *testing.T) {
// 	s, do := stepWithSState(nil, true)
// 	do()
// 	s.State().(*SState).Recover()
// 	cs := s.State().Get("step2.step2")
// 	if len(cs.(*State).Errs) != 0 || cs.(*State).DoneAt != nil {
// 		t.Fatalf("recover failed")
// 	}
// 	cs = s.State().Get("step2")
// 	if len(cs.(*State).Errs) != 0 || cs.(*State).DoneAt != nil {
// 		t.Fatalf("parent recover failed")
// 	}
// 	cs = s.State()
// 	if len(cs.(*State).Errs) != 0 || cs.(*State).DoneAt != nil {
// 		t.Fatalf("parent's parent recover failed")
// 	}
// }

func Test_dor(t *testing.T) {
	s := New(&SState{})
	s.DoR(helloWorld)
	if s.State().Get("helloWorld") == nil {
		t.Fatalf("auto get func name failed")
	}
}

func Test_async(t *testing.T) {
	s := New(&SState{})
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

// func Test_async_fail(t *testing.T) {
// 	s := New(&SState{})
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
// 	if len(s.State().(*SState).Errs) != 3 {
// 		t.Fatal("async fail errs error")
// 	}
// }

func stepWithSState(state *SState, withFail bool) (*Step, func()) {
	if state == nil {
		state = &SState{}
	}
	step := New(state)
	return step, func() {
		step.Do("step1", func(s *Step) {
			s.Do("step1", func(s *Step) {
				s.Succeed()
			})
			s.Do("step2", func(s *Step) {
				s.Succeed()
			})
		})
		step.Do("step2", func(s *Step) {
			s.Do("step1", func(s *Step) {
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
				s.Succeed()
			})
			s.Do("step2", func(s *Step) {
				s.Succeed()
			})
		})
	}
}

func Benchmark_StepWithSStateDo(b *testing.B) {
	s := New(&SState{})
	for i := 0; i < b.N; i++ {
		s.Do("", func(s *Step) {})
	}
}

func Benchmark_StepWithSStateDoX(b *testing.B) {
	s := New(&SState{})
	for i := 0; i < b.N; i++ {
		s.DoX(func(s *Step) {})
	}
}

func Benchmark_StepWithSStateDoR(b *testing.B) {
	s := New(&SState{})
	for i := 0; i < b.N; i++ {
		s.DoR(func(s *Step) {})
	}
}

// func Benchmark_StateWithSStateRecover(b *testing.B) {
// 	s, do := stepWithSState(nil, true)
// 	do()
// 	state := s.State().(*State)
// 	for i := 0; i < b.N; i++ {
// 		state.Recover()
// 	}
// }

func Benchmark_StepWithSStateDone(b *testing.B) {
	s := New(&SState{})
	for i := 0; i < b.N; i++ {
		s.Succeed()
	}
}

func Benchmark_StepWithSStateFail(b *testing.B) {
	s := New(&SState{})
	err := errors.New("hello")
	for i := 0; i < b.N; i++ {
		s.Fail(err)
	}
}
