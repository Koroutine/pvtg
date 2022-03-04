package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

func RunGit(commands [][]string) error {

	// Process git commands
	for _, c := range commands {
		errorBuffer := bytes.NewBufferString("")

		cmd := exec.Command("git", c...)
		cmd.Dir = dir
		cmd.Stderr = errorBuffer
		err := cmd.Run()

		errorOutput := errorBuffer.String()

		if strings.Contains(errorOutput, "unable to delete") && strings.Contains(errorOutput, "remote ref does not exist") {
			fmt.Println("Remote ref doesn't exist, skipping delete")
		} else if err != nil {
			CheckIfError(errors.New(errorOutput))
		}
	}

	return nil
}
