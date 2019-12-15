package chainrunner

import (
	"golang.org/x/crypto/ssh"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
		log.Print(command)
		if command == nil {
			continue
		}
		command.Execute()
	}
	return nil
}

type TreeChain struct {
	commands []ChainCommand
	config   *ssh.ClientConfig
	addr     string
	session  Session
}

func (t *TreeChain) ConnectHost() error {
	if t.addr == "" {
		return nil
	}
	t.config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	client, err := ssh.Dial("tcp", t.addr, t.config)
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	for _, command := range t.commands {
		simpleCommand, ok := command.(*SimpleCommand)
		if ok {
			simpleCommand.session = session
		}
	}
	t.session = session
	return nil
}

func (t *TreeChain) Commands() []ChainCommand {
	return t.commands
}

func (t *TreeChain) Execute() {
	log.Print("start tree chain")
	err := Run(t)
	log.Print(err)
	t.session.Close()
}

type SimpleCommand struct {
	command string
	session Session
}

type Session interface {
	Run(cmd string) error
	Close() error
}

type RootSession struct {
}

func (r *RootSession) Close() error { return nil }

func (r *RootSession) Run(cmd string) error {
	command := exec.Command("/bin/sh", "-c", cmd)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	return command.Run()
}

func (s *SimpleCommand) Execute() {
	log.Print("simple command:" + s.command)
	if s.session == nil {
		s.session = &RootSession{}
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
	return generateTreeChain(out["root"])
}

func NewTreeChain(addr string, config *ssh.ClientConfig) *TreeChain {
	chain := &TreeChain{
		addr:   addr,
		config: config,
	}
	return chain
}

func generateTreeChain(node interface{}) *TreeChain {
	var element ChainCommand
	commands, ok := node.([]interface{})
	if !ok {
		return nil
	}
	var possibleOptions []interface{}
	if len(commands) >= 3 {
		possibleOptions = commands[:3]
	}
	chain := NewTreeChain(getOptionalTreeChainConfig(possibleOptions))
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
			element = generateTreeChain(value)
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
			clientConfig.Auth = append(clientConfig.Auth, ssh.PasswordCallback(func() (string, error) { return value.(string), nil }))
			continue
		}
	}

	return addr, clientConfig
}
