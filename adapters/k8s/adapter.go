package k8s

import (
	"context"
	"sync"
	"time"

	"github.com/yang-zzhong/steps"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Adapter struct {
	Name     types.NamespacedName
	Resource client.Object
	Client   client.Client
	GetState func(client.Object) *K8State
	SetState func(client.Object, *K8State) client.Object

	rscGrabbed      bool
	rscGrabbledLock sync.Mutex
}

var _ steps.Adapter = &Adapter{}

type Info struct {
	Content interface{} `json:"content,omitempty"`
}

func (info *Info) DeepCopyInto(out *Info) {
	if info == nil {
		out = nil
		return
	}
	out.Content = runtime.DeepCopyJSONValue(info.Content)
}

func (info *Info) DeepCopy() *Info {
	if info == nil {
		return nil
	}
	out := new(Info)
	info.DeepCopyInto(out)
	return out
}

type K8State struct {
	// name is a important attr for state, it represents the State itself
	Name string `json:"name"`
	// this is for user logic state
	Info Info `json:"info"`
	// if step failed, this attr will record the err.Error()
	Errs []string `json:"errs"`
	// what time the step start to execute
	StartedAt *metav1.Time `json:"startedAt"`
	// what time the step is done, it will be filled whatever success or fail
	DoneAt *metav1.Time `json:"doneAt"`
	// sub step's states
	States []*K8State `json:"states"`
}

func (adapter *Adapter) State(state *steps.State) error {
	if _, err := adapter.GrabResource(); err != nil {
		return err
	}
	k8state := adapter.GetState(adapter.Resource)
	state = adapter.ToState(k8state)

	return nil
}

func (adapter *Adapter) GrabResource() (client.Object, error) {
	adapter.rscGrabbledLock.Lock()
	defer adapter.rscGrabbledLock.Unlock()
	if adapter.rscGrabbed {
		return adapter.Resource, nil
	}
	if err := adapter.Client.Get(context.TODO(), adapter.Name, adapter.Resource); err != nil {
		return nil, err
	}
	adapter.rscGrabbed = true
	return adapter.Resource, nil
}

func (adapter *Adapter) SaveState(state *steps.State) error {
	adapter.SetState(adapter.Resource, adapter.ToK8State(state))
	return adapter.Client.Status().Update(context.TODO(), adapter.Resource)
}

func (adapter *Adapter) ToState(k8state *K8State) *steps.State {
	return &steps.State{
		Name: k8state.Name,
		Info: k8state.Info.Content,
		Errs: k8state.Errs,
		StartedAt: func() *time.Time {
			if k8state.StartedAt != nil {
				return &k8state.StartedAt.Time
			}
			return nil
		}(),
		DoneAt: func() *time.Time {
			if k8state.DoneAt != nil {
				return &k8state.DoneAt.Time
			}
			return nil
		}(),
		States: func() []*steps.State {
			ret := make([]*steps.State, len(k8state.States))
			for i, s := range k8state.States {
				ret[i] = adapter.ToState(s)
			}
			return ret
		}(),
	}
}

func (adapter *Adapter) ToK8State(state *steps.State) *K8State {
	return &K8State{
		Name: state.Name,
		Info: Info{state.Info},
		Errs: state.Errs,
		StartedAt: func() *metav1.Time {
			if state.StartedAt != nil {
				t := metav1.NewTime(*state.StartedAt)
				return &t
			}
			return nil
		}(),
		DoneAt: func() *metav1.Time {
			if state.DoneAt != nil {
				t := metav1.NewTime(*state.DoneAt)
				return &t
			}
			return nil
		}(),
		States: func() []*K8State {
			ret := make([]*K8State, len(state.States))
			for i, s := range state.States {
				ret[i] = adapter.ToK8State(s)
			}
			return ret
		}(),
	}
}
