package steps

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func Test_state(t *testing.T) {
	s, do := step(nil, false)
	do()
	printx(s.State())
	isPathRight(t, s, "test")
	isPathRight(t, s, "test.step1.step1")
	isPathRight(t, s, "test.step1.step2")
	isPathRight(t, s, "test.step2.step1")
	isPathRight(t, s, "test.step2.step2")
	isPathRight(t, s, "test.step2.step3")
	isPathRight(t, s, "test.step3.step1")
	isPathRight(t, s, "test.step3.step2")
}

func Test_failed(t *testing.T) {
	s, do := step(nil, true)
	do()
	cs := s.State().Get("test.step2.step2")
	if !(cs != nil && cs.Errs[0] == "failed because withFail setted") {
		t.Fatalf("state is wrong when failed")
	}
	if s.State().Get("test.step2.step3") != nil {
		t.Fatalf("next step executed after failed")
	}
	if s.State().Get("test.step3") != nil {
		t.Fatalf("next steps executed after failed")
	}
}

func Test_recover(t *testing.T) {
	s, do := step(nil, true)
	do()
	s.State().Recover()
	cs := s.State().Get("test.step2.step2")
	if len(cs.Errs) != 0 || cs.DoneAt != nil {
		t.Fatalf("recover failed")
	}
	cs = s.State().Get("test.step2")
	if len(cs.Errs) != 0 || cs.DoneAt != nil {
		t.Fatalf("parent recover failed")
	}
	cs = s.State().Get("test")
	if len(cs.Errs) != 0 || cs.DoneAt != nil {
		t.Fatalf("parent's parent recover failed")
	}
}

func Test_concurrence(t *testing.T) {
	s, do := step(nil, true)
	go func() {
		for {
			json.Marshal(s.State())
		}
	}()
	for i := 0; i < 1000000; i++ {
		go do()
	}
}

func helloWorld(s *Step) {}

func Test_dor(t *testing.T) {
	s := New(&State{Name: "test"})
	s.DoR(helloWorld)
	if s.State().Get("test.helloWorld") == nil {
		t.Fatalf("auto get func name failed")
	}
}

func Test_async(t *testing.T) {
	s := New(&State{Name: "test"})
	s.Async(func() {
		s.Do("work1", func(s *Step) {
			s.Done()
		})
		s.Do("work2", func(s *Step) {
			s.Done()
		})
		s.Do("work3", func(s *Step) {
			s.Done()
		})
	})
	isPathRight(t, s, "test")
	isPathRight(t, s, "test.work1")
	isPathRight(t, s, "test.work2")
	isPathRight(t, s, "test.work3")
}

func Test_async_fail(t *testing.T) {
	s := New(&State{Name: "test"})
	s.Async(func() {
		s.Do("work1", func(s *Step) {
			s.Fail(errors.New("failed"))
		})
		s.Do("work2", func(s *Step) {
			s.Fail(errors.New("failed"))
		})
		s.Do("work3", func(s *Step) {
			s.Fail(errors.New("failed"))
		})
	})
	if len(s.State().Errs) != 3 {
		t.Fatal("async fail errs error")
	}
}

func isPathRight(t *testing.T, s *Step, path string) {
	idx := strings.LastIndex(path, ".")
	if s.State().Get(path).Name != path[idx+1:] {
		t.Fatalf("%s error", path)
	}
}

func step(state *State, withFail bool) (*Step, func()) {
	if state == nil {
		state = &State{Name: "test", Info: "an test state"}
	}
	step := New(state)
	return step, func() {
		step.Do("step1", func(s *Step) {
			s.Do("step1", func(s *Step) {
				s.With("step 1 executed").Done()
			})
			s.Do("step2", func(s *Step) {
				s.With("step 2 executed").Done()
			})
		})
		step.Do("step2", func(s *Step) {
			s.Do("step1", func(s *Step) {
				s.With("step 1 executed").Done()
			})
			s.Do("step2", func(s *Step) {
				if withFail {
					s.Fail(errors.New("failed because withFail setted"))
					return
				}
				s.Done()
			})
			s.Do("step3", func(s *Step) {
				s.Done()
			})
		})
		step.Do("step3", func(s *Step) {
			s.Do("step1", func(s *Step) {
				s.With("step 1 executed").Done()
			})
			s.Do("step2", func(s *Step) {
				s.Done()
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

func Benchmark_Step_Do(b *testing.B) {
	s := New(&State{Name: "hello"})
	for i := 0; i < b.N; i++ {
		s.Do("", func(s *Step) {})
	}
}

func Benchmark_Step_DoX(b *testing.B) {
	s := New(&State{Name: "hello"})
	for i := 0; i < b.N; i++ {
		s.DoX(func(s *Step) {})
	}
}

func Benchmark_Step_DoR(b *testing.B) {
	s := New(&State{Name: "test"})
	for i := 0; i < b.N; i++ {
		s.DoR(func(s *Step) {})
	}
}

func Benchmark_State_Recover(b *testing.B) {
	s, do := step(nil, true)
	do()
	for i := 0; i < b.N; i++ {
		s.State().Recover()
	}
}

func Benchmark_Step_Done(b *testing.B) {
	s := New(&State{Name: "hello"})
	for i := 0; i < b.N; i++ {
		s.Done()
	}
}

func Benchmark_Step_Fail(b *testing.B) {
	s := New(&State{Name: "hello"})
	err := errors.New("hello")
	for i := 0; i < b.N; i++ {
		s.Fail(err)
	}
}
