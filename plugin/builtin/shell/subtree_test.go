package shell

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/evergreen-ci/evergreen/command"
	"github.com/evergreen-ci/evergreen/plugin/plugintest"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSubtreeCleanup(t *testing.T) {
	Convey("With a tracked long-running shell command", t, func() {
		id := "testID"
		buf := &bytes.Buffer{}
		env := os.Environ()
		env = append(env, "EVR_TASK_ID=bogus")
		env = append(env, "EVR_AGENT_PID=12345")
		env = append(env, fmt.Sprintf("EVR_TASK_ID=%v", id))
		env = append(env, fmt.Sprintf("EVR_AGENT_PID=%v", os.Getpid()))
		localCmd := &command.LocalCommand{
			CmdString:   "while true; do sleep 1; done; echo 'finish'",
			Stdout:      buf,
			Stderr:      buf,
			ScriptMode:  true,
			Environment: env,
		}
		So(localCmd.Start(), ShouldBeNil)
		trackProcess(id, localCmd.Cmd.Process.Pid, &plugintest.MockLogger{})
		grip.Alert(int32(localCmd.Cmd.Process.Pid))
		grip.Alert(message.CollectProcessInfoWithChildren(int32(localCmd.Cmd.Process.Pid)))
		grip.AlertMany(message.CollectProcessInfoSelfWithChildren()...)

		Convey("running KillSpawnedProcs should kill the process before it finishes", func() {
			So(KillSpawnedProcs(id, &plugintest.MockLogger{}), ShouldBeNil)
			grip.Alert(int32(localCmd.Cmd.Process.Pid))
			grip.Alert(message.CollectProcessInfoWithChildren(int32(localCmd.Cmd.Process.Pid)))
			grip.AlertMany(message.CollectProcessInfoSelfWithChildren()...)
			So(localCmd.Cmd.Wait(), ShouldNotBeNil)
			So(buf.String(), ShouldNotContainSubstring, "finish")
		})
	})
}
