package cli

import (
	"errors"
	"io"
	"os/exec"
)

func runExternalCommand(name string, args []string, stdin io.Reader, stdout, stderr io.Writer) (handled bool, exitCode int, err error) {
	binary := "nucleus-" + name
	path, lookErr := exec.LookPath(binary)
	if lookErr != nil {
		return false, 0, nil
	}

	cmd := exec.Command(path, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if runErr := cmd.Run(); runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			return true, exitErr.ExitCode(), nil
		}
		return true, 1, runErr
	}

	return true, 0, nil
}
