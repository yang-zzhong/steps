package steps

type Adapter interface {
	State(state *State) error
	SaveState(state *State) error
}

type Task interface {
	Name() string
	Tasks() []Task
}

type task struct {
	handler func(step *Step)
	name    string
	tasks   []*task
}

type Plan struct {
	tasks []*task
}

func NewPlan(prepare func(plan *Plan)) *Plan {
	plan := &Plan{}
	prepare(plan)
	return plan
}

func NewTask(name string) *task {
	return &task{name: name}
}

func (task *task) Name() string {
	return task.name
}

func (task *task) Tasks() []Task {
	ret := make([]Task, len(task.tasks))
	for i, t := range task.tasks {
		ret[i] = t
	}
	return ret
}

func (task *task) With(t *task) *task {
	task.tasks = append(task.tasks, task)
	return task
}

func (task *task) Prepare(handler func(step *Step)) *task {
	task.handler = handler
	return task
}

func (plan *Plan) With(task *task) *Plan {
	plan.tasks = append(plan.tasks, task)
	return plan
}

func (plan *Plan) Tasks() []Task {
	ret := make([]Task, len(plan.tasks))
	for i, t := range plan.tasks {
		ret[i] = t
	}
	return ret
}

type Executer struct {
	adapter Adapter
	step    *Step
}

func NewExecuter(adapter Adapter) *Executer {
	return &Executer{adapter: adapter}
}

func (executer *Executer) Execute(plan *Plan) error {
	var state State
	if err := executer.adapter.State(&state); err != nil {
		return err
	}
	executer.step = New(&state)
	step := executer.step
	for _, task := range plan.tasks {
		executer.doTask(step, task)
	}
	return executer.adapter.SaveState(&state)
}

func (executer *Executer) doTask(step *Step, t *task) {
	if t.handler != nil {
		step.Do(t.name, t.handler)
	} else if len(t.tasks) > 0 {
		step.Do(t.name, func(step *Step) {
			for _, t := range t.tasks {
				executer.doTask(step, t)
			}
		})
	}
}

func (executer *Executer) State() *State {
	return executer.step.State()
}
