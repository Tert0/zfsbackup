package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func ExecuteCommand(verbose bool, args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	outputError := stderr.String()
	if outputError != "" {
		return "", fmt.Errorf("%s", outputError)
	}
	if err != nil {
		return "", err
	}
	if verbose {
		fmt.Println(">>> " + strings.Join(args, " ") + " <<<")
		fmt.Print(output)
		fmt.Println(strings.Repeat("=", len(strings.Join(args, " "))+8))
	}
	return output, nil
}
