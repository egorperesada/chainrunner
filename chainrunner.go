package chainrunner

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	yaml "gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Chain interface {
	ConnectHost() error
	Commands() []ChainCommand
}

type ChainCommand interface {
	Execute()
}

func Run(chain Chain) error {
	err := chain.ConnectHost()
	if err != nil {
		return err
	}

	for _, command := range chain.Commands() {
		if command == nil {
			continue
		}
		command.Execute()
	}
	return nil
}

type TreeChain struct {
	name     string
	commands []ChainCommand
	config   *ssh.ClientConfig
	addr     string
	client   Client
}

func (t *TreeChain) ConnectHost() error {
	if t.addr == "" {
		return nil
	}
	t.enrichWithPublicKeys()
	log.Printf("connecting to %s:%s", t.name, t.addr)
	t.config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	for _, method := range t.config.Auth {
		log.Println(method)
	}
	client, err := ssh.Dial("tcp", t.addr, t.config)
	if err != nil {
		return err
	}
	log.Printf("connected to %s:%s", t.name, t.addr)
	session := &SshSession{
		client: client,
	}
	for _, command := range t.commands {
		simpleCommand, ok := command.(*SimpleCommand)
		if ok {
			simpleCommand.session = session
		}

		treeChain, ok := command.(*TreeChain)
		if ok {
			treeChain.client = &SshClient{client: client}
		}
	}

	return nil
}

func (t *TreeChain) Commands() []ChainCommand {
	return t.commands
}

func (t *TreeChain) Execute() {
	err := Run(t)
	if err != nil {
		log.Print(err)
	}
	t.client.Close()
}

type SimpleCommand struct {
	command string
	session Session
}

type Client interface {
	NewSession() (Session, error)
	Close() error
}

type ParentClient struct {
}

func (p *ParentClient) NewSession() (Session, error) {
	return &RootSession{os.Stdout}, nil
}

func (p *ParentClient) Close() error { return nil }

type SshClient struct {
	client *ssh.Client
}

func (s *SshClient) NewSession() (Session, error) {
	session, err := s.client.NewSession()
	if session != nil {
		session.Stdout = os.Stdout
		session.Stdin = os.Stdin
		session.Stderr = os.Stderr
	}
	return &SshSession{
		client:  s.client,
		session: session,
	}, err
}

func (s *SshClient) Close() error {
	return s.client.Close()
}

type Session interface {
	Run(cmd string) error
	Close() error
	SetStdout(writer io.Writer)
	GetStdout() io.Writer
}

type SshSession struct {
	client  *ssh.Client
	session *ssh.Session
}

func (s *SshSession) SetStdout(writer io.Writer) {
	s.session.Stdout = writer
}

func (s *SshSession) GetStdout() io.Writer {
	if s.session == nil {
		return os.Stdout
	}

	return s.session.Stdout
}

func (s *SshSession) Close() error {
	if s.session != nil {
		return s.session.Close()
	}

	return nil
}

func (s *SshSession) Run(cmd string) error {
	session, err := s.client.NewSession()
	if err != nil {
		return err
	}
	stdout := s.GetStdout()
	s.session = session
	defer s.Close()
	session.Stdin = os.Stdin
	session.Stdout = stdout
	session.Stderr = os.Stderr
	return session.Run(cmd)
}

type RootSession struct {
	stdout io.Writer
}

func (r *RootSession) SetStdout(writer io.Writer) {
	r.stdout = writer
}

func (r *RootSession) GetStdout() io.Writer {
	return r.stdout
}

func (r *RootSession) Close() error { return nil }

func (r *RootSession) Run(cmd string) error {
	command := exec.Command("/bin/sh", "-c", cmd)
	command.Stdout = r.stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	return command.Run()
}

func (s *SimpleCommand) Execute() {
	log.Print("$ " + s.command)
	if s.session == nil {
		s.session = &RootSession{os.Stdout}
	}
	s.session.Run(s.command)
}

func FromYaml(source string, isRaw bool) Chain {
	var data = source
	if !isRaw {
		rawData, _ := ioutil.ReadFile(source)
		data = string(rawData)
	}
	var out map[interface{}]interface{}
	yaml.Unmarshal([]byte(data), &out)
	return generateTreeChain("root", out["root"])
}

func NewTreeChain(name string, addr string, config *ssh.ClientConfig) *TreeChain {
	chain := &TreeChain{
		name:   name,
		addr:   addr,
		config: config,
		client: &ParentClient{},
	}
	return chain
}

func generateTreeChain(name string, node interface{}) *TreeChain {
	var element ChainCommand
	commands, ok := node.([]interface{})
	if !ok {
		return nil
	}
	addr, clientConfig := getOptionalTreeChainConfig(commands)
	chain := NewTreeChain(name, addr, clientConfig)
	for _, command := range commands {
		nestedChain, ok := command.(map[interface{}]interface{})
		if ok {
			var value interface{}
			var name interface{}
			for name, value = range nestedChain {
				break
			}
			if name == "addr" || name == "user" || name == "password" {
				continue
			}
			element = generateTreeChain(name.(string), value)
			if element != nil {
				chain.commands = append(chain.commands, element)
			}
			continue
		}
		_, ok = command.(string)
		if !ok {
			continue
		}
		element = &SimpleCommand{
			command: command.(string),
			session: nil,
		}
		chain.commands = append(chain.commands, element)
	}

	return chain
}

func getOptionalTreeChainConfig(options []interface{}) (string, *ssh.ClientConfig) {
	var addr string
	var clientConfig = &ssh.ClientConfig{}
	for _, option := range options {
		keyValue, ok := option.(map[interface{}]interface{})
		if !ok {
			continue
		}

		var key interface{}
		var value interface{}
		for key, value = range keyValue {
			break
		}

		if key == "addr" {
			addr = value.(string)
			continue
		}

		if key == "user" {
			clientConfig.User = value.(string)
			continue
		}

		if key == "password" {
			clientConfig.Auth = append(clientConfig.Auth, ssh.Password(value.(string)))
			continue
		}
	}

	return addr, clientConfig
}

func (t *TreeChain) enrichWithPublicKeys() {
	session, err := t.client.NewSession()
	if err != nil {
		return
	}

	sshDir := getSshDir(session)
	files, err := getFiles(session, sshDir)
	if err != nil {
		log.Println(err)
		return
	}

	for _, file := range files {
		if len(file) <= 4 {
			continue
		}

		extension := file[len(file)-4:]
		if extension != ".pub" {
			continue
		}

		path := sshDir + "/" + file[:len(file)-4]
		data, err := readFile(session, path)
		if err != nil {
			log.Println(err)
			continue
		}
		signer, err := ssh.ParsePrivateKey(data)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("add rsa-key from %s len %d", path, len(data))
		t.config.Auth = append(t.config.Auth, ssh.PublicKeys(signer))
	}
}

func getSshDir(session Session) string {
	defer session.Close()
	buf := bytes.NewBuffer([]byte(``))
	oldStdout := session.GetStdout()
	session.SetStdout(buf)
	err := session.Run("echo $USER")
	session.SetStdout(oldStdout)
	if err != nil {
		return ""
	}
	home, err := ioutil.ReadAll(buf)
	if err != nil {
		return ""
	}

	return "/home/" + strings.Replace(string(home), "\n", "", -1) + "/.ssh"
}

func getFiles(session Session, dir string) ([]string, error) {
	defer session.Close()
	buf := bytes.NewBuffer([]byte(``))
	oldStdout := session.GetStdout()
	session.SetStdout(buf)
	err := session.Run(fmt.Sprintf("ls %s", dir))
	session.SetStdout(oldStdout)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(buf)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(data), "\n"), nil
}

func readFile(session Session, path string) ([]byte, error) {
	defer session.Close()
	buf := bytes.NewBuffer([]byte(``))
	oldStdout := session.GetStdout()
	session.SetStdout(buf)
	err := session.Run(fmt.Sprintf("cat %s", path))
	session.SetStdout(oldStdout)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(buf)
	if err != nil {
		return nil, err
	}

	return data, nil
}
