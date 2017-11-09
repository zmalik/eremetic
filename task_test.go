package eremetic

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTask(t *testing.T) {
	Convey("WasRunning", t, func() {
		Convey("A task that was running", func() {
			task := Task{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
					Status{2, "TASK_FINISHED"},
				},
			}

			So(task.WasRunning(), ShouldBeTrue)
		})

		Convey("A task that is running", func() {
			task := Task{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
				},
			}

			So(task.WasRunning(), ShouldBeTrue)
		})

		Convey("A task that never was running", func() {
			task := Task{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_FAILED"},
				},
			}

			So(task.WasRunning(), ShouldBeFalse)
		})
	})

	Convey("IsTerminated", t, func() {
		Convey("A task that was running", func() {
			task := Task{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
					Status{2, "TASK_FINISHED"},
				},
			}

			So(task.IsTerminated(), ShouldBeTrue)
		})

		Convey("A task that is running", func() {
			task := Task{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
				},
			}

			So(task.IsTerminated(), ShouldBeFalse)
		})

		Convey("A task that never was running", func() {
			task := Task{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_FAILED"},
				},
			}

			So(task.IsTerminated(), ShouldBeTrue)
		})

		Convey("A empty task", func() {
			task := Task{}

			So(task.IsTerminated(), ShouldBeTrue)
		})
	})

	Convey("IsRunning", t, func() {
		Convey("A task that was running", func() {
			task := Task{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
					Status{2, "TASK_FINISHED"},
				},
			}

			So(task.IsRunning(), ShouldBeFalse)
		})

		Convey("A task that is running", func() {
			task := Task{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
				},
			}

			So(task.IsRunning(), ShouldBeTrue)
		})

		Convey("A empty task", func() {
			task := Task{}

			So(task.IsRunning(), ShouldBeFalse)
		})
	})

	Convey("LastUpdated", t, func() {
		Convey("A task that is running", func() {
			task := Task{
				Status: []Status{
					Status{1449682262, "TASK_STAGING"},
					Status{1449682265, "TASK_RUNNING"},
				},
			}

			s := task.LastUpdated()

			So(s.Unix(), ShouldEqual, 1449682265)
		})

		Convey("A empty task", func() {
			task := Task{}

			s := task.LastUpdated()

			So(s.Unix(), ShouldEqual, 0)
		})
	})

	Convey("NewTask", t, func() {
		request := Request{
			TaskCPUs:    0.5,
			TaskMem:     22.0,
			DockerImage: "busybox",
			Command:     "echo hello",
		}

		Convey("No volume or environment specified", func() {
			task, err := NewTask(request)

			So(err, ShouldBeNil)
			So(task, ShouldNotBeNil)
			So(task.Command, ShouldEqual, "echo hello")
			So(task.User, ShouldEqual, "root")
			So(task.Environment, ShouldBeEmpty)
			So(task.Image, ShouldEqual, "busybox")
			So(task.Volumes, ShouldBeEmpty)
			So(task.Status, ShouldHaveLength, 1)
			So(task.Status[0].Status, ShouldEqual, TaskQueued)
		})

		Convey("Given a volume and environment", func() {
			var volumes []Volume
			var environment = make(map[string]string)
			environment["foo"] = "bar"
			volumes = append(volumes, Volume{
				ContainerPath: "/var/www",
				HostPath:      "/var/www",
			})
			request.Volumes = volumes
			request.Environment = environment

			task, err := NewTask(request)

			So(err, ShouldBeNil)
			So(task.Environment, ShouldContainKey, "foo")
			So(task.Environment["foo"], ShouldEqual, "bar")
			So(task.Volumes[0].ContainerPath, ShouldEqual, volumes[0].ContainerPath)
			So(task.Volumes[0].HostPath, ShouldEqual, volumes[0].HostPath)
		})

		Convey("Given a masked environment", func() {
			var maskedEnv = make(map[string]string)
			maskedEnv["foo"] = "bar"

			request.MaskedEnvironment = maskedEnv
			task, err := NewTask(request)

			So(err, ShouldBeNil)
			So(task.MaskedEnvironment, ShouldContainKey, "foo")
			So(task.MaskedEnvironment["foo"], ShouldEqual, "bar")
		})

		Convey("Given a network type", func() {
			network := "HOST"

			request.Network = network
			task, err := NewTask(request)

			So(err, ShouldBeNil)
			So(task.Network, ShouldEqual, "HOST")
		})

		Convey("Given a dns", func() {
			dns := "172.17.0.1"

			request.DNS = dns
			task, err := NewTask(request)

			So(err, ShouldBeNil)
			So(task.DNS, ShouldEqual, "172.17.0.1")
		})

		Convey("Given URI (via uris) to download", func() {
			request.URIs = []string{"http://foobar.local/kitten.jpg"}

			task, err := NewTask(request)

			So(err, ShouldBeNil)
			So(task.FetchURIs, ShouldHaveLength, 1)
			So(task.FetchURIs[0].URI, ShouldEqual, request.URIs[0])
			So(task.FetchURIs[0].Executable, ShouldBeFalse)
			So(task.FetchURIs[0].Extract, ShouldBeFalse)
			So(task.FetchURIs[0].Cache, ShouldBeFalse)
		})

		Convey("Given URI (via uris) to download and extract", func() {
			request.URIs = []string{"http://foobar.local/kittens.tar.gz"}

			task, err := NewTask(request)

			So(err, ShouldBeNil)
			So(task.FetchURIs, ShouldHaveLength, 1)
			So(task.FetchURIs[0].URI, ShouldEqual, request.URIs[0])
			So(task.FetchURIs[0].Executable, ShouldBeFalse)
			So(task.FetchURIs[0].Extract, ShouldBeTrue)
			So(task.FetchURIs[0].Cache, ShouldBeFalse)
		})

		Convey("Given URI (via fetch) to download and extract", func() {
			request.Fetch = []URI{URI{
				URI:     "http://foobar.local/kittens.tar.gz",
				Extract: true,
			}}

			task, err := NewTask(request)

			So(err, ShouldBeNil)
			So(task.FetchURIs, ShouldHaveLength, 1)
			So(task.FetchURIs[0].URI, ShouldEqual, request.Fetch[0].URI)
			So(task.FetchURIs[0].Executable, ShouldBeFalse)
			So(task.FetchURIs[0].Extract, ShouldBeTrue)
			So(task.FetchURIs[0].Cache, ShouldBeFalse)
		})

		Convey("Given URI (via fetch) to download, cache and set executable", func() {
			request.Fetch = []URI{URI{
				URI:        "http://foobar.local/kittens.sh",
				Executable: true,
				Cache:      true,
			}}

			task, err := NewTask(request)

			So(err, ShouldBeNil)
			So(task.FetchURIs, ShouldHaveLength, 1)
			So(task.FetchURIs[0].URI, ShouldEqual, request.Fetch[0].URI)
			So(task.FetchURIs[0].Executable, ShouldBeTrue)
			So(task.FetchURIs[0].Extract, ShouldBeFalse)
			So(task.FetchURIs[0].Cache, ShouldBeTrue)
		})

		Convey("Given no Command", func() {
			request.Command = ""

			task, err := NewTask(request)

			So(err, ShouldBeNil)
			So(task.Command, ShouldBeEmpty)
		})

		Convey("Given Privileged", func() {
			request.Privileged = true

			task, err := NewTask(request)

			So(err, ShouldBeNil)
			So(task.Privileged, ShouldBeTrue)
		})

		Convey("Given Force Pull image", func() {
			request.ForcePullImage = true

			task, err := NewTask(request)

			So(err, ShouldBeNil)
			So(task.ForcePullImage, ShouldBeTrue)
		})

		Convey("Given a Name", func() {
			request.Name = "foobar"

			task, err := NewTask(request)

			So(err, ShouldBeNil)
			So(task.Name, ShouldEqual, "foobar")
		})

		Convey("Given a Label", func() {
			request.Labels = map[string]string{"label1": "label_value"}

			task, err := NewTask(request)

			So(err, ShouldBeNil)
			So(task.Labels["label1"], ShouldEqual, "label_value")
		})

		Convey("New task from empty request", func() {
			req := Request{}
			task, err := NewTask(req)

			So(err, ShouldBeNil)
			So(task, ShouldNotBeNil)
			So(task.WasRunning(), ShouldBeFalse)
			So(task.IsRunning(), ShouldBeFalse)
			So(task.IsTerminated(), ShouldBeFalse)
		})
	})
	Convey("states", t, func() {
		terminalStates := []TaskState{
			TaskFinished,
			TaskFailed,
			TaskKilled,
			TaskLost,
		}

		activeStates := []TaskState{
			TaskStaging,
			TaskStarting,
			TaskRunning,
			TaskError,
			TaskTerminating,
		}

		enqueuedStates := []TaskState{
			TaskQueued,
		}

		Convey("IsTerminal", func() {
			for _, state := range terminalStates {
				test := fmt.Sprintf("Should be true for %s", state)
				Convey(test, func() {
					So(IsTerminal(state), ShouldBeTrue)
				})
			}

			for _, state := range activeStates {
				test := fmt.Sprintf("Should be false for %s", state)
				Convey(test, func() {
					So(IsTerminal(state), ShouldBeFalse)
				})
			}

			for _, state := range enqueuedStates {
				test := fmt.Sprintf("Should be false for %s", state)
				Convey(test, func() {
					So(IsTerminal(state), ShouldBeFalse)
				})
			}
		})

		Convey("IsActive", func() {
			for _, state := range activeStates {
				test := fmt.Sprintf("Should be true for %s", state)
				Convey(test, func() {
					So(IsActive(state), ShouldBeTrue)
				})
			}

			for _, state := range terminalStates {
				test := fmt.Sprintf("Should be false for %s", state)
				Convey(test, func() {
					So(IsActive(state), ShouldBeFalse)
				})
			}

			for _, state := range enqueuedStates {
				test := fmt.Sprintf("Should be false for %s", state)
				Convey(test, func() {
					So(IsActive(state), ShouldBeFalse)
				})
			}
		})

		Convey("IsEnqueued", func() {
			for _, state := range enqueuedStates {
				test := fmt.Sprintf("Should be true for %s", state)
				Convey(test, func() {
					So(IsEnqueued(state), ShouldBeTrue)
				})
			}

			for _, state := range terminalStates {
				test := fmt.Sprintf("Should be false for %s", state)
				Convey(test, func() {
					So(IsEnqueued(state), ShouldBeFalse)
				})
			}

			for _, state := range activeStates {
				test := fmt.Sprintf("Should be false for %s", state)
				Convey(test, func() {
					So(IsEnqueued(state), ShouldBeFalse)
				})
			}
		})
	})

	Convey("task states", t, func() {
		task := Task{
			Name:   "foobar",
			Status: []Status{},
		}

		Convey("Task should not be in queued state", func() {
			So(task.IsEnqueued(), ShouldBeFalse)
		})
		Convey("Task should not be in active state", func() {
			So(task.IsActive(), ShouldBeFalse)
		})
		Convey("Task should be in terminated state", func() {
			So(task.IsTerminated(), ShouldBeTrue)
		})

		task.UpdateStatus(Status{0, "TASK_QUEUED"})
		Convey("TASK_QUEUED should be in queued state", func() {
			So(task.IsEnqueued(), ShouldBeTrue)
		})

		task.UpdateStatus(Status{0, "TASK_STAGING"})
		Convey("TASK_STAGING should be in active state", func() {
			So(task.IsActive(), ShouldBeTrue)
		})

		task.UpdateStatus(Status{0, "TASK_RUNNING"})
		Convey("TASK_RUNNING should be in active state", func() {
			So(task.IsActive(), ShouldBeTrue)
		})

		task.UpdateStatus(Status{0, "TASK_TERMINATING"})
		Convey("TASK_TERMINATING should be in terminating state", func() {
			So(task.IsTerminating(), ShouldBeTrue)
		})

		task.UpdateStatus(Status{0, "TASK_FAILED"})
		Convey("TASK_FAILED should be in terminated state", func() {
			So(task.IsTerminated(), ShouldBeTrue)
		})

		task.UpdateStatus(Status{0, "TASK_FINISHED"})
		Convey("TASK_FINISHED should be in terminated state", func() {
			So(task.IsTerminated(), ShouldBeTrue)
		})
	})

	Convey("task filter match", t, func() {

		task := Task{
			Name: "foobar",
			Status: []Status{
				Status{0, "TASK_STAGING"},
				Status{1, "TASK_RUNNING"},
				Status{2, "TASK_FINISHED"},
			},
		}

		Convey("Is Terminated", func() {
			taskFilter := TaskFilter{
				State: TerminatedState,
			}
			So(taskFilter.Match(&task), ShouldBeTrue)
		})

		Convey("Is Not Terminated", func() {
			task.Status = []Status{
				Status{0, "TASK_STAGING"},
			}
			taskFilter := TaskFilter{
				State: DefaultTaskFilterState,
			}
			So(taskFilter.Match(&task), ShouldBeTrue)
		})

		Convey("State doesn't match", func() {
			task.Status = []Status{
				Status{0, "TASK_STAGING"},
			}
			taskFilter := TaskFilter{
				State: "inventedState",
			}
			So(taskFilter.Match(&task), ShouldBeFalse)
		})

		Convey("Match Name", func() {
			taskFilter := TaskFilter{
				State: TerminatedState,
				Name:  "foobar",
			}
			So(taskFilter.Match(&task), ShouldBeTrue)
		})

		Convey("Doesn't Match Name", func() {
			taskFilter := TaskFilter{
				State: TerminatedState,
				Name:  "inventedName",
			}
			So(taskFilter.Match(&task), ShouldBeFalse)
		})

	})
}
