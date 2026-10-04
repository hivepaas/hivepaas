package dockerhelper

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/docker"
)

// A command line and its argv go both ways: what a screen shows is what a save
// splits back, quotes and all.
func TestCommandLineAndArgsGoBothWays(t *testing.T) {
	cases := map[string][]string{
		"plain words":                    {"node", "server.js"},
		"a shell's script":               {"sh", "-c", "npm run migrate && npm start"},
		"a variable for the container's": {"sh", "-c", "echo ${HOME} $PATH"},
		"quotes inside":                  {"echo", `it's "here"`},
		"an empty word":                  {"printf", ""},
	}
	for name, argv := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := CommandArgs(CommandLine(argv))
			assert.NoError(t, err)
			assert.Equal(t, argv, got)
		})
	}
	assert.Equal(t, "sh -c 'a && b'", CommandLine([]string{"sh", "-c", "a && b"}))
	assert.Empty(t, CommandLine(nil))
}

func TestContainerCommandApplySetsTheEntrypointAndTheCommand(t *testing.T) {
	cs := &swarm.ContainerSpec{Command: []string{"old"}, Args: []string{"old"}}
	assert.NoError(t, ContainerCommandApply(cs, "/bin/sh -c", `"echo hi && sleep 1"`))
	assert.Equal(t, []string{"/bin/sh", "-c"}, cs.Command)
	assert.Equal(t, []string{"echo hi && sleep 1"}, cs.Args)

	assert.NoError(t, ContainerCommandApply(cs, "", " "))
	assert.Nil(t, cs.Command, "empty: the image's entrypoint")
	assert.Nil(t, cs.Args, "and the image's command")

	assert.Error(t, ContainerCommandApply(cs, "", `echo "open`), "a quote left open")
}

func TestHealthcheckTest(t *testing.T) {
	cases := map[string]struct {
		mode    docker.HealthcheckMode
		command string
		want    []string
	}{
		"CMD is split": {docker.HealthcheckModeCmd, `curl -f 'http://localhost/a b'`,
			[]string{"CMD", "curl", "-f", "http://localhost/a b"}},
		"CMD-SHELL is handed whole": {docker.HealthcheckModeCmdShell, "pg_isready -U app",
			[]string{"CMD-SHELL", "pg_isready -U app"}},
		"no mode with a command is CMD-SHELL": {docker.HealthcheckModeInherit, "true",
			[]string{"CMD-SHELL", "true"}},
		"no mode and no command is the image's": {docker.HealthcheckModeInherit, "", nil},
		"NONE":                                  {docker.HealthcheckModeNone, "", []string{"NONE"}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := HealthcheckTest(tc.mode, tc.command)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
	_, err := HealthcheckTest("CMD-EXEC", "true")
	assert.ErrorIs(t, err, ErrHealthcheckMode)
	_, err = HealthcheckTest(docker.HealthcheckModeCmd, `curl "open`)
	assert.Error(t, err)
}

func TestHealthcheckCommand(t *testing.T) {
	mode, command := HealthcheckCommand([]string{"CMD", "curl", "-f", "http://localhost/a b"})
	assert.Equal(t, docker.HealthcheckModeCmd, mode)
	assert.Equal(t, `curl -f 'http://localhost/a b'`, command, "quoted, to split back")

	mode, command = HealthcheckCommand([]string{"CMD-SHELL", "pg_isready -U app"})
	assert.Equal(t, docker.HealthcheckModeCmdShell, mode)
	assert.Equal(t, "pg_isready -U app", command, "as it is")

	_, command = HealthcheckCommand([]string{"CMD-SHELL", "pg_isready", "-U", "app"})
	assert.Equal(t, "pg_isready -U app", command, "one split into words before: as it was typed")

	mode, command = HealthcheckCommand(nil)
	assert.Equal(t, docker.HealthcheckModeInherit, mode)
	assert.Empty(t, command)
}
