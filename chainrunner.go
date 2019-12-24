package chainrunner

import (
	"errors"
	"log"
)

type Command interface {
	Execute() error
}

// implements Command
type SingleCommand struct {
	cmd     string
	session Session
}

func NewSingleCommand(cmd string) *SingleCommand {
	return &SingleCommand{
		cmd: cmd,
	}
}

func (s *SingleCommand) Execute() error {
	if s.session == nil {
		return errors.New("no session specified for command")
	}
	err := s.session.Run(s.cmd)
	log.Printf("%s: $ %s", s.session.String(), s.cmd)
	if err != nil {
		log.Printf("%s err: %s", s.cmd, err)
	}
	return err
}

// implements command
type CommandsChain struct {
	name     string
	session  Session
	commands []Command
}

func NewCommandsChain(content Chain, session Session) *CommandsChain {
	var name, block interface{}
	for name, block = range content {
		break
	}
	data, ok := block.([]interface{})
	if !ok {
		return nil
	}
	chain := &CommandsChain{name: name.(string)}
	if session != nil {
		chain.session = session
	} else {
		chain.session = NewEmptyRemoteHost()
	}
	data, config := extractConfigFields(data)
	chain.session.Configure(config)
	for _, element := range data {
		command := getCommand(element)
		if command != nil {
			chain.AddCommand(command)
			continue
		}
	}

	return chain
}

func extractConfigFields(data []interface{}) ([]interface{}, map[string]interface{}) {
	config := make(map[string]interface{})
	var toDelete []int
	for i, element := range data {
		keyValue, ok := element.(Chain)
		if !ok {
			continue
		}
		var key, value interface{}
		for key, value = range keyValue {
			break
		}

		if key == "addr" || key == "user" || key == "password" {
			toDelete = append(toDelete, i)
			config[key.(string)] = value
		}
	}

	return deleteFromSlice(data, toDelete), config
}

func getCommand(element interface{}) Command {
	singleCommand, ok := element.(string)
	if ok {
		return NewSingleCommand(singleCommand)
	}

	data, ok := element.(Chain)
	if !ok {
		return nil
	}

	chain := NewCommandsChain(data, nil)
	if chain == nil {
		return nil
	}

	return chain
}

func (c *CommandsChain) Execute() error {
	log.Printf("run chain %s", c.name)
	defer log.Printf("finished chain %s", c.name)
	for _, command := range c.commands {
		singleCommand, ok := command.(*SingleCommand)
		if ok {
			singleCommand.session = c.session
		}
		err := command.Execute()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CommandsChain) AddCommand(command Command) *CommandsChain {
	c.commands = append(c.commands, command)
	return c
}
