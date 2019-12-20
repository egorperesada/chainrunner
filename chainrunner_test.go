package chainrunner

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestRunnerExecute(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	type testCase struct {
		source    string
		checkFunc func(t *testing.T)
	}
	cases := []testCase{
		{"test/chains/simpleChain.yaml", func(t *testing.T) {
			return
		}},
		{"test/chains/simpleChainRevert.yaml", func(t *testing.T) {
			return
		}},
		{"test/chains/chainWithSshConnections.yaml", func(t *testing.T) {
			return
		}},
		{"test/chains/chainWithSshConnectionsRevert.yaml", func(t *testing.T) {
			return
		}},
	}

	for _, chain := range cases {
		t.Run(chain.source, func(t *testing.T) {
			provider, err := NewYamlProviderFromFile(chain.source)
			assert.Nil(t, err)
			runner := provider.CreateChain()
			assert.NotNil(t, runner)
			err = runner.Execute()
			assert.Nil(t, err)
			chain.checkFunc(t)
		})
	}
}
