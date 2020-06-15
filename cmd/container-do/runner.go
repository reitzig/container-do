package main

import (
    "bufio"
    "bytes"
    "fmt"
    "go.uber.org/zap"
    "strings"
    "time"
)

type runnerExec interface {
    RunnerExecutable() string

    DetermineOsFlavor(c *container) error

    DoesContainerExist(c *container) (bool, error)
    IsContainerRunning(c *container) (bool, error)
    CreateContainer(c *container) error
    RestartContainer(c *container) error

    // Attach and block!
    ExecuteCommand(c *container, commandAndArguments []string) error
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

func parseOsReleaseFile(data []byte) (map[string]string, error) {
    stringMap := map[string]string{}

    scanner := bufio.NewScanner(bytes.NewReader(data))
    scanner.Split(bufio.ScanLines)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        parts := strings.SplitN(line, "=", 2)

        if len(parts) == 2 {
            stringMap[parts[0]] = strings.Trim(parts[1], "'\"")
        }
    }

    return stringMap, nil
}

func extractOsFlavorFromReleaseFile(out []byte) (string, error) {
    osMap, err := parseOsReleaseFile(out)
    if err != nil {
        return "", err
    }

    osFlavor := osMap["ID"]
    zap.L().Sugar().Debugf("Container OS ID: '%s'", osFlavor)
    if osNames, ok := osMap["ID_LIKE"]; ok {
        zap.L().Sugar().Debugf("Container OS ID_LIKE: '%s'", osNames)
        if osName, match := firstMatchInString(osNames, OsFlavors); match {
            osFlavor = osName
        }
    }

    if !stringInSlice(osFlavor, OsFlavors) {
        zap.L().Sugar().Warnf("Detected container OS flavor '%s' is not supported.", osFlavor)
    } else {
        zap.L().Sugar().Debugf("Detected container OS flavor '%s'.", osFlavor)
    }

    return osFlavor, nil
}

func nextContainerStopTime(c container) string {
    return fmt.Sprintf("%d", time.Now().Add(c.KeepAlive()).Unix())
}

func containerRunScript(c container) string {
    keepAliveString := fmt.Sprintf("%.0f", c.KeepAlive().Seconds())
    // NB: - Can not hard-code the first stop-time because otherwise `RestartContainer` wouldn't work!
    //     - `date +%s` outputs UNIX timestamps -- hopefully. Might break for non-standard implementations?
    //     - That other format seems to be needed in order for the addition to work; got that off Stack Overflow.
    //     - We use sh string comparison, which does the right thing for UNIX timestamps
    //       In particular, `keepAliveIndefinitely` is "larger" than any timestamp!

    dateCommand := "date -d '" + keepAliveString + "sec' +%s"
    switch c.osFlavor {
    case "gnu/linux", "debian", "fedora":
        break // default
    case "busybox", "alpine":
        dateCommand = "date -d@\"$(( $(date +%s)+" + keepAliveString + "))\" +%s"
    case "":
        zap.L().Panic("BUG: osFlavor not set")
    default:
        zap.L().Sugar().Warnf("OS Flavor '%s' not supported; assuming GNU compatibility!", c.osFlavor)
    }
    zap.L().Sugar().Debugf("Command to create first stop-time in container: `%s`", dateCommand)

    runCommand := dateCommand + " > " + keepAliveFile + "; " +
        "while [ $(cat " + keepAliveFile + ") \\> $(date +%s) ]; do sleep 1; done"
    zap.L().Sugar().Debugf("Command to start container with: `%s`", runCommand)

    return runCommand
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
