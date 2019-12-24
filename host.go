package chainrunner

import (
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

type Session interface {
	Run(cmd string) error
	String() string
	SetStdout(writer io.Writer)
	GetStdout() io.Writer
	Configure(data map[string]interface{})
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

func (l *LocalHost) Configure(data map[string]interface{}) {
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
	stdout       io.Writer
	client       *ssh.Client
	clientConfig *ssh.ClientConfig
	addr         string
}

func NewEmptyRemoteHost() *RemoteHost {
	return &RemoteHost{
		stdout: os.Stdout,
		clientConfig: &ssh.ClientConfig{
			Config:          ssh.Config{},
			User:            "",
			Auth:            []ssh.AuthMethod{},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         time.Second * 5,
		},
	}
}

func (r *RemoteHost) isConnected() bool {
	return r.client != nil
}

func (r *RemoteHost) connect() error {
	if r.isConnected() {
		return nil
	}
	log.Printf("connecting to remote session: %s@%s", r.clientConfig.User, r.addr)
	var err error
	r.client, err = ssh.Dial("tcp", r.addr, r.clientConfig)
	if err != nil {
		log.Printf("connection failed: %s", err)
	}
	return err
}

func (r *RemoteHost) Run(cmd string) error {
	err := r.connect()
	if err != nil {
		return err
	}

	session, err := r.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	session.Stdout = r.stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin
	err = session.Run(cmd)
	return err
}

func (r *RemoteHost) String() string {
	return r.addr
}

func (r *RemoteHost) SetStdout(writer io.Writer) {
	r.stdout = writer
}

func (r *RemoteHost) GetStdout() io.Writer {
	return r.stdout
}

func (r *RemoteHost) Configure(data map[string]interface{}) {
	for key, value := range data {
		if key == "addr" {
			r.addr = value.(string)
			continue
		}

		if key == "user" {
			r.clientConfig.User = value.(string)
			continue
		}

		if key == "password" {
			r.clientConfig.Auth = append(r.clientConfig.Auth, ssh.Password(value.(string)))
			continue
		}
	}
}
