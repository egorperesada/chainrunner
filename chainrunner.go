package chainrunner

import "log"

type Command interface {
	Execute() error
}

// implements Command
type SingleCommand struct {
	cmd  string
	host Session
}

func NewSingleCommand(cmd string) *SingleCommand {
	return &SingleCommand{
		cmd: cmd,
	}
}

func (s *SingleCommand) Execute() error {
	log.Printf("%s: $ %s", s.host.String(), s.cmd)
	return s.host.Run(s.cmd)
}

// implements command
type CommandsChain struct {
	name     string
	session  Session
	commands []Command
}

func NewCommandsChain(content map[interface{}]interface{}, session Session) *CommandsChain {
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
		chain.session = &RemoteHost{}
	}
	for _, element := range data {
		singleCommand, ok := element.(string)
		if ok {
			chain.AddCommand(NewSingleCommand(singleCommand))
		}

		data, ok := element.(map[interface{}]interface{})
		if ok {
			command := NewCommandsChain(data, nil)
			if command != nil {
				chain.AddCommand(command)
				continue
			}

			chain.session.Configure(data)
		}
	}

	return chain
}

func (c *CommandsChain) Execute() error {
	log.Printf("run chain %s", c.name)
	defer log.Printf("finished chain %s", c.name)
	for _, command := range c.commands {
		singleCommand, ok := command.(*SingleCommand)
		if ok {
			singleCommand.host = c.session
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
