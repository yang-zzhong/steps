## Features

1. save the runtime state to storage then you can retry from undone steps following the stored state
2. generate execute report
3. task steps management
4. statistics for optimization of application
5. structure your logic with steps by steps execute

## Examples

```go
s := New(steps.State{Name:"MoveToVPC"})

func ExecuteLogic(s *steps.Step) {
    s.Do("initialize", func(s *Step) {
        s.Do("initialize CVM", func(s *Step) {
            if err := initCVM(); err != nil {
                if errors.Is(err, ErrTimeout) {
                    // not fail so you could execute logic again
                    return
                }
                // failed, if you wanna retry, you should invoke s.Recover() before execute logic
                s.Fail(err)
                return
            }
            // this steps is ok.
            s.Done()
        })
        s.Do("initialize ENIs", func(s *Step) {
            s.With("this is an optional steps, so always set it done").Done()
        })
    })
    s.Do("preFlight", func(s *Step) {
        s.Do("chkAccount", func(s *Step) {
            s.With("check whether account usable or not")
            if err := chkAccount(); err != nil {
                s.Fail(err)
            }
            s.Done()
        })
        s.Do("chkNetwork", func(s *Step) {
            s.Info(func(info interface{}) {
                var deadIps []string
                if info != nil {
                    deadIps = info.([]string)
                }
                deadIps, err := getDeadIps(deadIps)
                if err != nil {
                    s.Fail(err)
                    return
                }
                s.With(deadIps)
                if len(deadIps) == 0 {
                    s.Done()
                }
            })
        })
    })
    s.Do("inFlight", func(s *Step) {
        s.Do("migrateInstances", func(s *Step) {
            s.With("modify instances' network configuration").Done()
        })
        s.Do("cloneENIs", func(s *Step) {
            s.Done()
        })
    })
}

ExecuteLogic(s)

bts, _ := json.Marshal(s.State())
// save bts to storage
```

`bts` can save in samewhere you prefer. next time you can execute from last unfinished state.

```golang
// attempt to execute again by the last unfinished state
var state steps.State
json.Unmarshal(bts, &state)

s := steps.New(&state)

ExecuteLogic(s)
```

the jsoned state is something like below
```json
{
	"name": "MoveToVpc",
	"info": "move cvm to another vpc",
	"err": "Can't check network",
	"startedAt": "2022-03-19T16:54:44.826117+08:00",
	"doneAt": "2022-03-19T16:54:44.82612+08:00",
	"states": [
		{
			"name": "initialize",
			"info": null,
			"err": "",
			"startedAt": "2022-03-19T16:54:44.826117+08:00",
			"doneAt": "2022-03-19T16:54:44.826118+08:00",
			"states": [
				{
					"name": "initialize CVM",
					"info": "MUST init CVM before migration",
					"err": "",
					"startedAt": "2022-03-19T16:54:44.826117+08:00",
					"doneAt": "2022-03-19T16:54:44.826117+08:00",
					"states": []
				},
				{
					"name": "initialize ENIs",
					"info": "this is an optional steps, so always set it done",
					"err": "",
					"startedAt": "2022-03-19T16:54:44.826118+08:00",
					"doneAt": "2022-03-19T16:54:44.826118+08:00",
					"states": []
				}
			]
		},
		{
			"name": "preFlight",
			"info": null,
			"err": "Can't check network",
			"startedAt": "2022-03-19T16:54:44.826118+08:00",
			"doneAt": "2022-03-19T16:54:44.82612+08:00",
			"states": [
				{
					"name": "chkAccount",
					"info": "check whether account usable or not",
					"err": "",
					"startedAt": "2022-03-19T16:54:44.826119+08:00",
					"doneAt": "2022-03-19T16:54:44.826119+08:00",
					"states": []
				},
				{
					"name": "chkNetwork",
					"info": "check if network is good for migration",
					"err": "Can't check network",
					"startedAt": "2022-03-19T16:54:44.82612+08:00",
					"doneAt": "2022-03-19T16:54:44.82612+08:00",
					"states": []
				}
			]
		}
	]
}
```

## Use it with K8S Operator Reconciler

This tool is perfectly matched with management of the k8s reconcile logic. because it can reconcile many times to achieve the last goal status. it will begin from undone steps in each reconciling

```golang
//+kubebuilder:rbac:groups=udious.com,resources=stepss/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=udious.com,resources=stepss/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Step object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *StepReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	var obj v1.YourObjType
	r.Get(context.Background(), okey, &obj)
	state := &steps.State{Name: "TestObjState"}
	if obj.State != "" {
		if err := json.Unmarshal(obj.State, &state); err != nil {
			return ctrl.Result{}, err
		}
	}
	s := steps.New(state)
	s.Do("step1", func(s *steps.Step) {
		// your step1 logic
	})
	s.Do("step2", func(s *steps.Step) {
		// your step2 logic
	})
	s.Do("step3", func(s *steps.Step) {
		// your step3 logic
	})
	s.Do("step4", func(s *steps.Step) {
		// your step4 logic
	})
	s.Do("step5", func(s *steps.Step) {
		// your step5 logic
	})
	var err error
	obj.State, err = json.Marshal(s.State())

	r.Update(context.Background(), obj)

	return ctrl.Result{Requeue: s.State().Proceeding()}, nil
}
```

## Sync async invokes

```golang

state := steps.State{Name:"sync"}

func invokeAsyncDo(s *steps.Step, asyncDo func() string, confirmAsyncDo func(taskId string) string) {
    s.Info(func(info interface{}) {
        if info == nil {
            taskId, err := asyncDo();
            if err != nil {
                if isFatal(err) {
                    s.Fail(err)
                }
                // retry next loop
                return
            }
            s.With(taskId)
            return
        }
        taskId := info.(string)
        state := confirmAsyncDo(taskId)
        switch state {
        case "Running":
            // reconfirm next time
            time.Sleep(time.Second)
        case "Fail":
            s.Fail(errors.New("async failed"))
        case "Success":
            s.Done()
        }
    })
}

func doTask() {
    splitInstancesRecovers := 0
    for !state.Done() && !state.Failed() {
        s := steps.New(state)
        s.Do("prepareInstances", func(s *steps.Step) {
            invokeAsyncDo(s, func() string {
                // invoke then return taskId
            }, func(taskId string) string {
                // use taskId to confirm whether it's done
            }) 
        })
        s.Do("splitInstances", func(s *steps.Step) {
            invokeAsyncDo(s, func() string {
                // invoke then return taskId
            }, func(taskId string) string {
                // use taskId to confirm whether it's done
            }) 
        })
        // splitIntances fail can recover
        if state.Failed() {
            // retry 5 times
            if splitInstancesRecovers < 5 {
                state.Recover()
                splitInstanceRecovers += 1
            }
        }
        s.Do("conbineInstances", func(s *steps.Step) {
            invokeAsyncDo(s, func() string {
                // invoke then return taskId
            }, func(taskId string) string {
                // use taskId to confirm whether it's done
            }) 
        })
        s.Do("confirmResult", func(s *steps.Step) {
            invokeAsyncDo(s, func() string {
                // invoke then return taskId
            }, func(taskId string) string {
                // use taskId to confirm whether it's done
            }) 
        })
    }
}

```


## performance

```
goos: darwin
goarch: amd64
pkg: github.com/yang-zzhong/steps
cpu: Intel(R) Core(TM) i5-8257U CPU @ 1.40GHz
Benchmark_Step_Do-8         	398724684	         3.049 ns/op
Benchmark_Step_DoX-8        	393528038	         3.061 ns/op
Benchmark_Step_DoR-8        	32740196	        32.54 ns/op
Benchmark_State_Recover-8   	146732630	         8.270 ns/op
Benchmark_Step_Done-8       	10623654	       121.9 ns/op
Benchmark_Step_Fail-8       	10088080	       127.3 ns/op
PASS
ok  	github.com/yang-zzhong/steps	11.189s
```
