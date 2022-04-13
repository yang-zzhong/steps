package steps

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
)

func TestPhaseMachine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	adapter := NewMockAdapter(ctrl)
	adapter.EXPECT().SaveState(gomock.Any()).DoAndReturn(func(state *State) error {
		return nil
	})
	adapter.EXPECT().State(gomock.Any()).DoAndReturn(func(state *State) error {
		state = &State{Name: "test"}
		return nil
	})
	plan := NewPlan(func(plan *Plan) {
		task := NewTask("step1")
		task.Prepare(func(step *Step) {
			step.Done()
		})
		plan.With(task)
	})

	mc := NewExecuter(adapter)

	if err := mc.Execute(plan); err != nil {
		t.Fatalf("execute plan failed: %s", err.Error())
		return
	}
	if !mc.step.State().Succeeded() {
		t.Fatalf("execute plan should succeeded")
	}
}
