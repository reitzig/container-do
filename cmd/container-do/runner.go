package main

import (
	"fmt"
	"time"
)

type runnerExec interface {
	DoesContainerExist(c container) (bool, error)
	IsContainerRunning(c container) (bool, error)
	CreateContainer(c container) error
	RestartContainer(c container) error
	ExecuteCommand(c container, commandAndArguments []string) error
}

/*
   We want to keep the container alive to avoid unnecessary overhead,
   but also eventually stop it unless it is used.
   We let the container run command check a token in a loop;
   when we run commands, we reset that token.

   Storing and checking the token _in_ the container avoids coupling to
   the host system: we do not pollute the project directory, and
   we do not have to keep any watchdog running.
*/

const keepAliveFile = "/keepalive" // TODO: make configurable?
const keepAliveIndefinitely = "running"

func nextContainerStopTime(c container) string {
	return fmt.Sprintf("%d", time.Now().Add(c.KeepAlive()).Unix())
}

func containerRunScript(c container) string {
	keepAliveString := fmt.Sprintf("%.0f", c.KeepAlive().Seconds())
	// NB: - Can not hard-code the first stop-time because otherwise `RestartContainer` wouldn't work!
	//     - `date +%s` outputs UNIX timestamps -- hopefully. Might break for non-standard implementations?
	//     - That other format seems to be needed in order for the addition to work; got that off Stack Overflow.
	//     - We use bash string comparison, which does the right thing for UNIX timestamps
	//       In particular, `keepAliveIndefinitely` is "larger" than any timestamp!
	return "date -d \"$(date '+%F %T') " + keepAliveString + " seconds\" +%s > " + keepAliveFile + "; " +
		"while [[ $(cat " + keepAliveFile + ") > $(date +%s) ]]; do sleep 1; done"
}

func setKeepAliveTokenScript(token string) string {
	return "echo '" + token + "' > " + keepAliveFile
}

func makeRunner(r runner) runnerExec {
	switch r {
	case docker:
		// TODO: check if docker exists?
		return DockerRunner{}
	case podman:
		panic("podman runner not yet implemented")
	default:
		panic("Invalid container runner: " + r)
	}
}
