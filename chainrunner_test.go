package chainrunner

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRunnerExecute(t *testing.T) {
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
