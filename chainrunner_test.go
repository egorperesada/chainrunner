package chainrunner

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"testing"
)

func TestChainRunner(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	cases := []*TestCase{
		{t,
			"SimpleChain",
			map[string]interface{}{"source": "test/chains/simpleChain.yaml"},
			pass,
			initFromFile,
			assertSimpleChain,
			pass,
		},
		{t,
			"SimpleChainRevert",
			map[string]interface{}{"source": "test/chains/simpleChainRevert.yaml"},
			pass,
			initFromFile,
			pass,
			pass,
		},
		{t,
			"chainOnRemoteHost",
			map[string]interface{}{"source": "test/chains/chainOnRemoteHost.yaml"},
			runEnvironment,
			initFromString,
			assertChainOnRemoteHost,
			stopEnvironment,
		},
		{t,
			"chainOnRemoteHostRevert",
			map[string]interface{}{"source": "test/chains/chainOnRemoteHostRevert.yaml"},
			pass,
			initFromFile,
			pass,
			pass,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, testCase.before(testCase))
			defer testCase.after(testCase)
			assert.True(t, testCase.init(testCase))
			assert.True(t, testCase.assert(testCase))
		})
	}
}

func assertChainOnRemoteHost(tc *TestCase) (result bool) {
	testEnv, ok := tc.artifacts["environment"].(*testEnvironment)
	assert.True(tc.t, ok)
	port, err := testEnv.GetSshPort()
	assert.NotEmpty(tc.t, port)
	assert.Nil(tc.t, err)
	config := &ssh.ClientConfig{
		Config:          ssh.Config{},
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.Password("password")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         0,
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("127.0.0.1:%s", port), config)
	if assert.Nil(tc.t, err) {
		defer client.Close()
	}
	session, err := client.NewSession()
	if assert.Nil(tc.t, err) {
		defer session.Close()
	}
	buf := bytes.NewBuffer([]byte(``))
	errors := bytes.NewBuffer([]byte(``))
	session.Stdout = buf
	session.Stderr = errors
	err = session.Run("cat chainrunner")
	assert.Nil(tc.t, err)
	assert.Equal(tc.t, "", errors.String())
	result = assert.Contains(tc.t, buf, "line was added on container")
	return
}

func assertSimpleChain(tc *TestCase) (result bool) {
	for _, value := range []string{"TestSimpleChain", "TestSimpleChain2"} {
		result = func(fileName string) bool {
			data, err := ioutil.ReadFile(fileName)
			assert.Nil(tc.t, err)
			actual := data[:len(data)-1] // delete "\n" in the end
			return assert.Equal(tc.t, []byte(fileName), actual)
		}(value)
		if result == false {
			return
		}
	}

	return
}

func initFromString(tc *TestCase) (result bool) {
	source := tc.artifacts["source"].(string)
	provider, err := NewYamlProviderFromString(source)
	assert.Nil(tc.t, err)
	runner := provider.CreateChain()
	assert.NotNil(tc.t, runner)
	err = runner.Execute()
	result = assert.Nil(tc.t, err)
	return
}

func initFromFile(tc *TestCase) (result bool) {
	source := tc.artifacts["source"].(string)
	provider, err := NewYamlProviderFromFile(source)
	assert.Nil(tc.t, err)
	runner := provider.CreateChain()
	assert.NotNil(tc.t, runner)
	err = runner.Execute()
	result = assert.Nil(tc.t, err)
	return
}

func runEnvironment(tc *TestCase) (result bool) {
	testEnv := NewTestEnvironment()
	err := testEnv.Start()
	assert.Nil(tc.t, err)
	tc.artifacts["environment"] = testEnv
	data, err := readFile(NewLocalHost(), tc.artifacts["source"].(string))
	assert.Nil(tc.t, err)
	port, err := testEnv.GetSshPort()
	result = assert.Nil(tc.t, err)
	tc.artifacts["source"] = strings.Replace(string(data), "{{port}}", port, -1)
	return
}

func stopEnvironment(tc *TestCase) (result bool) {
	testEnv, ok := tc.artifacts["environment"].(*testEnvironment)
	if !ok {
		tc.t.Error("can not get TestEnvironment link from artifacts")
	}
	err := testEnv.Close()
	result = assert.Nil(tc.t, err)
	return
}

type TestCaseStep func(tc *TestCase) (result bool)

type TestCase struct {
	t         *testing.T
	name      string
	artifacts map[string]interface{}
	before    TestCaseStep
	init      TestCaseStep
	assert    TestCaseStep
	after     TestCaseStep
}

func pass(tc *TestCase) bool { return true }

func NewTestEnvironment() *testEnvironment {
	return &testEnvironment{}
}

type testEnvironment struct {
	containerName string
}

func (t *testEnvironment) Start() error {
	err := exec.Command("/bin/sh", "-c", fmt.Sprintf("docker build -t chainrunner_ssh_server:latest -f test/sshServerImage/Dockerfile .")).Run()
	if err != nil {
		return err
	}
	data, err := getShellResult("docker run -d -P chainrunner_ssh_server", NewLocalHost())
	t.containerName = strings.Replace(string(data), "\n", "", -1)

	return err
}

func (t *testEnvironment) GetSshPort() (string, error) {
	data, err := getShellResult(fmt.Sprintf("docker port %s 22", t.containerName), NewLocalHost())
	if err != nil {
		return "", err
	}
	parts := strings.Split(string(data), ":")

	return strings.Replace(parts[1], "\n", "", -1), nil
}

func (t *testEnvironment) Close() error {
	exec.Command("/bin/sh", "-c", "docker stop "+t.containerName).Run()
	return exec.Command("/bin/sh", "-c", "docker rm "+t.containerName).Run()
}
