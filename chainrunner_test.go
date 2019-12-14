package chainrunner

import (
	"io/ioutil"
	"testing"
)

func TestSimpleChain(t *testing.T) {
	chain := FromYaml("simpleChain.yaml", false)
	err := Run(chain)
	if err != nil {
		t.Error(err)
	}

	data, err := ioutil.ReadFile("test")
	if err != nil {
		t.Error(err)
	}

	if string(data) != "test" {
		t.Errorf("expected test, got %s", string(data))
	}
}

func TestChainWithHostAttributes(t *testing.T) {
	chain := FromYaml("chainWithHostAttributes.yaml", false)
	tchain := chain.(*TreeChain)
	if tchain.addr != "127.0.0.1:22" {
		t.Errorf("expected 127.0.0.1:22, got %s", tchain.addr)
	}

	if tchain.config.User != "test" {
		t.Errorf("expected test, got %s", tchain.config.User)
	}

	if len(tchain.config.Auth) != 1 {
		t.Error("wrong treechain config")
	}
}
