package app

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

// Run will execute an external command within the set Config.Path
// printing & returning all Stdout & Stderr
func run(bin string, args ...string) (string, error) {
	cmd := exec.Command(bin, args...)
	cmd.Dir = Config.RepoDir

	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)

	cmd.Stdout = mw
	cmd.Stderr = mw

	if err := cmd.Run(); err != nil {
		return stdBuffer.String(), err
	}

	return stdBuffer.String(), nil
}

// RunQuiet will execute an external command within the set Config.Path
// returning all Stdout & Stderr
func runQuiet(bin string, args ...string) (string, error) {
	cmd := exec.Command(bin, args...)
	cmd.Dir = Config.RepoDir

	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(&stdBuffer)

	cmd.Stdout = mw
	cmd.Stderr = mw

	if err := cmd.Run(); err != nil {
		return stdBuffer.String(), err
	}

	return stdBuffer.String(), nil
}
