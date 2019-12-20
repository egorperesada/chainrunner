package chainrunner

import (
	"io"
	"os"
	"os/exec"
)

type Session interface {
	Run(cmd string) error
	String() string
	SetStdout(writer io.Writer)
	GetStdout() io.Writer
	Configure(keyValue map[interface{}]interface{})
}

func NewLocalHost() *LocalHost {
	return &LocalHost{stdout: os.Stdout}
}

type LocalHost struct {
	stdout io.Writer
}

func (l *LocalHost) SetStdout(writer io.Writer) {
	l.stdout = writer
}

func (l *LocalHost) GetStdout() io.Writer {
	return l.stdout
}

func (l *LocalHost) Configure(keyValue map[interface{}]interface{}) {
	return
}

func (l *LocalHost) Run(cmd string) error {
	command := exec.Command("/bin/sh", "-c", cmd)
	command.Stdout = l.stdout
	return command.Run()
}

func (l *LocalHost) String() string {
	return "localhost"
}

type RemoteHost struct {
	stdout io.Writer
}

func (r *RemoteHost) Run(cmd string) error {
	panic("implement me")
}

func (r *RemoteHost) String() string {
	panic("implement me")
}

func (r *RemoteHost) SetStdout(writer io.Writer) {
	r.stdout = writer
}

func (r *RemoteHost) GetStdout() io.Writer {
	return r.stdout
}

func (r *RemoteHost) Configure(keyValue map[interface{}]interface{}) {
	var key interface{}
	//var value interface{}
	for key, _ = range keyValue {
		break
	}

	if key == "addr" {
		return
	}

	if key == "user" {
		return
	}

	if key == "password" {
		return
	}
}
