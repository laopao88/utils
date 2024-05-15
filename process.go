package zaia

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"time"
)

type Process struct {
	ExecPath   string
	Args       []string
	Pid        int
	Cmd        *exec.Cmd
	ExitState  *os.ProcessState
	StdinPipe  io.WriteCloser
	StdoutPipe io.ReadCloser
	StderrPipe io.ReadCloser
	Cancel     context.CancelFunc
	StopSignal chan bool
	ExitCode   int
}

var carriageReturnLineSplitter bufio.SplitFunc = func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	a, t, e := bufio.ScanLines(data, atEOF)
	if !atEOF && a == 0 {
		firstCRIndex := bytes.IndexByte(data, '\r')
		if firstCRIndex >= 0 && len(data) > 1 {
			nextCRIndex := bytes.IndexByte(data[firstCRIndex+1:], '\r')
			if nextCRIndex == -1 {
				return len(data), data[firstCRIndex+1:], nil
			}
			return nextCRIndex + 1, data[firstCRIndex+1 : nextCRIndex+1], nil
		}
	}

	return a, t, e
}

func StartProcess(timeout time.Duration, dir string, name string, outReaderHandler func(string), errReaderHandler func(string), args ...string) (*Process, error) {
	proc, err := CreateProcess(timeout, dir, name, outReaderHandler, errReaderHandler, args...)
	if err != nil {
		return nil, err
	}
	if err := proc.Cmd.Start(); err != nil {
		return nil, err
	}
	proc.Pid = proc.Cmd.Process.Pid

	return proc, nil
}

func CreateProcess(timeout time.Duration, dir string, name string, outReaderHandler func(string), errReaderHandler func(string), args ...string) (*Process, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	var cmd = exec.CommandContext(ctx, name, args...) // is not yet started.
	if dir == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		cmd.Dir = pwd
	} else {
		cmd.Dir = dir
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if outReaderHandler != nil {
		go func() {
			scanner := bufio.NewScanner(stdout)
			scanner.Split(carriageReturnLineSplitter)
			for scanner.Scan() {
				outReaderHandler(scanner.Text())
			}
		}()
	}

	if errReaderHandler != nil {
		go func() {
			scanner := bufio.NewScanner(stderr)
			scanner.Split(carriageReturnLineSplitter)
			for scanner.Scan() {
				errReaderHandler(scanner.Text())
			}
		}()
	}

	proc := &Process{
		ExecPath:   name,
		Args:       args,
		Cmd:        cmd,
		ExitState:  nil,
		StdinPipe:  stdin,
		StdoutPipe: stdout,
		StderrPipe: stderr,
		Cancel:     cancel,
		StopSignal: make(chan bool, 1),
		ExitCode:   -1,
	}
	return proc, nil
}

func (proc *Process) Close() {
	proc.StdinPipe.Close()
	proc.StderrPipe.Close()
	proc.StdoutPipe.Close()
}

func (proc *Process) Stop(kill bool) error {
	if kill {
		return proc.Cmd.Process.Kill()
	}
	return proc.Cmd.Process.Signal(os.Interrupt)
}

func (proc *Process) Wait() {
	defer func() {
		proc.Cancel()
	}()
	err := proc.Cmd.Wait()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			proc.ExitState = exitError.ProcessState
			proc.ExitCode = exitError.ExitCode()
		}
	}
	proc.ExitState = proc.Cmd.ProcessState
	proc.StopSignal <- true
}

func (proc *Process) ReadAll() (stdout []byte, stderr []byte, err error) {
	outbz, err := io.ReadAll(proc.StdoutPipe)
	if err != nil {
		return nil, nil, err
	}
	errbz, err := io.ReadAll(proc.StderrPipe)
	if err != nil {
		return nil, nil, err
	}
	return outbz, errbz, nil
}
