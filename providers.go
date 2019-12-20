package chainrunner

import (
	"gopkg.in/yaml.v2"
)

type Provider interface {
	CreateChain() *CommandsChain
}

type YamlProvider struct {
	data map[interface{}]interface{}
}

func (y *YamlProvider) CreateChain() *CommandsChain {
	return NewCommandsChain(y.data, NewLocalHost())
}

func NewYamlProvider(path string, data string, isRaw bool) (*YamlProvider, error) {
	if !isRaw {
		session := NewLocalHost()
		rawData, err := readFile(session, path)
		if err != nil {
			return nil, err
		}
		data = string(rawData)
	}
	var out map[interface{}]interface{}
	err := yaml.Unmarshal([]byte(data), &out)
	if err != nil {
		return nil, err
	}

	return &YamlProvider{data: out}, nil
}

func NewYamlProviderFromFile(path string) (*YamlProvider, error) {
	return NewYamlProvider(path, "", false)
}

func NewYamlProviderFromString(data string) (*YamlProvider, error) {
	return NewYamlProvider("", data, true)
}
