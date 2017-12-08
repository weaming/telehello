package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func shellBasic(command string, args ...string) string {
	cmd := exec.Command(command, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("error: %v\nstderr: %v\n", err, stderr.String())
	}

	return stdout.String()
}

func ShellScript(script string) string {
	return shellBasic("bash", script)
}

func ShellCommand(command string) string {
	split := strings.Fields(command)
	l := len(split)
	//fmt.Printf("length of command %d", l)
	if l > 1 {
		return shellBasic(split[0], split[1:]...)
	} else {
		return shellBasic(split[0])
	}
}
