package shell

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/evergreen-ci/evergreen/command"
	"github.com/evergreen-ci/evergreen/plugin/plugintest"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/tychoish/grip"
	"github.com/tychoish/grip/message"
)

func TestSubtreeCleanup(t *testing.T) {
	Convey("With a tracked long-running shell command", t, func() {
		id := "testID"
		buf := &bytes.Buffer{}
		env := os.Environ()
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
		grip.InfoMany(message.CollectProcessInfoSelfWithChildren()...)
		trackProcess(id, localCmd.Cmd.Process.Pid, &plugintest.MockLogger{})

		Convey("running KillSpawnedProcs should kill the process before it finishes", func() {
			So(KillSpawnedProcs(id, &plugintest.MockLogger{}), ShouldBeNil)
			grip.InfoMany(message.CollectProcessInfoSelfWithChildren()...)
			So(localCmd.Cmd.Wait(), ShouldNotBeNil)
			grip.InfoMany(message.CollectProcessInfoSelfWithChildren()...)
			So(buf.String(), ShouldNotContainSubstring, "finish")
		})
	})
}
