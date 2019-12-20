package chainrunner

import (
	"gopkg.in/yaml.v2"
	"log"
)

type Chain map[interface{}]interface{}

type Provider interface {
	CreateChain() *CommandsChain
}

type YamlProvider struct {
	source string
	data   string
}

func (y *YamlProvider) CreateChain() *CommandsChain {
	chain, err := y.getData()
	if err != nil {
		log.Fatal(err)
	}
	return NewCommandsChain(chain, NewLocalHost())
}

func (y *YamlProvider) getData() (Chain, error) {
	var rawData []byte
	var err error
	if y.source != "" {
		session := NewLocalHost()
		rawData, err = readFile(session, y.source)
		if err != nil {
			return nil, err
		}
	}
	var out Chain
	err = yaml.Unmarshal(rawData, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func NewYamlProvider(path string, rawData string, isRaw bool) (*YamlProvider, error) {
	data, source := "", path
	if isRaw {
		data, source = rawData, ""
	}

	return &YamlProvider{data: data, source: source}, nil
}

func NewYamlProviderFromFile(path string) (*YamlProvider, error) {
	return NewYamlProvider(path, "", false)
}

func NewYamlProviderFromString(data string) (*YamlProvider, error) {
	return NewYamlProvider("", data, true)
}
