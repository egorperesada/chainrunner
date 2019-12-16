package chainrunner

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestSimpleChain(t *testing.T) {
	chain := FromYaml("test/chains/simpleChain.yaml", false)
	err := Run(chain)
	if err != nil {
		t.Error(err)
	}
	for _, fileName := range []string{"TestSimpleChain", "TestSimpleChain2"} {
		func(t *testing.T, fileName string) {
			defer os.Remove(fileName)
			data, err := ioutil.ReadFile(fileName)
			if err != nil {
				t.Error(err)
			}

			if !strings.Contains(string(data), fileName) {
				t.Errorf("expected %s, got '%s'", fileName, string(data))
			}
		}(t, fileName)
	}

}

func TestChainWithHostAttributes(t *testing.T) {
	chain := FromYaml("test/chains/chainWithHostAttributes.yaml", false)
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

func TestChainWithSshConnections(t *testing.T) {
	stopEnvironment, resultYaml, err, addr := startEnvironment()
	if err != nil {
		t.Error(err)
	}
	defer stopEnvironment()
	chain := FromYaml(resultYaml, true)
	defer os.Remove("chainrunner")
	err = Run(chain)
	if err != nil {
		t.Error(err)
	}
	config := &ssh.ClientConfig{
		Config:          ssh.Config{},
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.Password("password")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		t.Error(err)
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		t.Error(err)
	}
	defer session.Close()
	buf := bytes.NewBuffer([]byte(``))
	session.Stdout = buf
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin
	session.Run("cat chainrunner")
	content, err := ioutil.ReadAll(buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "line was added on host") {
		t.Error("chainrunner does not contain line from host")
	}
	if !strings.Contains(string(content), "line was added on container") {
		t.Error("chainrunner does not contain line from container")
	}
	if !strings.Contains(string(content), "2d command on container") {
		t.Error("chainrunner does not contain 2d command on container")
	}
}

func startEnvironment() (func(), string, error, string) {
	containerName := "test_sshd"
	exec.Command("/bin/sh", "-c", fmt.Sprintf("docker build -t ssh_server:latest -f test/sshServerImage/Dockerfile .")).Run()
	command := exec.Command("/bin/sh", "-c", fmt.Sprintf("docker run -d -P --name %s ssh_server", containerName))
	command.Stderr = os.Stderr
	command.Run()
	port := getSshPort(containerName)
	if len(port) == 0 {
		return nil, "", errors.New("did not get port for container"), ""
	}
	source, err := ioutil.ReadFile("test/chains/chainWithSshConnections.yaml")
	if err != nil {
		return nil, "", err, ""
	}
	resultYaml := strings.Replace(string(source), "{{port}}", port, -1)
	return func() {
		exec.Command("/bin/sh", "-c", "docker stop "+containerName).Run()
		exec.Command("/bin/sh", "-c", "docker rm "+containerName).Run()
	}, resultYaml, nil, "127.0.0.1:" + port
}

func getSshPort(containerName string) string {
	command := exec.Command("/bin/sh", "-c", fmt.Sprintf("docker port %s 22", containerName))
	buf := bytes.NewBuffer([]byte(``))
	command.Stdout = buf
	command.Stderr = os.Stderr
	command.Run()
	data, _ := ioutil.ReadAll(buf)
	parts := strings.Split(string(data), ":")
	rawPort := parts[1]
	return strings.Replace(rawPort, "\n", "", -1)
}
