package runner

import (
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func (e *Engine) killCmd(cmd *exec.Cmd) (pid int, err error) {
	pid = cmd.Process.Pid

	if e.config.Build.SendInterrupt {
		// Sending a signal to make it clear to the process that it is time to turn off
		if err = syscall.Kill(-pid, syscall.SIGINT); err != nil {
			return
		}
		time.Sleep(e.config.Build.KillDelay * time.Millisecond)
	}

	// https://stackoverflow.com/questions/22470193/why-wont-go-kill-a-child-process-correctly
	err = syscall.Kill(-pid, syscall.SIGKILL)

	// Wait releases any resources associated with the Process.
	_, _ = cmd.Process.Wait()
	return
}

func (e *Engine) startCmd(cmd string) (*exec.Cmd, io.WriteCloser, io.ReadCloser, io.ReadCloser, error) {
	c := exec.Command("/bin/sh", "-c", cmd)
	// assign the child process a new PGID identical to its PID
	// ref: https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stderr, err := c.StderrPipe()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	stdin, err := c.StdinPipe()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	err = c.Start()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return c, stdin, stdout, stderr, nil
}
