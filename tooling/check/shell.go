package check

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

type ShellCheck struct {
	BodyName string
	// TODO: this is going to be a side effect for now.
	Envs map[string]string
	Cmd  string
}

func Shell() ShellCheck {
	return ShellCheck{
		Envs: make(map[string]string),
	}
}

func (c ShellCheck) WithBody(name string) ShellCheck {
	c.BodyName = name
	return c
}

func (c ShellCheck) With(name string, value string, rest ...any) ShellCheck {
	v := fmt.Sprintf(value, rest...)
	c.Envs[name] = v
	return c
}

func (c ShellCheck) Succeeds(cmd string) ShellCheck {
	if c.Cmd != "" {
		panic("Calling Succeeds twice is not supported yet.")
	}

	c.Cmd = cmd
	return c
}

func (c ShellCheck) Check(v []byte) CheckOutput {
	// Generate a temporary folder and a temporary body file name.
	tmpDir, err := os.MkdirTemp("", "shellcheck")
	if err != nil {
		return CheckOutput{Err: err}
	}
	defer os.RemoveAll(tmpDir)

	tmpFile, err := os.CreateTemp(tmpDir, "body-*.txt")
	if err != nil {
		return CheckOutput{Err: err}
	}
	defer os.Remove(tmpFile.Name())

	// Copy the body to the temporary file name.
	_, err = tmpFile.Write(v)
	if err != nil {
		return CheckOutput{Err: err}
	}

	// Set the env "BodyName" to the temporary file name.
	c.Envs[c.BodyName] = tmpFile.Name()

	// Prepare a shell command.
	cmd := exec.Command("sh", "-c",
		fmt.Sprintf(`
			set -euxo pipefail;
			%s
			`,
			c.Cmd))

	// Set the working directory to the temporary folder.
	cmd.Dir = tmpDir

	// Set all the envs in Envs.
	env := os.Environ()
	for k, v := range c.Envs {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	// Run the command and check its output.
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return CheckOutput{
			Success: false,
			Reason:  fmt.Sprintf("the command failed: `%s`", stderr.String()),
			Err:     fmt.Errorf("command execution failed: %v", err),
		}
	}

	return CheckOutput{
		Success: true,
	}
}

var _ Check[[]byte] = ShellCheck{}
