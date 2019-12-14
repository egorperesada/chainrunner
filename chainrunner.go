package chainrunner

import (
	"golang.org/x/crypto/ssh"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
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
		command.Execute()
	}
	return nil
}

type TreeChain struct {
	commands []ChainCommand
	config   ssh.ClientConfig
	addr     string
}

func (t *TreeChain) ConnectHost() error {
	panic("implement me")
}

func (t *TreeChain) Commands() []ChainCommand {
	panic("implement me")
}

func (t *TreeChain) Execute() {
	panic("implement me")
}

type SimpleCommand struct {
	command string
	session *ssh.Session
}

func (s *SimpleCommand) Execute() {
	panic("implement me")
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

func NewTreeChain(addr string, config ssh.ClientConfig) *TreeChain {
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
	chain := NewTreeChain(getOptionalTreeChainConfig(commands[:3]))
	for _, command := range commands {
		nestedChain, ok := command.(map[interface{}]interface{})
		if ok {
			element = generateTreeChain(nestedChain)
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

func getOptionalTreeChainConfig(options []interface{}) (string, ssh.ClientConfig) {
	var addr string
	var clientConfig = ssh.ClientConfig{}
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
